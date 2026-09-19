package main

import (
	"context"

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
	createFn       func(context.Context, *models.UserRecipe) error
	getAllByUserFn func(context.Context, int64) ([]models.UserRecipe, error)
	getByIDFn      func(context.Context, int64, int64) (*models.UserRecipe, error)
	updateFn       func(context.Context, *models.UserRecipe) error
	deleteFn       func(context.Context, int64, int64) error
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
