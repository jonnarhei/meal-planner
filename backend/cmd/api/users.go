package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jonnarhei/meal-planner/backend/internal/jsonutil"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

type registerUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload registerUserPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := &models.User{
		Email: payload.Email,
	}

	if err := user.SetPassword(payload.Password); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := app.store.Users.Create(r.Context(), user); err != nil {
		slog.Error("failed to create user", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusCreated, user)
}

func (app *application) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := app.store.Users.ListUsers(r.Context())
	if err != nil {
		slog.Error("failed to list users", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, users)
}