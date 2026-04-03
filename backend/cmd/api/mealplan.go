package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jonnarhei/meal-planner/backend/internal/jsonutil"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

func (app *application) generateMealPlan(ctx context.Context, userID int64, preferences []string) (*models.MealPlan, error) {
	randomRecipes, err := app.spoonacular.GetRandomRecipes(ctx, 7, preferences)

	slog.Info("recipes fetched", "count", len(randomRecipes.Recipes))

	if err != nil {
		return nil, err
	}

	var mealPlanRecipes []models.MealPlanRecipe

	for index, recipe := range randomRecipes.Recipes {
		mealPlanRecipe := models.MealPlanRecipe{
			RecipeID:    recipe.RecipeID,
			RecipeTitle: recipe.Title,
			Image:       recipe.Image,
			SourceURL:   recipe.URL,
			Day:         index + 1,
		}
		mealPlanRecipes = append(mealPlanRecipes, mealPlanRecipe)
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)

	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	sunday := monday.AddDate(0, 0, 6)

	mealPlan := &models.MealPlan{
		UserID:    userID,
		StartDate: monday,
		EndDate:   sunday,
		Recipes:   mealPlanRecipes,
	}

	return mealPlan, nil
}

func (app *application) getCurrentMealPlanHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	plan, err := app.store.Mealplans.GetCurrent(r.Context(), claims.UserID)
	//check if there were errors, and ignore no rows error
	if err != nil && err != sql.ErrNoRows {
		slog.Error("failed to get current meal plan", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	//check if there was no plan returned
	if plan == nil {
		user, err := app.store.Users.GetByID(r.Context(), claims.UserID)
		if err != nil {
			slog.Error("failed to get user from db", "error", err)
			jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		plan, err = app.generateMealPlan(r.Context(), claims.UserID, user.DietaryPreferences)
		if err != nil {
			slog.Error("failed to generate meal plan", "error", err)
			jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if err = app.store.Mealplans.Create(r.Context(), plan); err != nil {
			slog.Error("failed to save meal plan in database", "error", err)
			jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
	jsonutil.WriteHttpJson(w, http.StatusOK, plan)
}

type changeRecipePayload struct {
	Day int64 `json:"day"`
}

func (app *application) changeRecipeForDay(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	currentPlan, err := app.store.Mealplans.GetCurrent(r.Context(), claims.UserID)

	if err != nil {
		slog.Error("failed to get current meal plan", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var payload changeRecipePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("Error decoding json into variable", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusBadRequest)
		return
	}

	user, err := app.store.Users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("failed to get user from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	recipeResponse, err := app.spoonacular.GetRandomRecipes(r.Context(), 1, user.DietaryPreferences)
	if err != nil {
		slog.Error("Could not get random recipe from api", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if payload.Day < 1 || int(payload.Day) > len(currentPlan.Recipes) {
		jsonutil.WriteError(w, "invalid day", http.StatusBadRequest)
		return
	}

	recipe := recipeResponse.Recipes[0]

	updatedRecipe := &models.MealPlanRecipe{
		ID: currentPlan.Recipes[payload.Day - 1].ID,
		MealPlanID: currentPlan.ID,
		RecipeID: recipe.RecipeID,
		RecipeTitle: recipe.Title,
		Image: recipe.Image,
		SourceURL: recipe.URL,
		Day: int(payload.Day),
	}

	if err = app.store.Mealplans.UpdateRecipeForDay(r.Context(), updatedRecipe); err != nil {
		slog.Error("Failed to update recipe in the database", "error", err)
		jsonutil.WriteError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, updatedRecipe)
}

func (app *application) regenerateMealPlanHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	
	if err := app.store.Mealplans.DeleteCurrent(r.Context(), claims.UserID); err != nil {
		slog.Error("Could not delete the current mealplan", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := app.store.Users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("failed to get user from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	plan, err := app.generateMealPlan(r.Context(), claims.UserID, user.DietaryPreferences)
	if err != nil {
		slog.Error("could not generate a new mealplan from api", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err = app.store.Mealplans.Create(r.Context(), plan); err != nil {
		slog.Error("Could not create new mealplan database entry")
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, plan)
}