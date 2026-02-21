package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"
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

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	// Define the expected parameters for creating a new user and the response structure
	type parameters struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	// Define the response structure for a single user
	type response struct {
		User
	}
	// Decode the JSON request body into the parameters struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// strong password validation can be added here before hashing the password and creating the user in the database
	err = auth.IsStrongPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Hash the user's password before storing it in the database
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}
	// Create a new user in the database using the provided parameters and the hashed password
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Username: params.Username,
		Email:    params.Email,
		Password: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}
	// Respond with the created user's information (excluding the password)
	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Username:  user.Username,
			Email:     user.Email,
		},
	})
}
