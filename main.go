package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/carloscfgos1980/taskSphere-api/internal/database"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

// apiConfig holds the dependencies for the API handlers.
type apiConfig struct {
	db        *database.Queries
	jwtSecret string
	port      string
}

func main() {
	// Load environment variables from .env file
	godotenv.Load()
	// Get configuration from environment variables
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	// Get the port from environment variables, default to 8080 if not set
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}
	// Get the JWT secret from environment variables
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}
	// Connect to the database
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer dbConn.Close()

	// database queries variable
	dbQueries := database.New(dbConn)
	// variable for the apiConfig struct
	apiCfg := apiConfig{
		db:        dbQueries,
		port:      port,
		jwtSecret: jwtSecret,
	}
	// Set up the HTTP server and routes
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUsersUpdate)
	mux.HandleFunc("GET /api/users", apiCfg.handlerUsersRetrieve)

	mux.HandleFunc("POST /v1/api/tasks", apiCfg.handlerTasksCreate)
	mux.HandleFunc("GET /v1/api/tasks/{taskID}", apiCfg.handlerTasksGet)
	mux.HandleFunc("GET /v1/api/personal/tasks", apiCfg.handlerTasksGetPersonal)
	mux.HandleFunc("GET /v1/api/collaborative/tasks/{parentID}", apiCfg.handlerTasksGetCollaborative)
	mux.HandleFunc("PUT /v1/api/tasks/{taskID}", apiCfg.handlerTasksUpdate)
	mux.HandleFunc("DELETE /v1/api/tasks/{taskID}", apiCfg.handlerTasksDelete)

	// Start the HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server is running http://localhost:%s", port)
	// Listen and serve
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
