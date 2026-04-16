package main

import (
	"log"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/carloscfgos1980/taskSphere-api/internal/config"
)

func main() {
	// create a context
	ctx := context.Background()

	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Connect to the database using cfg.DB_URL
	conn, err := pgx.Connect(ctx, cfg.DB_URL)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer conn.Close(ctx)

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

	// Start the server on the specified port
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
