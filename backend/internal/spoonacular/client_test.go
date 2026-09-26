package spoonacular

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRecipeInformationBulkDecodesIngredients(t *testing.T) {
	const body = `[{"id":555,"extendedIngredients":[
		{"id":11215,"name":"garlic cloves","nameClean":"garlic","measures":{"metric":{"amount":2,"unitShort":"cloves"}}},
		{"id":1012010,"name":"ground cinnamon","nameClean":"","measures":{"metric":{"amount":1,"unitShort":"tsp"}}}
	]}]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("ids"); got != "555,556" {
			t.Errorf("expected ids=555,556, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.baseURL = server.URL

	recipes, err := client.GetRecipeInformationBulk(context.Background(), []int64{555, 556})
	if err != nil {
		t.Fatalf("GetRecipeInformationBulk: %v", err)
	}

	if len(recipes) != 1 || recipes[0].ID != 555 {
		t.Fatalf("expected recipe 555, got %+v", recipes)
	}
	if len(recipes[0].Ingredients) != 2 {
		t.Fatalf("expected 2 ingredients, got %+v", recipes[0].Ingredients)
	}

	garlic := recipes[0].Ingredients[0]
	if garlic.ID != 11215 {
		t.Errorf("expected ingredient id 11215, got %d", garlic.ID)
	}
	if garlic.Name != "garlic" {
		t.Errorf("expected nameClean to win, got %q", garlic.Name)
	}
	if garlic.Amount != 2 || garlic.Unit != "cloves" {
		t.Errorf("expected 2 cloves, got %v %q", garlic.Amount, garlic.Unit)
	}

	// nameClean is sometimes empty; name is the fallback
	if recipes[0].Ingredients[1].Name != "ground cinnamon" {
		t.Errorf("expected the raw name as fallback, got %q", recipes[0].Ingredients[1].Name)
	}
}
