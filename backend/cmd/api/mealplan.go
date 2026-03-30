package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/jonnarhei/meal-planner/backend/internal/jsonutil"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

func (app *application) generateMealPlan(ctx context.Context, userID int64) (*models.MealPlan, error) {
	randomRecipes, err := app.spoonacular.GetRandomRecipes(ctx)

	if err != nil {
		return nil, err
	}

	var mealPlanRecipes []models.MealPlanRecipe

	for index, recipe := range randomRecipes.Recipes {
		mealPlanRecipe := models.MealPlanRecipe{
			RecipeID: recipe.RecipeID,
			RecipeTitle: recipe.Title,
			Image: recipe.Image,
			SourceURL: recipe.URL,
			Day: index + 1,
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
		UserID: userID,
		StartDate: monday,
		EndDate: sunday,
		Recipes: mealPlanRecipes,
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
		plan, err = app.generateMealPlan(r.Context(), claims.UserID)
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