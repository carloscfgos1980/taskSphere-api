package main

import (
	"database/sql"
	"log"

	"github.com/carloscfgos1980/taskSphere-api/internal/database"
	"github.com/carloscfgos1980/taskSphere-api/internal/handlers"
	"github.com/carloscfgos1980/taskSphere-api/internal/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/carloscfgos1980/taskSphere-api/internal/config"
)

func main() {
	// create a context

	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Connect to the database
	dbConn, err := sql.Open("postgres", cfg.DB_URL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer dbConn.Close()

	// Create a new database queries instance
	db := database.New(dbConn)

	cfg.DB = db

	// Initialize the Gin router
	var router *gin.Engine = gin.Default()

	// Set trusted proxies to nil to avoid warnings in Gin 1.7+
	router.SetTrustedProxies(nil)

	// Define a simple health check route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":  "Todo API is running",
			"status":   "success",
			"database": "connected",
		})
	})
	// Register user-related routes
	router.POST("/auth/register", handlers.CreateUserHandler(cfg))
	router.POST("/auth/login", handlers.LoginUserHandler(cfg))

	// Register token-related routes
	router.POST("/auth/refresh", handlers.RefreshTokenHandler(cfg))
	router.POST("/auth/revoke", handlers.RevokeRefreshTokenHandler(cfg))

	// Register task-related routes
	taskRoutes := router.Group("/tasks")
	taskRoutes.Use(middleware.AuthMiddleware(cfg))
	taskRoutes.POST("/", handlers.CreateTaskHandler(cfg))
	taskRoutes.GET("/:taskID", handlers.GetTasksByIdHandler(cfg))

	// Start the server on the specified port
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
