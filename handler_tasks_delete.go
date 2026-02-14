package main

import (
	"net/http"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerTasksDelete(w http.ResponseWriter, r *http.Request) {
	// Extract task ID from URL
	taskIDString := r.PathValue("taskID")
	taskID, err := uuid.Parse(taskIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}
	// Authenticate user
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	// Validate JWT and get user ID
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	// Check if task exists and belongs to user
	dbTask, err := cfg.db.GetTaskByID(r.Context(), taskID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get task", err)
		return
	}
	if dbTask.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You can't delete this task", err)
		return
	}
	// Delete task
	err = cfg.db.DeleteTask(r.Context(), taskID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete task", err)
		return
	}
	// Respond with no content
	w.WriteHeader(http.StatusNoContent)
}
