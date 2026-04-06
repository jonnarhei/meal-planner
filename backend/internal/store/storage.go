package store

import (
	"context"
	"database/sql"

	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

type Storage struct {
	Users interface {
		Create(ctx context.Context, user *models.User) error
		GetByEmail(ctx context.Context, email string) (*models.User, error)
		GetByID(ctx context.Context, userID int64) (*models.User, error)
		UpdatePreferences(ctx context.Context, userID int64, preferences []string) error
	}
	Mealplans interface {
		Create( ctx context.Context, mealPlan *models.MealPlan) error
		GetCurrent(ctx context.Context, userID int64) (*models.MealPlan, error)
		UpdateRecipeForDay(ctx context.Context, mealPlanRecipe *models.MealPlanRecipe) error
		DeleteCurrent(ctx context.Context, userID int64) error
	}
	Shoppinglist interface {
		AddItems(ctx context.Context, userID int64, items []models.ShoppinglistItem) error
		GetAll(ctx context.Context, userID int64) ([]models.ShoppinglistItem, error)
		ToggleChecked(ctx context.Context, itemID int64, userID int64) error
		DeleteItem(ctx context.Context, itemID int64, userID int64) error
		DeleteChecked(ctx context.Context, userID int64) error
		DeleteBySource(ctx context.Context, userID int64, source string) error
	}
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Users:        &UsersStore{db},
		Mealplans:    &MealPlanStore{db},
		Shoppinglist: &ShoppinglistStore{db},
	}
}
