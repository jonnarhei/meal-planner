package main

import (
	"context"

	"github.com/jonnarhei/meal-planner/backend/internal/recipeclient"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

type mockUserStore struct {
	createFn      func(context.Context, *models.User) error
	getByEmailFn  func(context.Context, string) (*models.User, error)
	getByIDFn     func(context.Context, int64) (*models.User, error)
	updatePrefsFn func(context.Context, int64, []string, []string, []string) error
}

func (m *mockUserStore) Create(ctx context.Context, user *models.User) error {
	return m.createFn(ctx, user)
}

func (m *mockUserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.getByEmailFn(ctx, email)
}

func (m *mockUserStore) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	return m.getByIDFn(ctx, userID)
}

func (m *mockUserStore) UpdatePreferences(ctx context.Context, userID int64, preferences []string, intolerances []string, excludedIngredients []string) error {
	return m.updatePrefsFn(ctx, userID, preferences, intolerances, excludedIngredients)
}

type mockRecipeStore struct {
	createFn                    func(context.Context, *models.UserRecipe) error
	getAllByUserFn              func(context.Context, int64) ([]models.UserRecipe, error)
	getByIDFn                   func(context.Context, int64, int64) (*models.UserRecipe, error)
	updateFn                    func(context.Context, *models.UserRecipe) error
	deleteFn                    func(context.Context, int64, int64) error
	getIngredientsByRecipeIDsFn func(context.Context, int64, []int64) (map[int64][]models.UserRecipeIngredient, error)
}

func (m *mockRecipeStore) Create(ctx context.Context, recipe *models.UserRecipe) error {
	return m.createFn(ctx, recipe)
}

func (m *mockRecipeStore) GetAllByUser(ctx context.Context, userID int64) ([]models.UserRecipe, error) {
	return m.getAllByUserFn(ctx, userID)
}

func (m *mockRecipeStore) GetByID(ctx context.Context, recipeID, userID int64) (*models.UserRecipe, error) {
	return m.getByIDFn(ctx, recipeID, userID)
}

func (m *mockRecipeStore) Update(ctx context.Context, recipe *models.UserRecipe) error {
	return m.updateFn(ctx, recipe)
}

func (m *mockRecipeStore) Delete(ctx context.Context, recipeID, userID int64) error {
	return m.deleteFn(ctx, recipeID, userID)
}

func (m *mockRecipeStore) GetIngredientsByRecipeIDs(ctx context.Context, userID int64, recipeIDs []int64) (map[int64][]models.UserRecipeIngredient, error) {
	return m.getIngredientsByRecipeIDsFn(ctx, userID, recipeIDs)
}

type mockMealPlanStore struct {
	createFn             func(context.Context, *models.MealPlan) error
	getCurrentFn         func(context.Context, int64) (*models.MealPlan, error)
	updateRecipeForDayFn func(context.Context, *models.MealPlanRecipe) error
	deleteCurrentFn      func(context.Context, int64) error
}

func (m *mockMealPlanStore) Create(ctx context.Context, mealPlan *models.MealPlan) error {
	return m.createFn(ctx, mealPlan)
}

func (m *mockMealPlanStore) GetCurrent(ctx context.Context, userID int64) (*models.MealPlan, error) {
	return m.getCurrentFn(ctx, userID)
}

func (m *mockMealPlanStore) UpdateRecipeForDay(ctx context.Context, recipe *models.MealPlanRecipe) error {
	return m.updateRecipeForDayFn(ctx, recipe)
}

func (m *mockMealPlanStore) DeleteCurrent(ctx context.Context, userID int64) error {
	return m.deleteCurrentFn(ctx, userID)
}

type mockShoppinglistStore struct {
	addItemsFn       func(context.Context, int64, []models.ShoppinglistItem) error
	getAllFn         func(context.Context, int64) ([]models.ShoppinglistItem, error)
	toggleCheckedFn  func(context.Context, int64, int64) error
	deleteItemFn     func(context.Context, int64, int64) error
	deleteCheckedFn  func(context.Context, int64) error
	deleteBySourceFn func(context.Context, int64, string) error
}

func (m *mockShoppinglistStore) AddItems(ctx context.Context, userID int64, items []models.ShoppinglistItem) error {
	return m.addItemsFn(ctx, userID, items)
}

func (m *mockShoppinglistStore) GetAll(ctx context.Context, userID int64) ([]models.ShoppinglistItem, error) {
	return m.getAllFn(ctx, userID)
}

func (m *mockShoppinglistStore) ToggleChecked(ctx context.Context, itemID, userID int64) error {
	return m.toggleCheckedFn(ctx, itemID, userID)
}

func (m *mockShoppinglistStore) DeleteItem(ctx context.Context, itemID, userID int64) error {
	return m.deleteItemFn(ctx, itemID, userID)
}

func (m *mockShoppinglistStore) DeleteChecked(ctx context.Context, userID int64) error {
	return m.deleteCheckedFn(ctx, userID)
}

func (m *mockShoppinglistStore) DeleteBySource(ctx context.Context, userID int64, source string) error {
	return m.deleteBySourceFn(ctx, userID, source)
}

type mockRecipeClient struct {
	getRandomRecipesFn         func(context.Context, int, []string, []string, []string) ([]recipeclient.Recipe, error)
	getRecipeInformationBulkFn func(context.Context, []int64) ([]recipeclient.RecipeWithIngredients, error)
}

func (m *mockRecipeClient) GetRandomRecipes(ctx context.Context, n int, preferences, intolerances, excludedIngredients []string) ([]recipeclient.Recipe, error) {
	return m.getRandomRecipesFn(ctx, n, preferences, intolerances, excludedIngredients)
}

func (m *mockRecipeClient) GetRecipeInformationBulk(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
	return m.getRecipeInformationBulkFn(ctx, ids)
}
