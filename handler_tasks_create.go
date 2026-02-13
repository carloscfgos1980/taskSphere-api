package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"

	"github.com/google/uuid"
)

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
	ParentID    uuid.UUID   `json:"parent_id,omitempty"`
	TaskEditors []uuid.UUID `json:"task_editors"`
}

func (cfg *apiConfig) handlerTasksCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title       string      `json:"title"`
		EndDate     time.Time   `json:"end_date"`
		Description string      `json:"description"`
		Priority    string      `json:"priority"`
		Tag         string      `json:"tag"`
		State       string      `json:"state"`
		ParentID    uuid.UUID   `json:"parent_id,omitempty"`
		TaskEditors []uuid.UUID `json:"task_editors"`
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
		ParentID    uuid.UUID   `json:"parent_id,omitempty"`
		TaskEditors []uuid.UUID `json:"task_editors"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization token", err)
		return
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	if err := checkPriority(params.Priority); err != "" {
		respondWithError(w, http.StatusBadRequest, err, nil)
		return
	}
	if err := checkState(params.State); err != "" {
		respondWithError(w, http.StatusBadRequest, err, nil)
		return
	}
	if err := checkTag(params.Tag); err != "" {
		respondWithError(w, http.StatusBadRequest, err, nil)
		return
	}
	if err := checkDateFormat(params.EndDate); err != "" {
		respondWithError(w, http.StatusBadRequest, err, nil)
		return
	}

	log.Printf("Creating task for user %s with title %s", userID, params.Title)
	task, err := cfg.db.CreateTask(r.Context(), database.CreateTaskParams{
		UserID:      userID,
		Title:       params.Title,
		EndDate:     params.EndDate,
		Description: params.Description,
		Priority:    params.Priority,
		Tag:         params.Tag,
		State:       params.State,
		ParentID:    uuid.NullUUID{UUID: params.ParentID, Valid: params.ParentID != uuid.Nil},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create task", err)
		return
	}
	if len(params.TaskEditors) > 0 {
		log.Printf("Adding %d editors to task", len(params.TaskEditors))
		for _, editorID := range params.TaskEditors {
			_, err = cfg.db.CreateTaskEditors(r.Context(), database.CreateTaskEditorsParams{
				TaskID:   task.ID,
				EditorID: editorID,
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't create task editors", err)
				return
			}
		}
	}
	taskEditors, err := cfg.db.GetTaskEditorsByTaskID(r.Context(), task.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve task editors", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		ID:          task.ID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		UserID:      task.UserID,
		Title:       task.Title,
		EndDate:     task.EndDate,
		Description: task.Description,
		Priority:    task.Priority,
		Tag:         task.Tag,
		State:       task.State,
		ParentID:    task.ParentID.UUID,
		TaskEditors: taskEditors,
	})
}

func checkPriority(priority string) (err string) {
	switch priority {
	case "low", "medium", "high", "urgent":
		return
	default:
		return "Invalid priority value"
	}
}

func checkState(state string) (err string) {
	switch state {
	case "pending", "in progress", "done", "cancelled":
		return ""
	default:
		return "Invalid state value"
	}
}

func checkTag(tag string) (err string) {
	switch tag {
	case "private", "collaborative":
		return ""
	default:
		return "Invalid tag value"
	}
}

func checkDateFormat(date time.Time) (err string) {
	if date.IsZero() {
		return "Invalid date format"
	}
	return ""
}
