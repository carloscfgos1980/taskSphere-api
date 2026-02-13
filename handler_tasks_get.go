package main

import (
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerTasksGet(w http.ResponseWriter, r *http.Request) {
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
		ParentID    uuid.UUID   `json:"parent_id,omitempty"`
		TaskEditors []uuid.UUID `json:"task_editors"`
	}

	taskIDString := r.PathValue("taskID")
	taskID, err := uuid.Parse(taskIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

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
	if len(dbTaskEditors) == 0 {
		dbTaskEditors = []uuid.UUID{}
	}

	if dbTask.Tag == "private" {
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

		if dbTask.UserID != userID {
			respondWithError(w, http.StatusForbidden, "You don't have access to this task", nil)
			return
		}
	}

	respondWithJSON(w, http.StatusOK, response{
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
		TaskEditors: dbTaskEditors,
	})
}
