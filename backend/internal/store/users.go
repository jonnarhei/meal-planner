package store

import (
	"context"
	"database/sql"

	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
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

func (u *UsersStore) ListUsers(ctx context.Context) ([]models.User, error) {
	query := `
	SELECT id, email, password, created_at FROM users
	`

	rows, err := u.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User
		err:= rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
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