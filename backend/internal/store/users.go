package store

import (
	"context"
	"database/sql"

	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
	"github.com/lib/pq"
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

func (u *UsersStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
	SELECT id, email, password, created_at FROM users
	WHERE email = $1
	`

	user := &models.User{}
	err := u.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UsersStore) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	query := `
	SELECT id, email, created_at, dietary_preferences, intolerances, excluded_ingredients FROM users
	WHERE id = $1
	`

	user := &models.User{}
	err := u.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		pq.Array(&user.DietaryPreferences),
		pq.Array(&user.Intolerances),
		pq.Array(&user.ExcludedIngredients),
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UsersStore) UpdatePreferences(ctx context.Context, userID int64, preferences []string, intolerances []string, excludedIngredients []string) error {
	query := `
	UPDATE users
	SET dietary_preferences = $1, intolerances = $2, excluded_ingredients = $3
	WHERE id = $4
	`

	_, err := u.db.ExecContext(ctx, query, pq.Array(preferences), pq.Array(intolerances), pq.Array(excludedIngredients), userID)
	return err
}