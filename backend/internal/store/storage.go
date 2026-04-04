package store

import (
	"context"
	"database/sql"

	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

type Storage struct {
	Users interface {
		Create(context.Context, *models.User) error
		ListUsers(context.Context) ([]models.User, error)
		GetByEmail(context.Context, string) (*models.User, error)
		GetByID(context.Context, int64) (*models.User, error)
		UpdatePreferences(context.Context, int64, []string) error
	}
	Mealplans interface {
		Create(context.Context, *models.MealPlan) error
		GetCurrent(context.Context, int64) (*models.MealPlan, error)
		UpdateRecipeForDay(context.Context, *models.MealPlanRecipe) error
		DeleteCurrent(context.Context, int64) error
	}
	Shoppinglist interface {
		AddItems(context.Context, int64, []models.ShoppinglistItem) error
		GetAll(context.Context, int64) ([]models.ShoppinglistItem, error)
		ToggleChecked(context.Context, int64, int64) error
		DeleteItem(context.Context, int64, int64) error
		DeleteChecked(context.Context, int64) error
		DeleteBySource(context.Context, int64, string) error
	}
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Users:        &UsersStore{db},
		Mealplans:    &MealPlanStore{db},
		Shoppinglist: &ShoppinglistStore{db},
	}
}
