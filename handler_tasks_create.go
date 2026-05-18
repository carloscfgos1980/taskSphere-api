package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"

	"github.com/google/uuid"
)

// Task represents the structure of a task in the system, including its ID, timestamps, user association, title, end date, description, priority, tag, state, parent task association, and any associated editors.
type Task struct {
	ID          uuid.UUID   `json:"id"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	UserID      uuid.UUID   `json:"user_id"`
	Title       string      `json:"title"`
	EndDate     time.Time   `json:"end_date"`
	Description string      `json:"description"`
	Priority    string      `json:"priority"`
	Tag         string      `json:"tag"`
	State       string      `json:"state"`
	ParentID    uuid.UUID   `json:"parent_id"`
	TaskEditors []uuid.UUID `json:"task_editors"`
}

func (cfg *apiConfig) handlerTasksCreate(w http.ResponseWriter, r *http.Request) {
	// Define the expected parameters for creating a new task and the response structure
	type parameters struct {
		Title       string      `json:"title" binding:"required"`
		EndDate     time.Time   `json:"end_date" binding:"required"`
		Description string      `json:"description" binding:"required"`
		Priority    string      `json:"priority"`
		Tag         string      `json:"tag"`
		State       string      `json:"state"`
		ParentID    uuid.UUID   `json:"parent_id,omitempty"`
		TaskEditors []uuid.UUID `json:"task_editors"`
	}

	// Validate the user's authorization to create a new task by checking the provided JWT token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
		return
	}
	// Validate the JWT token and extract the user ID from it
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization token", err)
		return
	}
	// Decode the JSON request body into the parameters struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// Validate the provided parameters for creating a new task (e.g., check if priority, state, tag, and date formats are valid)
	if params.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required", fmt.Errorf("title is required"))
		return
	}
	if params.EndDate.IsZero() {
		respondWithError(w, http.StatusBadRequest, "End date is required", fmt.Errorf("end date is required"))
		return
	}
	if params.Description == "" {
		respondWithError(w, http.StatusBadRequest, "Description is required", fmt.Errorf("description is required"))
		return
	}
	priority, err := auth.CheckPriority(params.Priority)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	state, err := auth.CheckState(params.State)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	tag, err := auth.CheckTag(params.Tag)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	log.Printf("Creating task for user %s with title %s", userID, params.Title)
	// Create a new task in the database using the provided parameters and the user ID extracted from the JWT token
	dbTask, err := cfg.db.CreateTask(r.Context(), database.CreateTaskParams{
		UserID:      userID,
		Title:       params.Title,
		EndDate:     params.EndDate,
		Description: params.Description,
		Priority:    priority,
		Tag:         tag,
		State:       state,
		ParentID:    uuid.NullUUID{UUID: params.ParentID, Valid: params.ParentID != uuid.Nil},
		TaskEditors: params.TaskEditors,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create task", err)
		return
	}

	// Respond with the created task's information, including its editors
	respondWithJSON(w, http.StatusCreated, dbTask)
}
