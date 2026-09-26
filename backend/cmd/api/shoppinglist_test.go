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
				Ingredients: []recipeclient.Ingredient{{ID: flourID, Name: "flour", Amount: 100, Unit: "ml"}},
			}}, nil
		},
		map[int64][]models.UserRecipeIngredient{
			77: {{Name: "flour", Amount: 2, Unit: "cup"}},
		})

	if len(items) != 2 {
		t.Fatalf("expected an item from each source, got %d: %+v", len(items), items)
	}
	for _, item := range items {
		if item.IngredientID == nil {
			// 2 cups = 473.18 ml
			assertAmount(t, item, 473.18, "ml")
			continue
		}
		assertAmount(t, item, 100, "ml")
	}
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

// spoonacular ingredient ids: garlic is always 11215, whatever a recipe calls it
const (
	garlicID = 11215
	flourID  = 20081
	onionID  = 11282
)

func countItems(items []models.ShoppinglistItem, name string) int {
	n := 0
	for _, item := range items {
		if item.Name == name {
			n++
		}
	}
	return n
}

func assertIngredientID(t *testing.T, item models.ShoppinglistItem, want int64) {
	t.Helper()
	if item.IngredientID == nil {
		t.Fatalf("%s: expected ingredient id %d, got nil", item.Name, want)
	}
	if *item.IngredientID != want {
		t.Errorf("%s: expected ingredient id %d, got %d", item.Name, want, *item.IngredientID)
	}
}

func TestGenerateShoppingListAggregatesBySpoonacularID(t *testing.T) {
	// the issue: two recipes naming the same ingredient differently
	plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{
		spoonSlot(1, 555),
		spoonSlot(2, 556),
	}}

	items := runGenerate(t, plan,
		func(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
			return []recipeclient.RecipeWithIngredients{
				{ID: 555, Ingredients: []recipeclient.Ingredient{
					{ID: garlicID, Name: "garlic", Amount: 2, Unit: "clove"},
				}},
				{ID: 556, Ingredients: []recipeclient.Ingredient{
					{ID: garlicID, Name: "garlic cloves", Amount: 3, Unit: "cloves"},
				}},
			}, nil
		}, nil)

	if len(items) != 1 {
		t.Fatalf("expected the two garlics to merge on ingredient id, got %d: %+v", len(items), items)
	}
	assertAmount(t, items[0], 5, "clove")
	assertIngredientID(t, items[0], garlicID)
	if items[0].Name != "garlic" {
		t.Errorf("expected the first name seen to win, got %q", items[0].Name)
	}
}

func TestGenerateShoppingListKeepsDifferentIngredientIDsApart(t *testing.T) {
	// spoonacular has several ingredients called "salt"; they are not the same thing
	plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{spoonSlot(1, 555)}}

	items := runGenerate(t, plan,
		func(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
			return []recipeclient.RecipeWithIngredients{{ID: 555, Ingredients: []recipeclient.Ingredient{
				{ID: 1082047, Name: "salt", Amount: 5, Unit: "g"},
				{ID: 1102047, Name: "salt", Amount: 10, Unit: "g"},
			}}}, nil
		}, nil)

	if countItems(items, "salt") != 2 {
		t.Fatalf("expected two distinct salts, got %+v", items)
	}
}

func TestGenerateShoppingListFallsBackToNameWithoutID(t *testing.T) {
	// spoonacular occasionally returns no id; those still have to merge on name
	plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{spoonSlot(1, 555)}}

	items := runGenerate(t, plan,
		func(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
			return []recipeclient.RecipeWithIngredients{{ID: 555, Ingredients: []recipeclient.Ingredient{
				{Name: "flour", Amount: 100, Unit: "g"},
				{Name: "flour", Amount: 50, Unit: "g"},
			}}}, nil
		}, nil)

	if len(items) != 1 {
		t.Fatalf("expected the id-less flours to merge on name, got %d: %+v", len(items), items)
	}
	assertAmount(t, items[0], 150, "g")
	if items[0].IngredientID != nil {
		t.Errorf("expected no ingredient id, got %d", *items[0].IngredientID)
	}
}

func TestGenerateShoppingListDoesNotMergeIdentifiedWithUserIngredients(t *testing.T) {
	// user recipes carry no spoonacular id, so they cannot match an identified
	// ingredient even when the name and unit line up. accepted trade-off until
	// user ingredients are resolved to spoonacular ids.
	id := int64(77)
	plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{
		spoonSlot(1, 555),
		userRecipeSlot(2, &id),
	}}

	items := runGenerate(t, plan,
		func(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
			return []recipeclient.RecipeWithIngredients{{ID: 555, Ingredients: []recipeclient.Ingredient{
				{ID: flourID, Name: "flour", Amount: 100, Unit: "ml"},
			}}}, nil
		},
		map[int64][]models.UserRecipeIngredient{
			77: {{Name: "flour", Amount: 50, Unit: "ml"}},
		})

	if countItems(items, "flour") != 2 {
		t.Fatalf("expected the identified and user flour to stay apart, got %+v", items)
	}
	for _, item := range items {
		if item.Source != "meal_plan" {
			t.Errorf("expected source meal_plan, got %q", item.Source)
		}
	}
}

func TestGenerateShoppingListKeepsSameIngredientInDifferentUnitsApart(t *testing.T) {
	// no conversion exists between cloves and grams, so these cannot be summed
	plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{spoonSlot(1, 555)}}

	items := runGenerate(t, plan,
		func(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
			return []recipeclient.RecipeWithIngredients{{ID: 555, Ingredients: []recipeclient.Ingredient{
				{ID: garlicID, Name: "garlic", Amount: 2, Unit: "clove"},
				{ID: garlicID, Name: "garlic", Amount: 30, Unit: "g"},
			}}}, nil
		}, nil)

	if len(items) != 2 {
		t.Fatalf("expected cloves and grams to stay separate, got %+v", items)
	}
}

func TestGenerateShoppingListDropsSizeUnitsBeforeAggregating(t *testing.T) {
	// "2 large onions" and "1 onion" are the same shopping line
	plan := &models.MealPlan{ID: 10, UserID: 7, Recipes: []models.MealPlanRecipe{spoonSlot(1, 555)}}

	items := runGenerate(t, plan,
		func(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
			return []recipeclient.RecipeWithIngredients{{ID: 555, Ingredients: []recipeclient.Ingredient{
				{ID: onionID, Name: "onion", Amount: 2, Unit: "large"},
				{ID: onionID, Name: "onion", Amount: 1, Unit: ""},
			}}}, nil
		}, nil)

	if len(items) != 1 {
		t.Fatalf("expected the size unit to be dropped and the onions merged, got %+v", items)
	}
	assertAmount(t, items[0], 3, "")
}

func TestIsValidUnit(t *testing.T) {
	cases := map[string]bool{
		"g": true, "cup": true, "": true,
		"serving": false, "servings": false, "  SERVINGS  ": false,
	}
	for unit, want := range cases {
		if got := isValidUnit(unit); got != want {
			t.Errorf("isValidUnit(%q) = %v, want %v", unit, got, want)
		}
	}
}

func TestIsValidWholeItem(t *testing.T) {
	cases := map[string]bool{
		"clove": true, "g": true, "": true,
		"large": false, "medium": false, "small": false,
		"larges": false, "mediums": false, "smalls": false, " Large ": false,
	}
	for item, want := range cases {
		if got := isValidWholeItem(item); got != want {
			t.Errorf("isValidWholeItem(%q) = %v, want %v", item, got, want)
		}
	}
}
