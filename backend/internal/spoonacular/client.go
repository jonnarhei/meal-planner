package spoonacular

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL	string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		baseURL: "https://api.spoonacular.com",
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