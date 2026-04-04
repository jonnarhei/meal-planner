package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jonnarhei/meal-planner/backend/internal/env"
	"github.com/jonnarhei/meal-planner/backend/internal/spoonacular"
	"github.com/jonnarhei/meal-planner/backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

type application struct {
	config      config
	store       store.Storage
	spoonacular spoonacular.Client
}

type config struct {
	addr string
	db   dbConfig
	jwt  jwtConfig
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type jwtConfig struct {
	secret string
	expiry int
}

func (app *application) mount() http.Handler {
	allowedOrigins := env.GetString("ALLOWED_ORIGINS", "http://localhost:5173")
	
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{allowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	//standard middleware stack from chi
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(httprate.LimitByIP(60, time.Minute))

	router.Use(middleware.Timeout(60 * time.Second))

	router.Get("/health", app.healthCheckHandler)

	router.Route("/users", func(r chi.Router) {
		r.With(httprate.LimitByIP(10, time.Minute)).Post("/", app.registerUserHandler)
		r.With(httprate.LimitByIP(10, time.Minute)).Post("/login", app.loginUserHandler)

		r.Group(func(r chi.Router) {
			r.Use(app.AuthMiddleware)
			r.Route("/me", func(r chi.Router) {
				r.Get("/", app.getMeHandler)
				r.Put("/preferences", app.updateDietaryPreferences)
			})
		})
	})

	router.Route("/meal-plans", func(r chi.Router) {
		r.Use(app.AuthMiddleware)
		r.Get("/current", app.getCurrentMealPlanHandler)
		r.Patch("/current/recipe", app.changeRecipeForDay)
		r.Post("/current/regenerate", app.regenerateMealPlanHandler)
	})

	router.Route("/shopping-list", func(r chi.Router) {
		r.Use(app.AuthMiddleware)
		r.Get("/", app.getShoppingListHandler)
		r.Post("/items", app.addShoppingListItemsHandler)
		r.Post("/from-meal-plan", app.addFromMealPlanHandler)
		r.Patch("/items/{id}", app.toggleCheckedHandler)
		r.Delete("/items/{id}", app.deleteItemHandler)
		r.Delete("/checked", app.deleteCheckedHandler)
	})

	return router
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Minute,
	}

	slog.Info("Server has started at", "addr", srv.Addr)

	return srv.ListenAndServe()
}
