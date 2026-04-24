package users

import (
	"time"

	"github.com/google/uuid"
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
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the struct for the request body when logging in a user
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
