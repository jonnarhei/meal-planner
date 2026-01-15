package store

import (
	"context"
	"database/sql"
	"meal-planner-demo-backend/internal/store/models"
)

type UsersStore struct {
	db *sql.DB
}

func (u *UsersStore) Create(ctx context.Context, user *models.User) error {
	query := `
	INSERT INTO users (email, password)
	VALUES ($1, $2) RETURNING id, created_at
	`

	err := u.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.Password,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}
