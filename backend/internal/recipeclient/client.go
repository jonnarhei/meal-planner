package recipeclient

import "context"

type Recipe struct {
	RecipeID    int64
	Title       string
	Image       string
	URL         string
	Ingredients []Ingredient
}

type Ingredient struct {
	Name   string
	Amount float64
	Unit   string
}

type RecipeWithIngredients struct {
	ID          int64
	Ingredients []Ingredient
}

type Client interface {
	GetRandomRecipes(ctx context.Context, n int, preferences []string) ([]Recipe, error)
	GetRecipeInformationBulk(ctx context.Context, ids []int64) ([]RecipeWithIngredients, error)
}
