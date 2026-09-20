package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/jonnarhei/meal-planner/backend/internal/recipeclient"
	"github.com/jonnarhei/meal-planner/backend/internal/store"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

// a 2 day plan: day 1 is a spoonacular recipe, day 2 is one of the user's own
func planWithTwoDays() *models.MealPlan {
	userRecipeID := int64(77)
	return &models.MealPlan{
		ID:     10,
		UserID: 7,
		Recipes: []models.MealPlanRecipe{
			{ID: 101, MealPlanID: 10, RecipeID: 555, RecipeTitle: "Spoon Curry", Day: 1,
				Source: models.RecipeSourceSpoonacular},
			{ID: 102, MealPlanID: 10, RecipeTitle: "My Lasagna", Day: 2,
				Source: models.RecipeSourceUser, UserRecipeID: &userRecipeID},
		},
	}
}

func TestChangeRecipeForDayUserRecipe(t *testing.T) {
	t.Run("puts the user's recipe in the slot", func(t *testing.T) {
		var saved *models.MealPlanRecipe
		var gotRecipeID, gotUserID int64

		app := newTestApp(t, store.Storage{
			Mealplans: &mockMealPlanStore{
				getCurrentFn: func(ctx context.Context, userID int64) (*models.MealPlan, error) {
					return planWithTwoDays(), nil
				},
				updateRecipeForDayFn: func(ctx context.Context, r *models.MealPlanRecipe) error {
					saved = r
					return nil
				},
			},
			Recipes: &mockRecipeStore{
				getByIDFn: func(ctx context.Context, recipeID, userID int64) (*models.UserRecipe, error) {
					gotRecipeID, gotUserID = recipeID, userID
					return &models.UserRecipe{
						ID: 88, UserID: userID, Title: "Pancakes",
						Image: "img.png", SourceUrl: "https://example.com/pancakes",
					}, nil
				},
			},
		})
		// left nil on purpose: this path must not call spoonacular
		app.recipes = nil

		body := map[string]any{"day": 1, "user_recipe_id": 88}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusOK, rr.Code)

		if gotRecipeID != 88 || gotUserID != 7 {
			t.Errorf("expected recipe lookup with id 88 / user 7, got %d / %d", gotRecipeID, gotUserID)
		}

		if saved == nil {
			t.Fatal("expected UpdateRecipeForDay to be called")
		}
		if saved.Source != models.RecipeSourceUser {
			t.Errorf("expected source %q, got %q", models.RecipeSourceUser, saved.Source)
		}
		if saved.UserRecipeID == nil || *saved.UserRecipeID != 88 {
			t.Errorf("expected user_recipe_id 88, got %v", saved.UserRecipeID)
		}
		if saved.RecipeID != 0 {
			t.Errorf("expected spoonacular recipe id to be cleared, got %d", saved.RecipeID)
		}
		if saved.RecipeTitle != "Pancakes" || saved.Image != "img.png" || saved.SourceURL != "https://example.com/pancakes" {
			t.Errorf("recipe fields not copied onto the slot: %+v", saved)
		}
		if saved.MealPlanID != 10 || saved.Day != 1 || saved.ID != 101 {
			t.Errorf("expected slot id 101 / plan 10 / day 1, got %d / %d / %d", saved.ID, saved.MealPlanID, saved.Day)
		}

		var resp models.MealPlanRecipe
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if resp.Source != models.RecipeSourceUser {
			t.Errorf("expected source in response body, got %+v", resp)
		}
	})

	t.Run("recipe not owned is 404 and does not touch the plan", func(t *testing.T) {
		app := newTestApp(t, store.Storage{
			Mealplans: &mockMealPlanStore{
				getCurrentFn: func(ctx context.Context, userID int64) (*models.MealPlan, error) {
					return planWithTwoDays(), nil
				},
				// nil updateRecipeForDayFn: panics (500) if the handler wrongly saves
			},
			Recipes: &mockRecipeStore{
				getByIDFn: func(ctx context.Context, recipeID, userID int64) (*models.UserRecipe, error) {
					return nil, sql.ErrNoRows
				},
			},
		})

		body := map[string]any{"day": 1, "user_recipe_id": 88}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusNotFound, rr.Code)
	})

	t.Run("recipe lookup error is 500", func(t *testing.T) {
		app := newTestApp(t, store.Storage{
			Mealplans: &mockMealPlanStore{
				getCurrentFn: func(ctx context.Context, userID int64) (*models.MealPlan, error) {
					return planWithTwoDays(), nil
				},
			},
			Recipes: &mockRecipeStore{
				getByIDFn: func(ctx context.Context, recipeID, userID int64) (*models.UserRecipe, error) {
					return nil, errors.New("boom")
				},
			},
		})

		body := map[string]any{"day": 1, "user_recipe_id": 88}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestChangeRecipeForDayRandom(t *testing.T) {
	t.Run("swapping a user day back to spoonacular clears the link", func(t *testing.T) {
		var saved *models.MealPlanRecipe

		app := newTestApp(t, store.Storage{
			Mealplans: &mockMealPlanStore{
				getCurrentFn: func(ctx context.Context, userID int64) (*models.MealPlan, error) {
					return planWithTwoDays(), nil
				},
				updateRecipeForDayFn: func(ctx context.Context, r *models.MealPlanRecipe) error {
					saved = r
					return nil
				},
			},
			Users: &mockUserStore{
				getByIDFn: func(ctx context.Context, userID int64) (*models.User, error) {
					return &models.User{ID: userID}, nil
				},
			},
		})
		app.recipes = &mockRecipeClient{
			getRandomRecipesFn: func(ctx context.Context, n int, p, i, e []string) ([]recipeclient.Recipe, error) {
				return []recipeclient.Recipe{{RecipeID: 999, Title: "Random Stew", Image: "stew.png", URL: "https://example.com/stew"}}, nil
			},
		}

		// day 2 currently holds a user recipe
		body := map[string]any{"day": 2}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusOK, rr.Code)

		if saved == nil {
			t.Fatal("expected UpdateRecipeForDay to be called")
		}
		if saved.Source != models.RecipeSourceSpoonacular {
			t.Errorf("expected source %q, got %q", models.RecipeSourceSpoonacular, saved.Source)
		}
		if saved.UserRecipeID != nil {
			t.Errorf("expected user_recipe_id to be cleared, got %v", *saved.UserRecipeID)
		}
		if saved.RecipeID != 999 || saved.RecipeTitle != "Random Stew" {
			t.Errorf("spoonacular recipe not copied onto the slot: %+v", saved)
		}
	})

	t.Run("no suitable recipe is 500", func(t *testing.T) {
		app := newTestApp(t, store.Storage{
			Mealplans: &mockMealPlanStore{
				getCurrentFn: func(ctx context.Context, userID int64) (*models.MealPlan, error) {
					return planWithTwoDays(), nil
				},
			},
			Users: &mockUserStore{
				getByIDFn: func(ctx context.Context, userID int64) (*models.User, error) {
					return &models.User{ID: userID, ExcludedIngredients: []string{"beef"}}, nil
				},
			},
		})
		app.recipes = &mockRecipeClient{
			getRandomRecipesFn: func(ctx context.Context, n int, p, i, e []string) ([]recipeclient.Recipe, error) {
				return []recipeclient.Recipe{{
					RecipeID:    1,
					Title:       "Beef Stew",
					Ingredients: []recipeclient.Ingredient{{Name: "Beef chuck"}},
				}}, nil
			},
		}

		body := map[string]any{"day": 1}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestChangeRecipeForDayValidation(t *testing.T) {
	newApp := func(t *testing.T, plan *models.MealPlan, planErr error) *application {
		return newTestApp(t, store.Storage{
			Mealplans: &mockMealPlanStore{
				getCurrentFn: func(ctx context.Context, userID int64) (*models.MealPlan, error) {
					return plan, planErr
				},
			},
			Recipes: &mockRecipeStore{},
		})
	}

	t.Run("no current meal plan is 404", func(t *testing.T) {
		app := newApp(t, nil, sql.ErrNoRows)

		body := map[string]any{"day": 1, "user_recipe_id": 88}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusNotFound, rr.Code)
	})

	t.Run("plan lookup error is 500", func(t *testing.T) {
		app := newApp(t, nil, errors.New("boom"))

		body := map[string]any{"day": 1}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("day above the plan length is 400", func(t *testing.T) {
		app := newApp(t, planWithTwoDays(), nil)

		body := map[string]any{"day": 3, "user_recipe_id": 88}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("day zero is 400", func(t *testing.T) {
		app := newApp(t, planWithTwoDays(), nil)

		body := map[string]any{"day": 0, "user_recipe_id": 88}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing auth header", func(t *testing.T) {
		app := newApp(t, planWithTwoDays(), nil)

		body := map[string]any{"day": 1}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, ""), app)
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("update error is 500", func(t *testing.T) {
		app := newTestApp(t, store.Storage{
			Mealplans: &mockMealPlanStore{
				getCurrentFn: func(ctx context.Context, userID int64) (*models.MealPlan, error) {
					return planWithTwoDays(), nil
				},
				updateRecipeForDayFn: func(ctx context.Context, r *models.MealPlanRecipe) error {
					return errors.New("boom")
				},
			},
			Recipes: &mockRecipeStore{
				getByIDFn: func(ctx context.Context, recipeID, userID int64) (*models.UserRecipe, error) {
					return &models.UserRecipe{ID: recipeID, Title: "Pancakes"}, nil
				},
			},
		})

		body := map[string]any{"day": 1, "user_recipe_id": 88}
		rr := executeRequest(newJSONRequest(t, http.MethodPatch, "/meal-plans/current/recipe", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusInternalServerError, rr.Code)
	})
}
