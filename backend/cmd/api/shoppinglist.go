package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jonnarhei/meal-planner/backend/internal/jsonutil"
	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

func isValidIngredient(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	// check long ingredient names (usually arent ingredients)
	if len(name) > 50 {
		return false
	}
	//most recipes dont contain colons
	if strings.Contains(name, ":") || strings.Contains(name, "(") || strings.Contains(name, ")") {
		return false
	}
	if strings.Contains(strings.ToLower(name), "tsp") || strings.Contains(strings.ToLower(name), "tbsp"){
		return false
	}
	if strings.Contains(strings.ToLower(name), "shopping") || strings.Contains(strings.ToLower(name), "list") {
		return false
	}
	return true
}

func isValidUnit(unit string) bool {
	invalid := []string{"serving", "servings"}
	normalized := strings.ToLower(strings.TrimSpace(unit))
	for _, u := range invalid {
		if normalized == u {
			return false
		}
	}

	return true
}

func normalizeUnit(unit string) string {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "tsp", "tsps", "teaspoon", "teaspoons":
		return "tsp"
	case "tbsp", "tbsps", "tablespoon", "tablespoons":
		return "tbsp"
	case "g", "gram", "grams":
		return "g"
	case "ml", "milliliter", "milliliters", "mls", "millilitre", "millilitres":
		return "ml"
	case "l", "liter", "liters", "litre", "litres":
		return "l"
	case "kg", "kgs", "kilogram", "kilograms":
		return "kg"
	case "oz", "ounce", "ounces":
		return "oz"
	case "lb", "lbs", "pound", "pounds":
		return "lb"
	case "cup", "cups":
		return "cup"
	case "clove", "cloves":
		return "clove"
	case "can", "cans":
		return "can"
	default:
		return strings.ToLower(strings.TrimSpace(unit))
	}
}

type measurement struct {
	amount float64
	unit   string
}

func toBaseUnit(amount float64, unit string) measurement {
	switch unit {
	// turn tablespoons into teaspoons
	case "tbsp":
		return measurement{amount * 3, "tsp"}
	// volume, turn into milliliters
	case "cup":
        return measurement{amount * 236.59, "ml"}
    case "fl oz":
        return measurement{amount * 29.57, "ml"}
    case "l":
        return measurement{amount * 1000, "ml"}
	// weight, turn into grams
    case "oz":
        return measurement{amount * 28.35, "g"}
    case "lb":
        return measurement{amount * 453.59, "g"}
    case "kg":
        return measurement{amount * 1000, "g"}
    default:
        return measurement{amount, unit}
	}
}

// fix for ON DUPLICATE giving errors on batch inserts into db
func duplicateItems(items []models.ShoppinglistItem) []models.ShoppinglistItem {
	type key struct {
		name string
		unit string
	}

	seen := make(map[key]int)
	var result []models.ShoppinglistItem

	for _, item := range items {
		k := key{name: item.Name, unit: item.Unit}
		if idx, exists := seen[k]; exists {
			result[idx].Amount += item.Amount
		} else {
			seen[k] = len(result)
			result = append(result, item)
		}
	}

	return result
}

func (app *application) generateShoppingListFromPlan(ctx context.Context, userID int64, plan *models.MealPlan) error {
	ids := make([]int64, len(plan.Recipes))
	for i, recipe := range plan.Recipes {
		ids[i] = recipe.RecipeID
	}

	start := time.Now()
	recipes, err := app.recipes.GetRecipeInformationBulk(ctx, ids)
	slog.Info("GetRecipeInformationBulk took", "duration", time.Since(start))

	if err != nil {
		return err
	}

	var items []models.ShoppinglistItem
	for _, recipe := range recipes {
		for _, ingredient := range recipe.Ingredients {
			if !isValidIngredient(ingredient.Name) || !isValidUnit(ingredient.Unit) {
				continue
			}
			base := toBaseUnit(ingredient.Amount, normalizeUnit(ingredient.Unit))
			items = append(items, models.ShoppinglistItem{
				UserID: userID,
				Name:   ingredient.Name,
				Amount: base.amount,
				Unit:   base.unit,
				Source: "meal_plan",
			})
		}
	}

	items = duplicateItems(items)

	return app.store.Shoppinglist.AddItems(ctx, userID, items)
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
			Name:   item.Name,
			Amount: item.Amount,
			Unit:   item.Unit,
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
		return
	}
	if err != nil {
		slog.Error("failed to get current mealplan from db", "error", err)
		jsonutil.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := app.generateShoppingListFromPlan(r.Context(), claims.UserID, plan); err != nil {
		slog.Error("failed to generate shopping list from plan", "error", err)
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
