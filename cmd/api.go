package main

import (
	"log"
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/authmiddleware"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"
	"github.com/carloscfgos1980/taskSphere-api/internal/refresh"
	"github.com/carloscfgos1980/taskSphere-api/internal/tasks"
	"github.com/carloscfgos1980/taskSphere-api/internal/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
)

// application is the main application struct that holds the configuration and database connection
type application struct {
	config config
	db     *pgx.Conn
}

// config holds the configuration for the application
type config struct {
	addr      string
	db        dbConfig
	JWTSecret string
}

// dbConfig holds the database configuration for the application
type dbConfig struct {
	dsn string
}

// mount sets up the routes and middleware for the application
func (app *application) mount() http.Handler {
	// create a new router
	r := chi.NewRouter()
	// set up middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good for now"))
	})
	// users endpoints
	// create the user service and handler
	userService := users.NewService(database.New(app.db), app.db)
	userHandler := users.NewHandler(userService, app.config.JWTSecret)
	// set up the users routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", userHandler.CreateUser)
		r.Post("/login", userHandler.LoginUser)
	})

	// refresh token endpoint
	refreshService := refresh.NewService(database.New(app.db), app.db)
	refreshHandler := refresh.NewHandler(refreshService, app.config.JWTSecret)
	r.Post("/refresh", refreshHandler.RefreshToken)
	r.Post("/revoke", refreshHandler.RevokeRefreshToken)

	// protected routes
	r.Route("/api", func(r chi.Router) {
		// Add authentication middleware here if available
		r.Use(func(next http.Handler) http.Handler {
			return authmiddleware.AuthMiddleware(next, app.config.JWTSecret)
		})
		// create the task service and handler
		taskService := tasks.NewService(database.New(app.db), app.db)
		taskHandler := tasks.NewHandler(taskService, app.config.JWTSecret)
		r.Post("/tasks", taskHandler.CreateTask)
		r.Get("/tasks/{taskID}", taskHandler.GetTaskByID)

	})
	return r
}

// run starts the HTTP server
func (app *application) run(h http.Handler) error {
	// create the HTTP server
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Starting server on %s", app.config.addr)
	// start the server
	return srv.ListenAndServe()
}
