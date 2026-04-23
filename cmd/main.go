package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/carloscfgos1980/taskSphere-api/internal/env"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	godotenv.Load()
	// Get the port from environment variables, default to 8080 if not set
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}
	// create a context
	ctx := context.Background()
	// load env variables
	cfg := config{
		addr: ":" + port,
		db: dbConfig{
			dsn: env.GetEnv("DB_URL", "postgres://carlosinfante:@localhost:5432/tasksphere?sslmode=disable"),
		},
	}
	// initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// database connection
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		slog.Error("unable to connect to database", "err", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	logger.Info("successfully connected to database", "dsn", cfg.db.dsn)
	// create the application
	api := &application{
		config: cfg,
		db:     conn,
	}
	// run the application
	if err := api.run(api.mount()); err != nil {
		slog.Error("server has failed to start", "err", err)
		os.Exit(1)
	}
}
