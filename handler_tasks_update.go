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
		Title       string    `json:"title,omitempty"`
		EndDate     time.Time `json:"end_date,omitempty"`
		Description string    `json:"description,omitempty"`
		Priority    string    `json:"priority,omitempty"`
		State       string    `json:"state,omitempty"`
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
	// Validate the parameters (e.g., check date format, priority and state values)
	err = CheckPriority(params.Priority)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	err = CheckState(params.State)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Update the task in the database with the new parameters
	updatedTask, err := cfg.db.UpdateTask(r.Context(), database.UpdateTaskParams{
		ID:          taskID,
		Title:       params.Title,
		EndDate:     params.EndDate,
		Description: params.Description,
		Priority:    params.Priority,
		State:       params.State,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update task", err)
		return
	}

	// Respond with the updated task details
	respondWithJSON(w, http.StatusOK, updatedTask)
}
