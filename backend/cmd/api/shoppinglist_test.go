package main

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/jonnarhei/meal-planner/backend/internal/recipeclient"
	"github.com/jonnarhei/meal-planner/backend/internal/store"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

func userRecipeSlot(day int, userRecipeID *int64) models.MealPlanRecipe {
	return models.MealPlanRecipe{
		ID: int64(100 + day), MealPlanID: 10, Day: day,
		RecipeTitle: "My Recipe", Source: models.RecipeSourceUser, UserRecipeID: userRecipeID,
	}
}

func spoonSlot(day int, recipeID int64) models.MealPlanRecipe {
	return models.MealPlanRecipe{
		ID: int64(100 + day), MealPlanID: 10, Day: day, RecipeID: recipeID,
		RecipeTitle: "Spoon Recipe", Source: models.RecipeSourceSpoonacular,
	}
}

func findItem(items []models.ShoppinglistItem, name string) (models.ShoppinglistItem, bool) {
	for _, item := range items {
		if item.Name == name {
			return item, true
		}
	}
	return models.ShoppinglistItem{}, false
}

func assertAmount(t *testing.T, item models.ShoppinglistItem, want float64, wantUnit string) {
	t.Helper()
	if math.Abs(item.Amount-want) > 0.01 {
		t.Errorf("%s: expected amount %.2f, got %.2f", item.Name, want, item.Amount)
	}
	if item.Unit != wantUnit {
		t.Errorf("%s: expected unit %q, got %q", item.Name, wantUnit, item.Unit)
	}
}

// runs generateShoppingListFromPlan and returns whatever it handed to AddItems
func runGenerate(t *testing.T, plan *models.MealPlan,
	spoon func(context.Context, []int64) ([]recipeclient.RecipeWithIngredients, error),
	userIngredients map[int64][]models.UserRecipeIngredient,
) []models.ShoppinglistItem {
	t.Helper()

	var saved []models.ShoppinglistItem
	app := newTestApp(t, store.Storage{
		Shoppinglist: &mockShoppinglistStore{
			addItemsFn: func(ctx context.Context, userID int64, items []models.ShoppinglistItem) error {
				saved = items
				return nil
			},
		},
		Recipes: &mockRecipeStore{
			getIngredientsByRecipeIDsFn: func(ctx context.Context, userID int64, ids []int64) (map[int64][]models.UserRecipeIngredient, error) {
				return userIngredients, nil
			},
		},
	})
	// nil getRecipeInformationBulkFn panics if the spoonacular path runs when it should not
	app.recipes = &mockRecipeClient{getRecipeInformationBulkFn: spoon}

	if err := app.generateShoppingListFromPlan(context.Background(), 7, plan); err != nil {
		t.Fatalf("generateShoppingListFromPlan: %v", err)
	}
	return saved
}

func TestGenerateShoppingListFromPlanSpoonacular(t *testing.T) {
	plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{spoonSlot(1, 555)}}

	var gotIDs []int64
	items := runGenerate(t, plan,
		func(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
			gotIDs = ids
			return []recipeclient.RecipeWithIngredients{{
				ID: 555,
				Ingredients: []recipeclient.Ingredient{
					{Name: "flour", Amount: 100, Unit: "ml"},
					{Name: "Shopping list note", Amount: 1, Unit: "g"}, // filtered by isValidIngredient
					{Name: "rice", Amount: 2, Unit: "servings"},        // filtered by isValidUnit
				},
			}}, nil
		}, nil)

	if len(gotIDs) != 1 || gotIDs[0] != 555 {
		t.Errorf("expected spoonacular lookup for id 555, got %v", gotIDs)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item after filtering, got %d: %+v", len(items), items)
	}
	flour, ok := findItem(items, "flour")
	if !ok {
		t.Fatalf("expected a flour item, got %+v", items)
	}
	assertAmount(t, flour, 100, "ml")
	if flour.Source != "meal_plan" {
		t.Errorf("expected source meal_plan (what DeleteBySource removes), got %q", flour.Source)
	}
	if flour.UserID != 7 {
		t.Errorf("expected user id 7, got %d", flour.UserID)
	}
}

func TestGenerateShoppingListFromPlanUserRecipes(t *testing.T) {
	t.Run("user ingredients skip the spoonacular junk filters", func(t *testing.T) {
		id := int64(77)
		plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{userRecipeSlot(1, &id)}}

		items := runGenerate(t, plan, nil, map[int64][]models.UserRecipeIngredient{
			77: {
				{Name: "chicken (boneless)", Amount: 500, Unit: "g"},
				{Name: "salt", Amount: 2, Unit: "tsp"},
			},
		})

		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d: %+v", len(items), items)
		}
		if _, ok := findItem(items, "chicken (boneless)"); !ok {
			t.Errorf("expected parenthesised user ingredient to survive, got %+v", items)
		}
	})

	t.Run("the same recipe on two days counts twice", func(t *testing.T) {
		id := int64(77)
		plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{
			userRecipeSlot(1, &id),
			userRecipeSlot(2, &id),
		}}

		items := runGenerate(t, plan, nil, map[int64][]models.UserRecipeIngredient{
			77: {{Name: "salt", Amount: 2, Unit: "tsp"}},
		})

		salt, ok := findItem(items, "salt")
		if !ok {
			t.Fatalf("expected a salt item, got %+v", items)
		}
		assertAmount(t, salt, 4, "tsp")
	})

	t.Run("deleted recipe slot is skipped", func(t *testing.T) {
		id := int64(77)
		plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{
			userRecipeSlot(1, nil), // recipe was deleted, link set to NULL
			userRecipeSlot(2, &id),
		}}

		items := runGenerate(t, plan, nil, map[int64][]models.UserRecipeIngredient{
			77: {{Name: "salt", Amount: 2, Unit: "tsp"}},
		})

		if len(items) != 1 {
			t.Fatalf("expected only the surviving recipe's ingredient, got %+v", items)
		}
	})

	t.Run("recipe with no ingredients contributes nothing", func(t *testing.T) {
		id := int64(77)
		plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{userRecipeSlot(1, &id)}}

		items := runGenerate(t, plan, nil, map[int64][]models.UserRecipeIngredient{})

		if len(items) != 0 {
			t.Errorf("expected no items, got %+v", items)
		}
	})
}

func TestGenerateShoppingListFromPlanMixed(t *testing.T) {
	id := int64(77)
	plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{
		spoonSlot(1, 555),
		userRecipeSlot(2, &id),
	}}

	items := runGenerate(t, plan,
		func(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
			return []recipeclient.RecipeWithIngredients{{
				ID:          555,
				Ingredients: []recipeclient.Ingredient{{Name: "flour", Amount: 100, Unit: "ml"}},
			}}, nil
		},
		map[int64][]models.UserRecipeIngredient{
			77: {{Name: "flour", Amount: 2, Unit: "cup"}},
		})

	if len(items) != 1 {
		t.Fatalf("expected the two flours to merge into 1 item, got %d: %+v", len(items), items)
	}
	// 2 cups = 473.18 ml, plus the spoonacular 100 ml
	assertAmount(t, items[0], 573.18, "ml")
}

func TestGenerateShoppingListFromPlanErrors(t *testing.T) {
	t.Run("spoonacular is not called when the plan has no spoonacular recipes", func(t *testing.T) {
		id := int64(77)
		plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{userRecipeSlot(1, &id)}}

		// nil spoon function: the test panics if the handler calls it
		runGenerate(t, plan, nil, map[int64][]models.UserRecipeIngredient{
			77: {{Name: "salt", Amount: 1, Unit: "tsp"}},
		})
	})

	t.Run("spoonacular error is returned", func(t *testing.T) {
		plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{spoonSlot(1, 555)}}

		app := newTestApp(t, store.Storage{
			Shoppinglist: &mockShoppinglistStore{},
			Recipes:      &mockRecipeStore{},
		})
		app.recipes = &mockRecipeClient{
			getRecipeInformationBulkFn: func(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
				return nil, errors.New("boom")
			},
		}

		if err := app.generateShoppingListFromPlan(context.Background(), 7, plan); err == nil {
			t.Error("expected an error to be returned")
		}
	})

	t.Run("user ingredient lookup error is returned", func(t *testing.T) {
		id := int64(77)
		plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{userRecipeSlot(1, &id)}}

		app := newTestApp(t, store.Storage{
			Shoppinglist: &mockShoppinglistStore{},
			Recipes: &mockRecipeStore{
				getIngredientsByRecipeIDsFn: func(ctx context.Context, userID int64, ids []int64) (map[int64][]models.UserRecipeIngredient, error) {
					return nil, errors.New("boom")
				},
			},
		})

		if err := app.generateShoppingListFromPlan(context.Background(), 7, plan); err == nil {
			t.Error("expected an error to be returned")
		}
	})

	t.Run("ingredient lookup is scoped to the plan's user", func(t *testing.T) {
		id := int64(77)
		plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{userRecipeSlot(1, &id)}}

		var gotUserID int64
		var gotIDs []int64
		app := newTestApp(t, store.Storage{
			Shoppinglist: &mockShoppinglistStore{
				addItemsFn: func(ctx context.Context, userID int64, items []models.ShoppinglistItem) error { return nil },
			},
			Recipes: &mockRecipeStore{
				getIngredientsByRecipeIDsFn: func(ctx context.Context, userID int64, ids []int64) (map[int64][]models.UserRecipeIngredient, error) {
					gotUserID, gotIDs = userID, ids
					return nil, nil
				},
			},
		})

		if err := app.generateShoppingListFromPlan(context.Background(), 7, plan); err != nil {
			t.Fatal(err)
		}
		if gotUserID != 7 {
			t.Errorf("expected lookup scoped to user 7, got %d", gotUserID)
		}
		if len(gotIDs) != 1 || gotIDs[0] != 77 {
			t.Errorf("expected lookup for recipe 77, got %v", gotIDs)
		}
	})
}
