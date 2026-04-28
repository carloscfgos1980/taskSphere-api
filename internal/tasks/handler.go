package tasks

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/carloscfgos1980/taskSphere-api/internal/json"
	"github.com/carloscfgos1980/taskSphere-api/internal/utils"
)

// handler is the HTTP handler for users endpoints
type handler struct {
	service   Service
	jwtSecret string
}

// NewHandler creates a new handler for users endpoints
func NewHandler(service Service, jwtSecret string) *handler {
	return &handler{
		service:   service,
		jwtSecret: jwtSecret,
	}
}

// CreateTask handles the creation of a new task for a user
func (h *handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	// Define the expected parameters for creating a new task and the response structure
	type parameters struct {
		Title       string      `json:"title" binding:"required"`
		EndDate     time.Time   `json:"end_date" binding:"required"`
		Description string      `json:"description" binding:"required"`
		Priority    string      `json:"priority"`
		Tag         string      `json:"tag"`
		State       string      `json:"state"`
		ParentID    uuid.UUID   `json:"parent_id"`
		TaskEditors []uuid.UUID `json:"task_editors"`
	}
	// Get the user ID from the request context (set by the authentication middleware)
	userIDValue := r.Context().Value("userID")
	// Check if the user ID is present in the context
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Assert the user ID value to a UUID type
	userUUID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Convert the user ID to a string for database queries
	userID := userUUID.String()
	// Check if the user exists in the database
	_, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Println(err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	// Parse the JSON request body into a TaskRequest struct
	var taskReq parameters
	if err := json.ReadJSON(r, &taskReq); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Check if any field is empty
	if taskReq.Title == "" || taskReq.Description == "" || taskReq.EndDate.Equal((time.Time{})) {
		http.Error(w, "Title, description, and end date are required", http.StatusBadRequest)
		return
	}
	// Validate the priority, state, and tag values
	priority, err := utils.CheckPriority(taskReq.Priority)
	if err != nil {
		http.Error(w, "Invalid priority", http.StatusBadRequest)
		return
	}
	state, err := utils.CheckState(taskReq.State)
	if err != nil {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	tag, err := utils.CheckTag(taskReq.Tag)
	if err != nil {
		http.Error(w, "Invalid tag", http.StatusBadRequest)
		return
	}
	// Create a Task struct with the parsed data and the user ID
	task := Task{
		Title:       taskReq.Title,
		Description: taskReq.Description,
		EndDate:     taskReq.EndDate,
		UserID:      userUUID,
		Priority:    priority,
		State:       state,
		Tag:         tag,
		ParentID:    taskReq.ParentID,
		TaskEditors: taskReq.TaskEditors,
	}
	// Call the service to create the task in the database
	createdTask, err := h.service.CreateTask(r.Context(), task)
	if err != nil {
		log.Printf("create task error: %v", err)
		http.Error(w, "Failed to create task: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Convert []pgtype.UUID to []uuid.UUID
	var taskEditors []uuid.UUID
	if len(createdTask.TaskEditors) > 0 {
		taskEditors = make([]uuid.UUID, len(createdTask.TaskEditors))
		for i, editor := range createdTask.TaskEditors {
			taskEditors[i] = uuid.UUID(editor.Bytes)
		}
	}
	// Create a response struct to send back to the client
	response := Task{
		ID:          uuid.UUID(createdTask.ID.Bytes),
		Title:       createdTask.Title,
		Description: createdTask.Description,
		EndDate:     createdTask.EndDate.Time,
		UserID:      uuid.UUID(createdTask.UserID.Bytes),
		Priority:    createdTask.Priority,
		State:       createdTask.State,
		Tag:         createdTask.Tag,
		ParentID:    uuid.UUID(createdTask.ParentID.Bytes),
		TaskEditors: taskEditors,
	}
	// Write the response as JSON
	if err := json.WriteJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// GetTaskByID handles the retrieval of a task by its ID
func (h *handler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	// Get the task ID from the URL parameters
	taskID := chi.URLParam(r, "taskID")
	// Call the service to get the task from the database
	task, err := h.service.GetTaskByID(r.Context(), taskID)
	if err != nil {
		log.Printf("get task error: %v", err)
		http.Error(w, "Failed to get task: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Get the user ID from the request context (set by the authentication middleware)
	userIDValue := r.Context().Value("userID")
	// Check if the user ID is present in the context
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Assert the user ID value to a UUID type
	userUUID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Convert the user ID to a string for database queries
	userID := userUUID.String()
	// Check if the user has access to the task (is the owner or an editor)
	if task.UserID.String() != userID {
		// Check if the user is an editor of the task
		hasAccess := false
		for _, editor := range task.TaskEditors {
			if editor.String() == userID {
				hasAccess = true
				break
			}
		}
		// If the user is not the owner and not an editor, return a forbidden error
		if !hasAccess {
			http.Error(w, "You do not have access to this task", http.StatusForbidden)
			return
		}
	}
	// Convert []pgtype.UUID to []uuid.UUID
	var taskEditors []uuid.UUID
	if len(task.TaskEditors) > 0 {
		taskEditors = make([]uuid.UUID, len(task.TaskEditors))
		for i, editor := range task.TaskEditors {
			taskEditors[i] = uuid.UUID(editor.Bytes)
		}
	}
	// Create a response struct to send back to the client
	response := Task{
		ID:          uuid.UUID(task.ID.Bytes),
		Title:       task.Title,
		Description: task.Description,
		EndDate:     task.EndDate.Time,
		UserID:      uuid.UUID(task.UserID.Bytes),
		Priority:    task.Priority,
		State:       task.State,
		Tag:         task.Tag,
		ParentID:    uuid.UUID(task.ParentID.Bytes),
		TaskEditors: taskEditors,
	}
	// Write the response as JSON
	if err := json.WriteJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
