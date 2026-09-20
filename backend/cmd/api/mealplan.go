package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jonnarhei/meal-planner/backend/internal/jsonutil"
	"github.com/jonnarhei/meal-planner/backend/internal/recipeclient"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

func containsExcluded(recipe recipeclient.Recipe, excluded []string) bool {
	for _, ingredient := range recipe.Ingredients {
		for _, excl := range excluded {
			if strings.Contains(strings.ToLower(ingredient.Name), strings.ToLower(excl)) {
				return true
			}
		}
	}
	return false
}

func (app *application) generateMealPlan(ctx context.Context, userID int64, preferences []string, intolerances []string, excludedIngredients []string) (*models.MealPlan, error) {
	randomRecipes, err := app.recipes.GetRandomRecipes(ctx, 14, preferences, intolerances, excludedIngredients)

	if err != nil {
		return nil, err
	}

	seen := make(map[int64]bool)
	var mealPlanRecipes []models.MealPlanRecipe
	day := 1

	for _, recipe := range randomRecipes {
		if day > 7 {
			break
		}
		if seen[recipe.RecipeID] {
			continue
		}
		if containsExcluded(recipe, excludedIngredients) {
			continue
		}
		seen[recipe.RecipeID] = true

		mealPlanRecipes = append(mealPlanRecipes, models.MealPlanRecipe{
			RecipeID:    recipe.RecipeID,
			RecipeTitle: recipe.Title,
			Image:       recipe.Image,
			SourceURL:   recipe.URL,
			Day:         day,
			Source:      models.RecipeSourceSpoonacular,
		})
		day++
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)

	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	sunday := monday.AddDate(0, 0, 6)

	mealPlan := &models.MealPlan{
		UserID:    userID,
		StartDate: monday,
		EndDate:   sunday,
		Recipes:   mealPlanRecipes,
	}

	return mealPlan, nil
}

func (app *application) getCurrentMealPlanHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	plan, err := app.store.Mealplans.GetCurrent(r.Context(), claims.UserID)
	//check if there were errors, and ignore no rows error
	if err != nil && err != sql.ErrNoRows {
		slog.Error("failed to get current meal plan", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	//check if there was no plan returned
	if plan == nil {
		user, err := app.store.Users.GetByID(r.Context(), claims.UserID)
		if err != nil {
			slog.Error("failed to get user from db", "error", err)
			jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		plan, err = app.generateMealPlan(r.Context(), claims.UserID, user.DietaryPreferences, user.Intolerances, user.ExcludedIngredients)
		if err != nil {
			slog.Error("failed to generate meal plan", "error", err)
			jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if err = app.store.Mealplans.Create(r.Context(), plan); err != nil {
			slog.Error("failed to save meal plan in database", "error", err)
			jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if err = app.generateShoppingListFromPlan(r.Context(), claims.UserID, plan); err != nil {
			slog.Error("failed to generate shopping list from plan", "error", err)
		}
	}
	jsonutil.WriteHttpJson(w, http.StatusOK, plan)
}

var errNoSuitableRecipe = errors.New("could not find a suitable recipe")

type changeRecipePayload struct {
	Day          int64  `json:"day"`
	UserRecipeID *int64 `json:"user_recipe_id"`
}

func (app *application) changeRecipeForDay(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	var payload changeRecipePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonutil.WriteError(w, "bad request", http.StatusBadRequest)
		return
	}

	currentPlan, err := app.store.Mealplans.GetCurrent(r.Context(), claims.UserID)
	if err == sql.ErrNoRows {
		jsonutil.WriteError(w, "no active meal plan found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("failed to get current meal plan", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if payload.Day < 1 || int(payload.Day) > len(currentPlan.Recipes) {
		jsonutil.WriteError(w, "invalid day", http.StatusBadRequest)
		return
	}

	updatedRecipe := &models.MealPlanRecipe{
		ID:         currentPlan.Recipes[payload.Day-1].ID,
		MealPlanID: currentPlan.ID,
		Day:        int(payload.Day),
	}

	if payload.UserRecipeID != nil {
		userRecipe, err := app.store.Recipes.GetByID(r.Context(), *payload.UserRecipeID, claims.UserID)
		if err == sql.ErrNoRows {
			jsonutil.WriteError(w, "recipe not found", http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("failed to get user recipe from db", "error", err)
			jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		updatedRecipe.RecipeID = 0
		updatedRecipe.RecipeTitle = userRecipe.Title
		updatedRecipe.Image = userRecipe.Image
		updatedRecipe.SourceURL = userRecipe.SourceUrl
		updatedRecipe.Source = models.RecipeSourceUser
		updatedRecipe.UserRecipeID = &userRecipe.ID
	} else {
		recipe, err := app.pickRandomRecipe(r.Context(), claims.UserID)
		if errors.Is(err, errNoSuitableRecipe) {
			jsonutil.WriteError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err != nil {
			slog.Error("could not get random recipe", "error", err)
			jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		updatedRecipe.RecipeID = recipe.RecipeID
		updatedRecipe.RecipeTitle = recipe.Title
		updatedRecipe.Image = recipe.Image
		updatedRecipe.SourceURL = recipe.URL
		updatedRecipe.Source = models.RecipeSourceSpoonacular
	}

	if err := app.store.Mealplans.UpdateRecipeForDay(r.Context(), updatedRecipe); err != nil {
		slog.Error("failed to update recipe in the database", "error", err)
		jsonutil.WriteError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, updatedRecipe)
}

func (app *application) pickRandomRecipe(ctx context.Context, userID int64) (*recipeclient.Recipe, error) {
	user, err := app.store.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	recipes, err := app.recipes.GetRandomRecipes(ctx, 5, user.DietaryPreferences, user.Intolerances, user.ExcludedIngredients)
	if err != nil {
		return nil, err
	}

	for i := range recipes {
		if !containsExcluded(recipes[i], user.ExcludedIngredients) {
			return &recipes[i], nil
		}
	}

	return nil, errNoSuitableRecipe
}

func (app *application) regenerateMealPlanHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	if err := app.store.Shoppinglist.DeleteBySource(r.Context(), claims.UserID, "meal_plan"); err != nil {
		slog.Error("failed to delete old shopping list items", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := app.store.Mealplans.DeleteCurrent(r.Context(), claims.UserID); err != nil {
		slog.Error("could not delete the current mealplan", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := app.store.Users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("failed to get user from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	plan, err := app.generateMealPlan(r.Context(), claims.UserID, user.DietaryPreferences, user.Intolerances, user.ExcludedIngredients)
	if err != nil {
		slog.Error("could not generate a new mealplan from api", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err = app.store.Mealplans.Create(r.Context(), plan); err != nil {
		slog.Error("could not create new mealplan database entry", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err = app.generateShoppingListFromPlan(r.Context(), claims.UserID, plan); err != nil {
		slog.Error("failed to generate shopping list from plan", "error", err)
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, plan)
}
