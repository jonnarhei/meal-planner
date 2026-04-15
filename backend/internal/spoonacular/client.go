package spoonacular

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jonnarhei/meal-planner/backend/internal/recipeclient"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    "https://api.spoonacular.com",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type spoonacularRecipe struct {
	ID                  int64        `json:"id"`
	Title               string       `json:"title"`
	Image               string       `json:"image"`
	SourceURL           string       `json:"sourceUrl"`
	ExtendedIngredients []ingredient `json:"extendedIngredients"`
}

type complexSearchResponse struct {
	Results []spoonacularRecipe `json:"results"`
}

type measures struct {
	Amount    float64 `json:"amount"`
	UnitShort string  `json:"unitShort"`
}

type metricMeasures struct {
	Metric measures `json:"metric"`
}

type ingredient struct {
	Name     string         `json:"name"`
	Measures metricMeasures `json:"measures"`
}

type recipeInformation struct {
	ID                  int64        `json:"id"`
	ExtendedIngredients []ingredient `json:"extendedIngredients"`
}

func (c *Client) get(ctx context.Context, path string, params url.Values, target any) error {
	params.Set("apiKey", c.apiKey)
	url := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

type Recipe struct {
	RecipeID int64  `json:"id"`
	Title    string `json:"title"`
	Image    string `json:"image"`
	URL      string `json:"sourceUrl"`
}

type RandomRecipeResponse struct {
	Recipes []Recipe `json:"recipes"`
}

func (c *Client) GetRandomRecipes(ctx context.Context, n int, preferences []string, intolerances []string, excludedIngredients []string) ([]recipeclient.Recipe, error) {
	var response complexSearchResponse

	params := url.Values{}
	params.Set("number", strconv.Itoa(n))
	params.Set("sort", "random")
	params.Set("type", "main course")
	params.Set("diet", strings.Join(preferences, ","))
	params.Set("intolerances", strings.Join(intolerances, ","))
	params.Set("excludeIngredients", strings.Join(excludedIngredients, ","))
	params.Set("addRecipeInformation", "true")

	err := c.get(ctx, "/recipes/complexSearch", params, &response)
	if err != nil {
		return nil, err
	}

	recipes := make([]recipeclient.Recipe, len(response.Results))
	for i, r := range response.Results {
		ingredients := make([]recipeclient.Ingredient, 0, len(r.ExtendedIngredients))
		for _, ingr := range r.ExtendedIngredients {
			ingredients = append(ingredients, recipeclient.Ingredient{
				Name:   ingr.Name,
				Amount: ingr.Measures.Metric.Amount,
				Unit:   ingr.Measures.Metric.UnitShort,
			})
		}

		recipes[i] = recipeclient.Recipe{
			RecipeID:    r.ID,
			Title:       r.Title,
			Image:       r.Image,
			URL:         r.SourceURL,
			Ingredients: ingredients,
		}
	}

	return recipes, nil
}

func (c *Client) GetRecipeInformationBulk(ctx context.Context, ids []int64) ([]recipeclient.RecipeWithIngredients, error) {
	var response []recipeInformation

	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = strconv.FormatInt(id, 10)
	}

	params := url.Values{}
	params.Set("ids", strings.Join(idStrings, ","))

	err := c.get(ctx, "/recipes/informationBulk", params, &response)
	if err != nil {
		return nil, err
	}

	result := make([]recipeclient.RecipeWithIngredients, len(response))
	for i, r := range response {
		ingredients := make([]recipeclient.Ingredient, len(r.ExtendedIngredients))
		for j, ing := range r.ExtendedIngredients {
			ingredients[j] = recipeclient.Ingredient{
				Name:   ing.Name,
				Amount: ing.Measures.Metric.Amount,
				Unit:   ing.Measures.Metric.UnitShort,
			}
		}
		result[i] = recipeclient.RecipeWithIngredients{
			ID:          r.ID,
			Ingredients: ingredients,
		}
	}

	return result, nil
}
