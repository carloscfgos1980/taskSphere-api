package main

import (
	"net/http"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"
)

func (cfg *apiConfig) handlerUsersRetrieve(w http.ResponseWriter, r *http.Request) {
	// Only authenticated users can retrieve the list of users
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	// Validate the JWT and extract the user ID
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	// Ensure the user exists in the database
	_, err = cfg.db.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find user", err)
		return
	}
	// Retrieve users from the database
	dbUsers, err := cfg.db.GetUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}
	// Map database users to API users
	users := []User{}
	for _, dbUser := range dbUsers {

		users = append(users, User{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Username:  dbUser.Username,
			Email:     dbUser.Email,
		})
	}
	// Respond with the list of users
	respondWithJSON(w, http.StatusOK, users)
}
