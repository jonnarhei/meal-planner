package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/jonnarhei/meal-planner/backend/internal/database"
	"github.com/jonnarhei/meal-planner/backend/internal/env"
	"github.com/jonnarhei/meal-planner/backend/internal/spoonacular"
	"github.com/jonnarhei/meal-planner/backend/internal/store"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://user:adminpassword@localhost:5432/mealplanner?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		jwt: jwtConfig{
			secret: env.GetString("JWT_SECRET", "changeme"),
			expiry: env.GetInt("JWT_EXPIRY", 86400),
		},
	}

	db, err := database.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Database connection pool established")

	store := store.NewStorage(db)

	spoonacularClient := spoonacular.NewClient(env.GetString("SPOONACULAR_API_KEY", ""))

	app := &application{
		config:      cfg,
		store:       *store,
		spoonacular: *spoonacularClient,
	}

	mux := app.mount()
	if err := app.run(mux); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
