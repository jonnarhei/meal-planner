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
	}
	Mealplans interface {
		Create(context.Context, *models.MealPlan) error
		GetCurrent(context.Context, int64) (*models.MealPlan, error)
	}
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Users: &UsersStore{db},
		Mealplans: &MealPlanStore{db},
	}
}
