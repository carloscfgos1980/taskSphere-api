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

// CreateUser is the handler for creating a new user in the system
func CreateUser(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Bind the JSON request body to the UserRequest struct
		var req UserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Validate the password strength
		err := utils.IsStrongPassword(req.Password)
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
