package handlers

import (
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/config"
	"github.com/carloscfgos1980/taskSphere-api/internal/utils"
	"github.com/gin-gonic/gin"
)

// RefreshTokenHandler is the handler for refreshing JWT tokens using a valid refresh token
func RefreshTokenHandler(cfg *config.Config) gin.HandlerFunc {
	// Define a response struct to return the new access token in the response body
	type response struct {
		Token string `json:"token"`
	}
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Extract the refresh token from the Authorization header of the incoming request
		refreshToken, err := utils.GetBearerToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh token"})
			return
		}
		// Retrieve the user associated with the provided refresh token from the database
		user, err := cfg.DB.GetUserFromRefreshToken(c, refreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
		// Generate a new JWT token for the user to authenticate future requests
		accessToken, err := utils.MakeJWT(
			user.ID,
			cfg.JWTSecret,
			24*7*time.Hour,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
			return
		}
		// Return the new JWT token in the response to the client
		c.JSON(http.StatusOK, response{Token: accessToken})

	}
}

// RevokeRefreshTokenHandler is the handler for revoking refresh tokens
func RevokeRefreshTokenHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Extract the refresh token from the Authorization header of the incoming request
		refreshToken, err := utils.GetBearerToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh token"})
			return
		}
		// Revoke the provided refresh token in the database to prevent further use
		_, err = cfg.DB.RevokeRefreshToken(c, refreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke refresh token"})
			return
		}
		// Return a success message in the response to the client
		c.JSON(http.StatusOK, gin.H{"message": "refresh token revoked successfully"})
	}
}
