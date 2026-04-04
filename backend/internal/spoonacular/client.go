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

func (c *Client) GetRandomRecipes(ctx context.Context, n int, preferences []string) (*RandomRecipeResponse, error) {
	var response RandomRecipeResponse

	params := url.Values{}
	params.Set("number", strconv.Itoa(n))

	tags := append([]string{"main course"}, preferences...)
	params.Set("include-tags", strings.Join(tags, ","))

	err := c.get(ctx, "/recipes/random", params, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type Measures struct {
	Amount   float64 `json:"amount"`
	UnitLong string  `json:"unitLong"`
}

type MetricMeasures struct {
	Metric Measures `json:"metric"`
}

type Ingredient struct {
	Name     string         `json:"name"`
	Measures MetricMeasures `json:"Measures"`
}

type RecipeInformation struct {
	ID                  int64        `json:"id"`
	ExtendedIngredients []Ingredient `json:"extendedIngredients"`
}

func (c *Client) GetIngredientsForRecipes(ctx context.Context, recipeIDs []int64) ([]RecipeInformation, error) {
	var response []RecipeInformation

	idStrings := make([]string, len(recipeIDs))
	for index, id := range recipeIDs {
		idStrings[index] = strconv.FormatInt(id, 10)
	}

	params := url.Values{}
	params.Set("ids", strings.Join(idStrings, ","))

	err := c.get(ctx, "/recipes/informationBulk", params, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
