package tasks

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/carloscfgos1980/taskSphere-api/internal/database"
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

	userIDValue := r.Context().Value("userID")
	userUUID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := userUUID.String()

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
	createdTask, err := h.service.CreateTask(r.Context(), task)
	if err != nil {
		log.Printf("create task error: %v", err)
		http.Error(w, "Failed to create task: "+err.Error(), http.StatusInternalServerError)
		return
	}
	type response struct {
		Task database.Task `json:"task"`
	}
	resp := response{
		Task: createdTask,
	}
	if err := json.WriteJSON(w, http.StatusOK, resp); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
