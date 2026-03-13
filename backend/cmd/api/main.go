package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/jonnarhei/meal-planner/backend/internal/database"
	"github.com/jonnarhei/meal-planner/backend/internal/env"
	"github.com/jonnarhei/meal-planner/backend/internal/store"
)

func main() {

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://user:adminpassword@localhost:5432/mealplanner?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}

	db, err := database.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("Database connection pool established")
  
	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  *store,
	}

	
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mux := app.mount()
	if err := app.run(mux); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
