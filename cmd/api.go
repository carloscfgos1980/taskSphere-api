package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/authmiddleware"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"
	jsonresponse "github.com/carloscfgos1980/taskSphere-api/internal/json"
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

	startedAt    time.Time
	requestStats *requestMetrics
	initOnce     sync.Once
}

type requestMetrics struct {
	totalRequests      atomic.Uint64
	totalDurationNanos atomic.Int64
	lastDurationNanos  atomic.Int64
	status2xx          atomic.Uint64
	status4xx          atomic.Uint64
	status5xx          atomic.Uint64
}

type metricsSnapshot struct {
	totalRequests uint64
	averageMS     float64
	lastMS        float64
	status2xx     uint64
	status4xx     uint64
	status5xx     uint64
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
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
	app.initRuntimeMetrics()

	// create a new router
	r := chi.NewRouter()
	// set up middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(app.metricsMiddleware)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good for now"))
	})
	r.Get("/metrics", app.Metrics)
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
		// set up the tasks routes
		r.Post("/tasks", taskHandler.CreateTask)
		r.Get("/tasks/{taskID}", taskHandler.GetTaskByID)
		r.Get("/tasks", taskHandler.GetTasks)
		r.Get("/tasks/collaborative", taskHandler.GetParentsCollaborativeTasks)
		r.Put("/tasks/{taskID}", taskHandler.UpdateTask)
		r.Delete("/tasks/{taskID}", taskHandler.DeleteTask)

	})
	return r
}

func (app *application) initRuntimeMetrics() {
	app.initOnce.Do(func() {
		app.startedAt = time.Now().UTC()
		app.requestStats = &requestMetrics{}
	})
}

func (app *application) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(recorder, r)

		elapsed := time.Since(started)
		app.recordRequestMetrics(recorder.statusCode, elapsed)
	})
}

func (app *application) recordRequestMetrics(statusCode int, elapsed time.Duration) {
	if app.requestStats == nil {
		return
	}

	app.requestStats.totalRequests.Add(1)
	app.requestStats.totalDurationNanos.Add(elapsed.Nanoseconds())
	app.requestStats.lastDurationNanos.Store(elapsed.Nanoseconds())

	switch {
	case statusCode >= 500:
		app.requestStats.status5xx.Add(1)
	case statusCode >= 400:
		app.requestStats.status4xx.Add(1)
	default:
		app.requestStats.status2xx.Add(1)
	}
}

func (app *application) snapshotMetrics() metricsSnapshot {
	if app.requestStats == nil {
		return metricsSnapshot{}
	}

	total := app.requestStats.totalRequests.Load()
	totalNanos := app.requestStats.totalDurationNanos.Load()
	lastNanos := app.requestStats.lastDurationNanos.Load()

	avgMS := 0.0
	if total > 0 {
		avgMS = float64(totalNanos) / float64(total) / float64(time.Millisecond)
	}

	return metricsSnapshot{
		totalRequests: total,
		averageMS:     avgMS,
		lastMS:        float64(lastNanos) / float64(time.Millisecond),
		status2xx:     app.requestStats.status2xx.Load(),
		status4xx:     app.requestStats.status4xx.Load(),
		status5xx:     app.requestStats.status5xx.Load(),
	}
}

func (app *application) getUserCount(r *http.Request) (int64, error) {
	var userCount int64
	err := app.db.QueryRow(r.Context(), "SELECT COUNT(*) FROM users").Scan(&userCount)
	return userCount, err
}

func (app *application) getTaskCount(r *http.Request) (int64, error) {
	var taskCount int64
	err := app.db.QueryRow(r.Context(), "SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	return taskCount, err
}

func (app *application) Metrics(w http.ResponseWriter, r *http.Request) {
	snapshot := app.snapshotMetrics()
	uptime := time.Since(app.startedAt)

	userCount, userErr := app.getUserCount(r)
	taskCount, taskErr := app.getTaskCount(r)

	response := map[string]any{
		"service":        "taskSphere-api",
		"uptime_seconds": uptime.Seconds(),
		"api_response_times_ms": map[string]any{
			"average": snapshot.averageMS,
			"last":    snapshot.lastMS,
		},
		"requests": map[string]any{
			"total":      snapshot.totalRequests,
			"status_2xx": snapshot.status2xx,
			"status_4xx": snapshot.status4xx,
			"status_5xx": snapshot.status5xx,
		},
		"counts": map[string]any{
			"users": userCount,
			"tasks": taskCount,
		},
		"deployment": map[string]any{
			"bind_address": app.config.addr,
			"app_env":      os.Getenv("APP_ENV"),
			"go_version":   runtime.Version(),
			"started_at":   app.startedAt.Format(time.RFC3339),
		},
		"architecture": map[string]any{
			"router":            "chi",
			"auth":              "JWT access token + refresh token",
			"database":          "PostgreSQL via pgx + sqlc",
			"service_structure": "handler -> service -> database",
		},
	}

	if userErr != nil || taskErr != nil {
		response["metrics_warnings"] = map[string]any{
			"user_count_error": userErr,
			"task_count_error": taskErr,
		}
	}

	if err := jsonresponse.WriteJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "unable to encode metrics response", http.StatusInternalServerError)
	}
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
