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

type apiConfig struct {
	db        *database.Queries
	jwtSecret string
	port      string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer dbConn.Close()

	dbQueries := database.New(dbConn)

	apiCfg := apiConfig{
		db:        dbQueries,
		port:      port,
		jwtSecret: jwtSecret,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUsersUpdate)
	mux.HandleFunc("GET /api/users", apiCfg.handlerUsersRetrieve)

	mux.HandleFunc("POST /api/tasks", apiCfg.handlerTasksCreate)
	mux.HandleFunc("GET /api/tasks/{taskID}", apiCfg.handlerTasksGet)
	mux.HandleFunc("GET /api/personal/tasks", apiCfg.handlerTasksGetPersonal)
	mux.HandleFunc("GET /api/collaborative/tasks/{parentID}", apiCfg.handlerTasksGetCollaborative)
	mux.HandleFunc("PUT /api/tasks/{taskID}", apiCfg.handlerTasksUpdate)
	mux.HandleFunc("DELETE /api/tasks/{taskID}", apiCfg.handlerTasksDelete)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server is running http://localhost:%s", port)

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
