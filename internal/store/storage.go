package store

import (
	"context"
	"database/sql"
	"meal-planner-demo-backend/internal/store/models"
)

type Storage struct {
	Users interface {
		Create(context.Context, *models.User) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users: &UsersStore{db},
	}
}
