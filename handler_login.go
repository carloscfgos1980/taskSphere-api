package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Define the expected parameters for user login and the response structure
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	// Define the response structure for a successful login, including the user's information and the generated JWT token and refresh token
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	// Decode the JSON request body into the parameters struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// Retrieve the user from the database using the provided email address
	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	// Check if the provided password matches the hashed password stored in the database for the retrieved user
	match, err := auth.CheckPasswordHash(params.Password, user.Password)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	// If the password is correct, generate a JWT token for the user to authenticate future requests
	token, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		24*7*time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT token", err)
		return
	}
	// Generate a refresh token for the user to allow them to obtain a new JWT token without re-entering their credentials
	refreshToken := auth.MakeRefreshToken()
	// Store the generated refresh token in the database associated with the user's ID and an expiration time
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}
	// Respond with the user's information (excluding the password) and the generated JWT token and refresh token
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}
