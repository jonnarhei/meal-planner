package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/jonnarhei/meal-planner/backend/internal/store"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

func TestValidateRecipePayload(t *testing.T) {
	tests := []struct {
		name    string
		payload recipePayload
		wantErr bool
	}{
		{"valid full", recipePayload{Title: "Pancakes", Servings: 4,
			Ingredients: []recipeIngredientPayload{{Name: "flour", Amount: 1.5, Unit: "cup"}}}, false},
		{"valid no ingredients", recipePayload{Title: "Pancakes"}, false},
		{"missing title", recipePayload{}, true},
		{"whitespace title", recipePayload{Title: "   "}, true},
		{"title too long", recipePayload{Title: strings.Repeat("a", 201)}, true},
		{"title at limit", recipePayload{Title: strings.Repeat("a", 200)}, false},
		{"negative servings", recipePayload{Title: "x", Servings: -1}, true},
		{"blank ingredient name", recipePayload{Title: "x",
			Ingredients: []recipeIngredientPayload{{Name: " "}}}, true},
		{"negative amount", recipePayload{Title: "x",
			Ingredients: []recipeIngredientPayload{{Name: "salt", Amount: -1}}}, true},
		{"too many ingredients", recipePayload{Title: "x",
			Ingredients: make([]recipeIngredientPayload, 101)}, true},
		{"instructions too long", recipePayload{Title: "x", Instructions: strings.Repeat("a", 5001)}, true},
		{"source url too long", recipePayload{Title: "x", SourceUrl: strings.Repeat("a", 501)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRecipePayload(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestCreateRecipeHandler(t *testing.T) {
	t.Run("valid recipe uses token user and returns created recipe", func(t *testing.T) {
		var got *models.UserRecipe
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			createFn: func(ctx context.Context, r *models.UserRecipe) error {
				got = r
				r.ID = 42
				return nil
			},
		}})

		body := validRecipeBody()
		body["user_id"] = 999 //is ignored

		rr := executeRequest(newJSONRequest(t, http.MethodPost, "/recipes", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusCreated, rr.Code)

		if got.UserID != 7 {
			t.Errorf("expected user id 7 from token, got %d", got.UserID)
		}
		if len(got.Ingredients) != 1 || got.Ingredients[0].Amount != 1.5 {
			t.Errorf("ingredients not mapped correctly: %+v", got.Ingredients)
		}

		var resp models.UserRecipe
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if resp.ID != 42 {
			t.Errorf("expected response id 42, got %d", resp.ID)
		}
	})

	t.Run("no ingredients is allowed", func(t *testing.T) {
		var got *models.UserRecipe
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			createFn: func(ctx context.Context, r *models.UserRecipe) error {
				got = r
				return nil
			},
		}})

		body := validRecipeBody()
		delete(body, "ingredients")

		rr := executeRequest(newJSONRequest(t, http.MethodPost, "/recipes", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusCreated, rr.Code)

		if got == nil {
			t.Fatal("expected store.Create to be called")
		}
		if len(got.Ingredients) != 0 {
			t.Errorf("expected no ingredients, got %+v", got.Ingredients)
		}
	})

	t.Run("missing auth header", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		rr := executeRequest(newJSONRequest(t, http.MethodPost, "/recipes", validRecipeBody(), ""), app)
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("malformed json", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		req, err := http.NewRequest(http.MethodPost, "/recipes", strings.NewReader("{not json"))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeaderFor(t, app, 7))

		rr := executeRequest(req, app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid payload never reaches store", func(t *testing.T) {
		// nil createFn would panic (and surface as a 500) if the handler called the store
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		body := validRecipeBody()
		body["title"] = ""

		rr := executeRequest(newJSONRequest(t, http.MethodPost, "/recipes", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("store error", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			createFn: func(ctx context.Context, r *models.UserRecipe) error {
				return errors.New("boom")
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodPost, "/recipes", validRecipeBody(), authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestListRecipesHandler(t *testing.T) {
	t.Run("returns only the token user's recipes", func(t *testing.T) {
		var gotUserID int64
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			getAllByUserFn: func(ctx context.Context, userID int64) ([]models.UserRecipe, error) {
				gotUserID = userID
				return []models.UserRecipe{{ID: 1, Title: "a"}, {ID: 2, Title: "b"}}, nil
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodGet, "/recipes", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusOK, rr.Code)

		if gotUserID != 7 {
			t.Errorf("expected store to be queried with user id 7, got %d", gotUserID)
		}

		var resp []models.UserRecipe
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if len(resp) != 2 {
			t.Errorf("expected 2 recipes, got %d", len(resp))
		}
	})

	t.Run("empty list encodes as [] not null", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			getAllByUserFn: func(ctx context.Context, userID int64) ([]models.UserRecipe, error) {
				return nil, nil
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodGet, "/recipes", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusOK, rr.Code)

		if got := strings.TrimSpace(rr.Body.String()); got != "[]" {
			t.Errorf("expected body [], got %s", got)
		}
	})

	t.Run("missing auth header", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		rr := executeRequest(newJSONRequest(t, http.MethodGet, "/recipes", nil, ""), app)
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("store error", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			getAllByUserFn: func(ctx context.Context, userID int64) ([]models.UserRecipe, error) {
				return nil, errors.New("boom")
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodGet, "/recipes", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestGetRecipeHandler(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		var gotID, gotUserID int64
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			getByIDFn: func(ctx context.Context, id, userID int64) (*models.UserRecipe, error) {
				gotID, gotUserID = id, userID
				return &models.UserRecipe{ID: id, UserID: userID, Title: "Pancakes"}, nil
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodGet, "/recipes/5", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusOK, rr.Code)

		if gotID != 5 || gotUserID != 7 {
			t.Errorf("expected store called with id 5 / user 7, got %d / %d", gotID, gotUserID)
		}

		var resp models.UserRecipe
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if resp.Title != "Pancakes" {
			t.Errorf("expected title Pancakes, got %q", resp.Title)
		}
	})

	t.Run("not found or not owned", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			getByIDFn: func(ctx context.Context, id, userID int64) (*models.UserRecipe, error) {
				return nil, sql.ErrNoRows
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodGet, "/recipes/5", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusNotFound, rr.Code)
	})

	t.Run("non-numeric id", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		rr := executeRequest(newJSONRequest(t, http.MethodGet, "/recipes/abc", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing auth header", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		rr := executeRequest(newJSONRequest(t, http.MethodGet, "/recipes/5", nil, ""), app)
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("store error", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			getByIDFn: func(ctx context.Context, id, userID int64) (*models.UserRecipe, error) {
				return nil, errors.New("boom")
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodGet, "/recipes/5", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestUpdateRecipeHandler(t *testing.T) {
	t.Run("passes url id and token user to store, returns refetched recipe", func(t *testing.T) {
		var updated *models.UserRecipe
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			updateFn: func(ctx context.Context, r *models.UserRecipe) error {
				updated = r
				return nil
			},
			getByIDFn: func(ctx context.Context, id, userID int64) (*models.UserRecipe, error) {
				return &models.UserRecipe{ID: id, UserID: userID, Title: "Pancakes v2"}, nil
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodPut, "/recipes/5", validRecipeBody(), authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusOK, rr.Code)

		if updated == nil {
			t.Fatal("expected store.Update to be called")
		}
		if updated.ID != 5 || updated.UserID != 7 {
			t.Errorf("expected id 5 / user 7, got %d / %d", updated.ID, updated.UserID)
		}
		if len(updated.Ingredients) != 1 {
			t.Errorf("expected 1 ingredient, got %d", len(updated.Ingredients))
		}

		var resp models.UserRecipe
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if resp.Title != "Pancakes v2" {
			t.Errorf("expected refetched recipe in response, got title %q", resp.Title)
		}
	})

	t.Run("not found or not owned", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			updateFn: func(ctx context.Context, r *models.UserRecipe) error {
				return sql.ErrNoRows
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodPut, "/recipes/5", validRecipeBody(), authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusNotFound, rr.Code)
	})

	t.Run("non-numeric id", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		rr := executeRequest(newJSONRequest(t, http.MethodPut, "/recipes/abc", validRecipeBody(), authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid payload never reaches store", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		body := validRecipeBody()
		body["title"] = ""

		rr := executeRequest(newJSONRequest(t, http.MethodPut, "/recipes/5", body, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("malformed json", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		req, err := http.NewRequest(http.MethodPut, "/recipes/5", strings.NewReader("{not json"))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeaderFor(t, app, 7))

		rr := executeRequest(req, app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing auth header", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		rr := executeRequest(newJSONRequest(t, http.MethodPut, "/recipes/5", validRecipeBody(), ""), app)
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("update store error", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			updateFn: func(ctx context.Context, r *models.UserRecipe) error {
				return errors.New("boom")
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodPut, "/recipes/5", validRecipeBody(), authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("refetch after update fails", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			updateFn: func(ctx context.Context, r *models.UserRecipe) error {
				return nil
			},
			getByIDFn: func(ctx context.Context, id, userID int64) (*models.UserRecipe, error) {
				return nil, errors.New("boom")
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodPut, "/recipes/5", validRecipeBody(), authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestDeleteRecipeHandler(t *testing.T) {
	t.Run("success passes url id and token user to store", func(t *testing.T) {
		var gotID, gotUserID int64
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			deleteFn: func(ctx context.Context, id, userID int64) error {
				gotID, gotUserID = id, userID
				return nil
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodDelete, "/recipes/5", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusNoContent, rr.Code)

		if gotID != 5 || gotUserID != 7 {
			t.Errorf("expected store called with id 5 / user 7, got %d / %d", gotID, gotUserID)
		}
	})

	t.Run("not found or not owned", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			deleteFn: func(ctx context.Context, id, userID int64) error {
				return sql.ErrNoRows
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodDelete, "/recipes/5", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusNotFound, rr.Code)
	})

	t.Run("non-numeric id", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		rr := executeRequest(newJSONRequest(t, http.MethodDelete, "/recipes/abc", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing auth header", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{}})

		rr := executeRequest(newJSONRequest(t, http.MethodDelete, "/recipes/5", nil, ""), app)
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("store error", func(t *testing.T) {
		app := newTestApp(t, store.Storage{Recipes: &mockRecipeStore{
			deleteFn: func(ctx context.Context, id, userID int64) error {
				return errors.New("boom")
			},
		}})

		rr := executeRequest(newJSONRequest(t, http.MethodDelete, "/recipes/5", nil, authHeaderFor(t, app, 7)), app)
		checkResponseCode(t, http.StatusInternalServerError, rr.Code)
	})
}
