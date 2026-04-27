package refresh

import (
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/json"
	"github.com/carloscfgos1980/taskSphere-api/internal/utils"
)

// handler is the HTTP handler for users endpoints
type handler struct {
	service   Service
	jwtSecret string
}

// NewHandler creates a new handler for users endpoints
func NewHandler(service Service, jwtSecret string) *handler {
	return &handler{
		service:   service,
		jwtSecret: jwtSecret,
	}
}

// RefreshToken handles the HTTP request for refreshing a user's access token
func (h *handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get the refresh token from the Authorization header
	refreshToken, err := utils.GetBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// Get the user associated with the refresh token from the database
	user, err := h.service.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}
	// Generate a new JWT token for the user to authenticate future requests
	accessToken, err := utils.MakeJWT(
		user.ID.Bytes,
		h.jwtSecret,
		24*7*time.Hour,
	)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Create a response struct to send back to the client with the new access token and refresh token
	type response struct {
		Token string `json:"token"`
	}
	resp := response{
		Token: accessToken,
	}
	// Write the response as JSON with a 200 OK status code
	if err := json.WriteJSON(w, http.StatusOK, resp); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// RevokeRefreshToken handles the HTTP request for revoking a user's refresh token
func (h *handler) RevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get the refresh token from the Authorization header
	refreshToken, err := utils.GetBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// Revoke the refresh token in the database
	err = h.service.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		http.Error(w, "Failed to revoke refresh token", http.StatusInternalServerError)
		return
	}
	// Write a success response with a 200 OK status code
	if err := json.WriteJSON(w, http.StatusOK, map[string]string{"message": "Refresh token revoked successfully"}); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
