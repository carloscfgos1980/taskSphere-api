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
	ParentID    uuid.UUID   `json:"parent_id,omitempty"`
	TaskEditors []uuid.UUID `json:"task_editors"`
}

func (cfg *apiConfig) handlerTasksCreate(w http.ResponseWriter, r *http.Request) {
	// Define the expected parameters for creating a new task and the response structure
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
	// Define the response structure for a single task
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
	if errPriority := CheckPriority(params.Priority); errPriority != "" {
		respondWithError(w, http.StatusBadRequest, errPriority, nil)
		return
	}
	if errState := CheckState(params.State); errState != "" {
		respondWithError(w, http.StatusBadRequest, errState, nil)
		return
	}
	resultTag, errTag := CheckTag(params.Tag)
	if errTag != "" {
		respondWithError(w, http.StatusBadRequest, errTag, nil)
		return
	}

	log.Printf("Creating task for user %s with title %s", userID, params.Title)
	// Create a new task in the database using the provided parameters and the user ID extracted from the JWT token
	dbTask, err := cfg.db.CreateTask(r.Context(), database.CreateTaskParams{
		UserID:      userID,
		Title:       params.Title,
		EndDate:     params.EndDate,
		Description: params.Description,
		Priority:    params.Priority,
		Tag:         resultTag,
		State:       params.State,
		ParentID:    uuid.NullUUID{UUID: params.ParentID, Valid: params.ParentID != uuid.Nil},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create task", err)
		return
	}
	// If task editors are provided in the request, create entries in the database to associate them with the newly created task
	dbTaskEditors := []uuid.UUID{}
	if len(params.TaskEditors) > 0 {
		log.Printf("Adding %d editors to task", len(params.TaskEditors))
		for _, editorID := range params.TaskEditors {
			_, err := cfg.db.CreateTaskEditors(r.Context(), database.CreateTaskEditorsParams{
				TaskID:   dbTask.ID,
				EditorID: editorID,
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't create task editors", err)
				return
			}
			dbTaskEditors = append(dbTaskEditors, editorID)
		}
	}
	// Respond with the created task's information, including its editors
	respondWithJSON(w, http.StatusCreated, response{
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

func CheckPriority(priority string) (err string) {
	switch priority {
	case "low", "medium", "high", "urgent":
		return
	default:
		return "Invalid priority value"
	}
}

func CheckState(state string) (err string) {
	switch state {
	case "pending", "in progress", "done", "cancelled":
		return ""
	default:
		return "Invalid state value"
	}
}

func CheckTag(tag string) (resultTag, err string) {
	switch tag {
	case "":
		return "private", ""
	case "private", "collaborative", "public":
		return tag, ""
	default:
		return "", "Invalid tag value"
	}
}
