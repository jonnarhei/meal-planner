package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jonnarhei/meal-planner/backend/internal/jsonutil"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

type recipePayload struct {
	Title        string                    `json:"title"`
	Image        string                    `json:"image"`
	SourceUrl    string                    `json:"source_url"`
	Instructions string                    `json:"instructions"`
	Servings     int                       `json:"servings"`
	Ingredients  []recipeIngredientPayload `json:"ingredients"`
}

type recipeIngredientPayload struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

func validateRecipePayload(p recipePayload) error {
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("title is required")
	}
	if len(p.Title) > 200 {
		return errors.New("title is too long, max 200 characters")
	}

	if len(p.Ingredients) > 100 {
		return errors.New("too many ingredients")
	}

	for _, ingredient := range p.Ingredients {
		if strings.TrimSpace(ingredient.Name) == "" {
			return errors.New("ingredient name is required")
		}

		if ingredient.Amount < 0 {
			return errors.New("ingredient amount cannot be negative")
		}
	}

	if len(p.Instructions) > 5000 {
		return errors.New("instructions are too long, max 5000 characters")
	}

	if len(p.SourceUrl) > 500 {
		return errors.New("url is too long, max 500 characters")
	}

	if p.Servings < 0 {
		return errors.New("servings cannot be negative")
	}

	return nil
}

func (app *application) createRecipeHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	var payload recipePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonutil.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateRecipePayload(payload); err != nil {
		jsonutil.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ingredients := make([]models.UserRecipeIngredient, len(payload.Ingredients))
	for i, ing := range payload.Ingredients {
		ingredients[i] = models.UserRecipeIngredient{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}

	recipe := &models.UserRecipe{
		UserID:       claims.UserID,
		Title:        payload.Title,
		Image:        payload.Image,
		SourceUrl:    payload.SourceUrl,
		Instructions: payload.Instructions,
		Servings:     payload.Servings,
		Ingredients:  ingredients,
	}

	if err := app.store.Recipes.Create(r.Context(), recipe); err != nil {
		slog.Error("failed to create recipe", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusCreated, recipe)
}

func (app *application) listRecipesHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	recipes, err := app.store.Recipes.GetAllByUser(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("failed to get recipes from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if recipes == nil {
		recipes = []models.UserRecipe{}
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, recipes)
}

func (app *application) getRecipeHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		jsonutil.WriteError(w, "invalid id", http.StatusBadRequest)
		return
	}

	recipe, err := app.store.Recipes.GetByID(r.Context(), id, claims.UserID)
	if err == sql.ErrNoRows {
		jsonutil.WriteError(w, "no existing recipe found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("failed to get recipes from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, recipe)
}

func (app *application) updateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	recipeID, ok := getIntParamUrl(r, "id")
	if !ok {
		jsonutil.WriteError(w, "invalid id", http.StatusBadRequest)
		return
	}

	var payload recipePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonutil.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateRecipePayload(payload); err != nil {
		jsonutil.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ingredients := make([]models.UserRecipeIngredient, len(payload.Ingredients))
	for i, ing := range payload.Ingredients {
		ingredients[i] = models.UserRecipeIngredient{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}

	recipe := &models.UserRecipe{
		ID:           recipeID,
		UserID:       claims.UserID,
		Title:        payload.Title,
		Image:        payload.Image,
		SourceUrl:    payload.SourceUrl,
		Instructions: payload.Instructions,
		Servings:     payload.Servings,
		Ingredients:  ingredients,
	}

	err := app.store.Recipes.Update(r.Context(), recipe)
	if err == sql.ErrNoRows {
		jsonutil.WriteError(w, "no existing recipe found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("failed to update recipe", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	updatedRecipe, err := app.store.Recipes.GetByID(r.Context(), recipe.ID, claims.UserID)
	if err != nil {
		slog.Error("failed to refetch updated recipe from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, updatedRecipe)
}

func (app *application) deleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	id, ok := getIntParamUrl(r, "id")
	if !ok {
		jsonutil.WriteError(w, "invalid id", http.StatusBadRequest)
		return
	}

	err := app.store.Recipes.Delete(r.Context(), id, claims.UserID)
	if err == sql.ErrNoRows {
		jsonutil.WriteError(w, "no existing recipe found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("failed to delete recipe from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
