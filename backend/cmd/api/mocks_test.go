package main

import (
	"context"

	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

type mockUserStore struct {
	createFn func(context.Context, *models.User) error
	getByEmailFn func(context.Context, string) (*models.User, error)
	getByIDFn func(context.Context, int64) (*models.User, error)
	updatePrefsFn func(context.Context, int64, []string) error
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

func (m *mockUserStore) UpdatePreferences(ctx context.Context, userID int64, preferences []string) error {
	return m.updatePrefsFn(ctx, userID, preferences)
}