package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jonnarhei/meal-planner/backend/internal/store"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

func TestRegisterUserHandler(t *testing.T) {
	t.Run("valid registration", func(t *testing.T) {
		app := newTestApp(t, store.Storage{
			Users: &mockUserStore{
				createFn: func(ctx context.Context, u *models.User) error {
					return nil
				},
			},
		})

		body, _ := json.Marshal(map[string]string{
			"email":    "test@test.com",
			"password": "password123",
		})

		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := executeRequest(req, app)
		checkResponseCode(t, http.StatusCreated, rr.Code)
	})

	t.Run("invalid email", func(t *testing.T) {
		app := newTestApp(t, store.Storage{
			Users: &mockUserStore{},
		})

		body, _ := json.Marshal(map[string]string{
			"email":    "notAnEmail",
			"password": "password123",
		})

		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := executeRequest(req, app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing password", func(t *testing.T) {
		app := newTestApp(t, store.Storage{
			Users: &mockUserStore{},
		})

		body, _ := json.Marshal(map[string]string{
			"email": "test@test.com",
		})

		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := executeRequest(req, app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})
}

func TestLoginHandler(t *testing.T) {
	t.Run("valid login", func(t *testing.T) {
		user := &models.User{
			ID: 1,
			Email: "test@test.com",
		}
		user.SetPassword("password123")

		app := newTestApp(t, store.Storage{
			Users: &mockUserStore{
				getByEmailFn: func(ctx context.Context, s string) (*models.User, error) {
					return user, nil
				},
			},
		})

		body, _ := json.Marshal(map[string]string{
			"email": "test@test.com",
			"password": "password123",
		})

		req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := executeRequest(req, app)
		checkResponseCode(t, http.StatusOK, rr.Code)
	})

	t.Run("wrong password", func(t *testing.T) {
		user := &models.User{
			ID: 1,
			Email: "test@test.com",
		}
		user.SetPassword("password123")

		app := newTestApp(t, store.Storage{
			Users: &mockUserStore{
				getByEmailFn: func(ctx context.Context, s string) (*models.User, error) {
					return user, nil
				},
			},
		})

		body, _ := json.Marshal(map[string]string{
			"email": "test@test.com",
			"password": "wrongPassword",
		})

		req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := executeRequest(req, app)
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})
}