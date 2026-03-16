package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jonnarhei/meal-planner/backend/internal/auth"
	"github.com/jonnarhei/meal-planner/backend/internal/jsonutil"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

// registering a new user from a POST request to the API
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

// mostly for my own curiosity, wouldnt be accessible in a public application
func (app *application) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	slog.Info("request by", "user_id", claims.UserID, "email", claims.Email)

	users, err := app.store.Users.ListUsers(r.Context())
	if err != nil {
		slog.Error("failed to list users", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, users)
}

type loginUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload loginUserPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		slog.Error("failed to get user by email", "error", err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !user.CheckPassword(payload.Password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, app.config.jwt.secret, app.config.jwt.expiry)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, map[string]string{
		"token": token,
	})
}


func (app *application) getMeHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	jsonutil.WriteHttpJson(w, http.StatusOK, map[string]any{
		"id": claims.UserID,
		"email": claims.Email,
	})
}