package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
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
		jsonutil.WriteError(w, "bad request", http.StatusBadRequest)
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
		ID:          currentPlan.Recipes[payload.Day-1].ID,
		MealPlanID:  currentPlan.ID,
		RecipeID:    recipe.RecipeID,
		RecipeTitle: recipe.Title,
		Image:       recipe.Image,
		SourceURL:   recipe.URL,
		Day:         int(payload.Day),
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

	if err := app.store.Shoppinglist.DeleteBySource(r.Context(), claims.UserID, "meal_plan"); err != nil {
		slog.Error("failed to delete old shopping list items", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

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

func (app *application) getShoppingListHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	shoppingList, err := app.store.Shoppinglist.GetAll(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("failed to get shopping list from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteHttpJson(w, http.StatusOK, shoppingList)
}

type addShoppingListItemPayload struct {
	Items []struct {
		Name   string  `json:"name"`
		Amount float64 `json:"amount"`
		Unit   string  `json:"unit"`
	} `json:"items"`
}

func (app *application) addShoppingListItemsHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	var payload addShoppingListItemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonutil.WriteError(w, "bad request", http.StatusBadRequest)
		return 
	}

	if len(payload.Items) == 0 {
		jsonutil.WriteError(w, "no items provided", http.StatusBadRequest)
		return 
	}

	items := make([]models.ShoppinglistItem, len(payload.Items))
	for index, item := range payload.Items {
		items[index] = models.ShoppinglistItem{
			UserID: claims.UserID,
			Name: item.Name,
			Amount: item.Amount,
			Unit: item.Unit,
			Source: "manual",
		}
	}

	if err := app.store.Shoppinglist.AddItems(r.Context(), claims.UserID, items); err != nil {
		slog.Error("failed to add shopping list items to db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return 
	}

	w.WriteHeader(http.StatusCreated)
}

func (app *application) addFromMealPlanHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	if err := app.store.Shoppinglist.DeleteBySource(r.Context(), claims.UserID, "meal_plan"); err != nil {
		slog.Error("failed to delete old shopping list items", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	plan, err := app.store.Mealplans.GetCurrent(r.Context(), claims.UserID)
	if err == sql.ErrNoRows {
		jsonutil.WriteError(w, "no active meal plan found", http.StatusNotFound)
	}
	if err != nil {
		slog.Error("failed to get current mealplan from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	ids := make([]int64, len(plan.Recipes))
	for index, recipe := range plan.Recipes {
		ids[index] = recipe.RecipeID
	}

	recipes, err := app.spoonacular.GetIngredientsForRecipes(r.Context(), ids)
	if err != nil {
		slog.Error("failed to get ingredients information for recipes", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var items []models.ShoppinglistItem
	for _, recipe := range recipes {
		for _, ingredient := range recipe.ExtendedIngredients {
			items = append(items, models.ShoppinglistItem{
				UserID: claims.UserID,
				Name: ingredient.Name,
				Amount: ingredient.Measures.Metric.Amount,
				Unit: ingredient.Measures.Metric.UnitLong,
				Source: "meal_plan",
			})
		}
	}

	if err := app.store.Shoppinglist.AddItems(r.Context(), claims.UserID, items); err != nil {
		slog.Error("failed to add shoppin list items to db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (app *application) toggleCheckedHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		jsonutil.WriteError(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := app.store.Shoppinglist.ToggleChecked(r.Context(), id, claims.UserID); err != nil {
		slog.Error("failed to toggle item in db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return 
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		jsonutil.WriteError(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := app.store.Shoppinglist.DeleteItem(r.Context(), id, claims.UserID); err != nil {
		slog.Error("failed to delete item from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return 
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) deleteCheckedHandler(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	if err := app.store.Shoppinglist.DeleteChecked(r.Context(), claims.UserID); err != nil {
		slog.Error("failed to delete checked items from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return 
	}

	w.WriteHeader(http.StatusNoContent)
}