package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jonnarhei/meal-planner/backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	config config
	store  store.Storage
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
	router := chi.NewRouter()

	//standard middleware stack from chi
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(middleware.Timeout(60 * time.Second))

	router.Get("/health", app.healthCheckHandler)

	router.Route("/users", func(r chi.Router) {
		r.Post("/", app.registerUserHandler)
		r.Post("/login", app.loginUserHandler)
		
		r.Group(func(r chi.Router) {
			r.Use(app.AuthMiddleware)
			r.Get("/", app.ListUsersHandler)
			r.Get("/me", app.getMeHandler)
		})
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
