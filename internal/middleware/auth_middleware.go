package middleware

import (
	"fmt"
	"net/http"

	"github.com/carloscfgos1980/taskSphere-api/internal/config"
	"github.com/carloscfgos1980/taskSphere-api/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a Gin middleware function that validates JWT tokens in incoming requests to protect routes that require authentication. It checks for the presence of a valid JWT token in the Authorization header of the request, verifies the token using the secret key from the configuration, and sets the user ID in the Gin context for use in subsequent handlers if the token is valid. If the token is missing or invalid, it returns a 401 Unauthorized response and aborts further processing of the request.
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router as middleware for routes that require authentication.
	return func(c *gin.Context) {
		// Extract the Authorization header from the incoming request
		token, err := utils.GetBearerToken(c)
		// If there is an error extracting the token (e.g., missing or malformed header), return a 401 Unauthorized response with an appropriate error message and abort the request processing.
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Unauthorized: %v", err)})
			c.Abort()
			return
		}
		// Validate the extracted token using the secret key from the configuration. If the token is invalid (e.g., expired, malformed, or signature mismatch), return a 401 Unauthorized response with an appropriate error message and abort the request processing.
		userID, err := utils.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			c.Abort()
			return
		}
		// If the token is valid, set the user ID in the Gin context (e.g., using c.Set("userID", userID)) for use in subsequent handlers that require authentication.
		c.Set("userID", userID)
		// Call the next handler in the chain to continue processing the request after successful authentication.
		c.Next()
	}
}
