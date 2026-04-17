package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/carloscfgos1980/taskSphere-api/internal/config"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"
	"github.com/carloscfgos1980/taskSphere-api/internal/utils"
	"github.com/gin-gonic/gin"
)

// structs and handler for creating a new user in the system
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
}

// UserRequest is the struct for the request body when creating a new user
type UserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CreateUserHandler is the handler for creating a new user in the system
func CreateUserHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Bind the JSON request body to the UserRequest struct
		var req UserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Validate email format
		err := utils.IsValidEmail(req.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Validate the password strength
		err = utils.IsStrongPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Hash the password before storing it in the database
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Create the user in the database using the provided configuration and request data
		user, err := cfg.DB.CreateUser(c, database.CreateUserParams{
			Username: req.Username,
			Email:    req.Email,
			Password: hashedPassword,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Return the created user as a response, excluding the password
		response := User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Username:  user.Username,
			Email:     user.Email,
		}
		// Send the response back to the client with a 200 OK status
		c.JSON(http.StatusOK, gin.H{"user": response})
	}
}

// LoginRequest is the struct for the request body when logging in a user
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginUserHandler is the handler for logging in a user and generating a JWT token
func LoginUserHandler(cfg *config.Config) gin.HandlerFunc {
	// Define a response struct that includes the user information and the generated tokens
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Bind the JSON request body to the LoginRequest struct
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Validate email format
		if err := utils.IsValidEmail(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Retrieve the user from the database using the provided email
		user, err := cfg.DB.GetUserByEmail(c, req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		// Check if the provided password matches the stored hashed password
		match, err := utils.CheckPasswordHash(req.Password, user.Password)
		if err != nil || !match {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		// Generate a JWT token for the authenticated user
		token, err := utils.MakeJWT(
			user.ID,
			cfg.JWTSecret,
			24*7*time.Hour,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}
		// Generate a refresh token and store it in the database
		refreshToken := utils.MakeRefreshToken()
		_, err = cfg.DB.CreateRefreshToken(c, database.CreateRefreshTokenParams{
			UserID:    user.ID,
			Token:     refreshToken,
			ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
			return
		}
		// Return the user information along with the generated tokens in the response
		response := response{
			User: User{
				ID:        user.ID,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Username:  user.Username,
				Email:     user.Email,
			},
			Token:        token,
			RefreshToken: refreshToken,
		}
		// Send the response back to the client with a 200 OK status
		c.JSON(http.StatusOK, response)
	}
}
