package store

import (
	"strings"
	"testing"

	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

func TestBuildInsertValues(t *testing.T) {
	id := int64(11215)
	items := []models.ShoppinglistItem{
		{UserID: 7, IngredientID: &id, Name: "garlic", Amount: 2, Unit: "clove", Source: "meal_plan"},
		{UserID: 7, Name: "flour", Amount: 100, Unit: "g", Source: "manual"},
	}

	values, args := buildInsertValues(items)

	want := "($1, $2, $3, $4, $5, $6),($7, $8, $9, $10, $11, $12)"
	if values != want {
		t.Errorf("expected %s, got %s", want, values)
	}

	// every argument must be bound by exactly one placeholder, or postgres rejects the batch
	if got := strings.Count(values, "$"); got != len(args) {
		t.Errorf("%d placeholders for %d arguments: %s", got, len(args), values)
	}

	if args[1] != &id {
		t.Errorf("expected the ingredient id in position 2, got %v", args[1])
	}
	// a nil *int64 in an any is not equal to nil, so assert on the typed value
	if got, ok := args[7].(*int64); !ok || got != nil {
		t.Errorf("expected a nil ingredient id for the manual item, got %v", args[7])
	}
}

func TestBuildInsertValuesEmpty(t *testing.T) {
	values, args := buildInsertValues(nil)

	if values != "" {
		t.Errorf("expected no values, got %q", values)
	}
	if len(args) != 0 {
		t.Errorf("expected no arguments, got %v", args)
	}
}
