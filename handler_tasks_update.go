package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerTasksUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title       string    `json:"title,omitempty"`
		EndDate     time.Time `json:"end_date,omitempty"`
		Description string    `json:"description,omitempty"`
		Priority    string    `json:"priority,omitempty"`
		State       string    `json:"state,omitempty"`
	}

	taskIDString := r.PathValue("taskID")
	taskID, err := uuid.Parse(taskIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
		return
	}

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

	dbTaskEditors, err := cfg.db.GetTaskEditorsByTaskID(r.Context(), dbTask.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve task editors", err)
		return
	}
	// Check if the user is authorized to update the task (either the owner or a task editor)
	isAuthorized := dbTask.UserID == userID
	if !isAuthorized {
		for _, editorID := range dbTaskEditors {
			if editorID == userID {
				isAuthorized = true
				break
			}
		}
	}

	if !isAuthorized {
		respondWithError(w, http.StatusForbidden, "You don't have permission to update this task", nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	errDate := CheckDateFormat(params.EndDate)
	if errDate != "" {
		respondWithError(w, http.StatusBadRequest, errDate, nil)
		return
	}
	errPriority := CheckPriority(params.Priority)
	if errPriority != "" {
		respondWithError(w, http.StatusBadRequest, errPriority, nil)
		return
	}
	errState := CheckState(params.State)
	if errState != "" {
		respondWithError(w, http.StatusBadRequest, errState, nil)
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

	type response struct {
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
		EditorIDs   []uuid.UUID `json:"editor_ids"`
	}

	respondWithJSON(w, http.StatusOK, response{
		ID:          updatedTask.ID,
		CreatedAt:   updatedTask.CreatedAt,
		UpdatedAt:   updatedTask.UpdatedAt,
		UserID:      updatedTask.UserID,
		Title:       updatedTask.Title,
		EndDate:     updatedTask.EndDate,
		Description: updatedTask.Description,
		Priority:    updatedTask.Priority,
		Tag:         updatedTask.Tag,
		State:       updatedTask.State,
		EditorIDs:   dbTaskEditors,
	})
}
