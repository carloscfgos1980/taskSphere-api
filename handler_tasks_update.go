package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerTasksUpdate(w http.ResponseWriter, r *http.Request) {
	// Define a struct to hold the parameters for updating the task
	type parameters struct {
		Title       *string    `json:"title,omitempty"`
		EndDate     *time.Time `json:"end_date,omitempty"`
		Description *string    `json:"description,omitempty"`
		Priority    *string    `json:"priority,omitempty"`
		State       *string    `json:"state,omitempty"`
	}
	// Extract the task ID from the URL path
	taskIDString := r.PathValue("taskID")
	taskID, err := uuid.Parse(taskIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}
	// Get the user ID from the JWT token in the Authorization header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
		return
	}
	// Validate the JWT token and extract the user ID
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	// Retrieve the task from the database using the provided ID
	dbTask, err := cfg.db.GetTaskByID(r.Context(), taskID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get task", err)
		return
	}

	// Check if the user is authorized to update the task (either the owner or a task editor)
	isAuthorized := dbTask.UserID == userID
	if !isAuthorized {
		for _, editorID := range dbTask.TaskEditors {
			if editorID == userID {
				isAuthorized = true
				break
			}
		}
	}
	// If the user is not authorized, respond with a 403 Forbidden error
	if !isAuthorized {
		respondWithError(w, http.StatusForbidden, "You don't have permission to update this task", errors.New("user not authorized to update task"))
		return
	}
	// Decode the request body to get the new task parameters
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	title := dbTask.Title
	if params.Title != nil {
		title = *params.Title
	}
	endDate := dbTask.EndDate
	if params.EndDate != nil {
		endDate = *params.EndDate
	}
	description := dbTask.Description
	if params.Description != nil {
		description = *params.Description
	}
	priority := dbTask.Priority
	if params.Priority != nil {
		priority = *params.Priority
	}
	state := dbTask.State
	if params.State != nil {
		state = *params.State
	}

	// Validate the parameters (e.g., check date format, priority and state values)

	updatedTask, err := cfg.db.UpdateTask(r.Context(), database.UpdateTaskParams{
		ID:          taskID,
		Title:       title,
		EndDate:     endDate,
		Description: description,
		Priority:    priority,
		State:       state,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update task", err)
		return
	}

	// Respond with the updated task details
	respondWithJSON(w, http.StatusOK, updatedTask)
}
