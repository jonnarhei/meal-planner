package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonnarhei/meal-planner/backend/internal/auth"
	"github.com/jonnarhei/meal-planner/backend/internal/store"
)

func newTestApp(t *testing.T, s store.Storage) *application {
	t.Helper()
	return &application{
		config: config{
			addr: ":8080",
			jwt: jwtConfig{
				secret: "test-secret",
				expiry: 86400,
			},
		},
		store: s,
	}
}

func executeRequest(req *http.Request, app *application) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.mount().ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected response code %d, got %d", expected, actual)
	}
}

func authHeaderFor(t *testing.T, app *application, userID int64) string {
	t.Helper()
	tok, err := auth.GenerateToken(userID, "test@test.com", app.config.jwt.secret, app.config.jwt.expiry)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + tok
}

func newJSONRequest(t *testing.T, method, path string, body any, authHeader string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, path, &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	return req
}

func validRecipeBody() map[string]any {
	return map[string]any{
		"title":    "Pancakes",
		"servings": 4,
		"ingredients": []map[string]any{
			{"name": "flour", "amount": 1.5, "unit": "cup"},
		},
	}
}
