package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

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