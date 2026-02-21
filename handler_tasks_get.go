package main

import (
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerTasksGet(w http.ResponseWriter, r *http.Request) {
	// Extract the task ID from the URL path and validate it
	taskIDString := r.PathValue("taskID")
	taskID, err := uuid.Parse(taskIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}
	// Retrieve the task from the database using the provided ID
	dbTask, err := cfg.db.GetTaskByID(r.Context(), taskID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get task", err)
		return
	}
	// If the task is marked as "private", validate the user's authorization to access it
	if dbTask.Tag == "private" {
		// Extract the Bearer token from the Authorization header
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
			return
		}
		// Verify JWT token and extract the user ID
		userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
			return
		}
		// Check if the user ID from the token matches the user ID associated with the task
		if dbTask.UserID != userID {
			respondWithError(w, http.StatusForbidden, "You don't have access to this task", nil)
			return
		}
	}
	// Respond with the task details in JSON format
	respondWithJSON(w, http.StatusOK, dbTask)
}

func (cfg *apiConfig) handlerTasksGetPersonal(w http.ResponseWriter, r *http.Request) {
	// Extract the Bearer token from the Authorization header and validate it to get the user ID
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
	// Retrieve the tasks associated with the user ID from the database
	dbTasks, err := cfg.db.GetTasksByUserID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get tasks", err)
		return
	}

	// Iterate through the retrieved tasks, get their editors, and build the response array
	var response []Task

	for _, dbTask := range dbTasks {
		response = append(response, Task{
			ID:          dbTask.ID,
			CreatedAt:   dbTask.CreatedAt,
			UpdatedAt:   dbTask.UpdatedAt,
			UserID:      dbTask.UserID,
			Title:       dbTask.Title,
			EndDate:     dbTask.EndDate,
			Description: dbTask.Description,
			Priority:    dbTask.Priority,
			Tag:         dbTask.Tag,
			State:       dbTask.State,
			ParentID:    dbTask.ParentID.UUID,
			TaskEditors: dbTask.TaskEditors,
		})
	}
	// Respond with the array of tasks in JSON format
	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) handlerTasksGetCollaborative(w http.ResponseWriter, r *http.Request) {
	// Extract the parent task ID from the URL path and validate it
	parentIDString := r.PathValue("parentID")
	ParentID, err := uuid.Parse(parentIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}
	// Retrieve the collaborative tasks associated with the parent task ID from the database

	dbGroupTasks, err := cfg.db.GetCollaborativeTasksByParentID(r.Context(), uuid.NullUUID{UUID: ParentID, Valid: true})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get collaborative tasks", err)
		return
	}
	// Extract the Bearer token from the Authorization header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
		return
	}
	// Verify JWT token and extract the user ID
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	// Define the response structure for a single collaborative task
	type taskResponse struct {
		ID          uuid.UUID   `json:"id"`
		CreatedAt   time.Time   `json:"created_at"`
		UpdatedAt   time.Time   `json:"updated_at"`
		UserID      uuid.UUID   `json:"user_id"`
		Username    string      `json:"username"`
		Email       string      `json:"email"`
		Title       string      `json:"title"`
		EndDate     time.Time   `json:"end_date"`
		Description string      `json:"description"`
		Priority    string      `json:"priority"`
		Tag         string      `json:"tag"`
		State       string      `json:"state"`
		ParentID    uuid.UUID   `json:"parent_id,omitempty"`
		TaskEditors []uuid.UUID `json:"task_editors"`
	}
	// Iterate through the retrieved collaborative tasks, check if the user is part of the work group, and build the response array
	isAuthorized := false
	var response []taskResponse
	for _, dbGroupTask := range dbGroupTasks {
		if dbGroupTask.UserID == userID {
			isAuthorized = true
		}
		// Append the collaborative task details along with its editors to the response array
		response = append(response, taskResponse{
			ID:          dbGroupTask.ID,
			CreatedAt:   dbGroupTask.CreatedAt,
			UpdatedAt:   dbGroupTask.UpdatedAt,
			UserID:      dbGroupTask.UserID,
			Username:    dbGroupTask.Username,
			Email:       dbGroupTask.Email,
			Title:       dbGroupTask.Title,
			EndDate:     dbGroupTask.EndDate,
			Description: dbGroupTask.Description,
			Priority:    dbGroupTask.Priority,
			Tag:         dbGroupTask.Tag,
			State:       dbGroupTask.State,
			ParentID:    dbGroupTask.ParentID.UUID,
			TaskEditors: dbGroupTask.TaskEditors,
		})
	}
	// If the user is not authorized, respond with an unauthorized error
	if !isAuthorized {
		respondWithError(w, http.StatusUnauthorized, "you are not part of this work group", nil)
		return
	}
	// Respond with the array of collaborative tasks in JSON format
	respondWithJSON(w, http.StatusOK, response)

}
