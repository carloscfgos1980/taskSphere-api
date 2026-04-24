package users

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/carloscfgos1980/taskSphere-api/internal/json"
	"github.com/carloscfgos1980/taskSphere-api/internal/utils"
	"github.com/jackc/pgx/v5/pgconn"
)

// handler is the HTTP handler for users endpoints
type handler struct {
	service Service
}

// NewHandler creates a new handler for users endpoints
func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

// CreateUser handles the HTTP request for creating a new user
func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body into a UserRequest struct
	var userReq UserRequest
	if err := json.ReadJSON(r, &userReq); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Check if any field is empty
	if userReq.Username == "" || userReq.Email == "" || userReq.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}
	// Validate email format
	err := utils.IsValidEmail(userReq.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Validate the password strength
	err = utils.IsStrongPassword(userReq.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Hash the password before storing it in the database
	hashedPassword, err := utils.HashPassword(userReq.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Update the user request with the hashed password
	userReq.Password = hashedPassword

	// Call the service to create the user
	user, err := h.service.CreateUser(r.Context(), userReq)
	if err != nil {
		log.Println(err)
		// Check if the error is a unique constraint violation (duplicate email)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// PostgreSQL unique violation error code
			if strings.Contains(pgErr.Message, "email") {
				http.Error(w, "Email already exists", http.StatusConflict)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create a response struct to send back to the client, excluding the password
	response := User{
		ID:        user.ID.Bytes,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Username:  user.Username,
		Email:     user.Email,
	}
	// Write the response as JSON with a 201 Created status code
	if err := json.WriteJSON(w, http.StatusCreated, response); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
