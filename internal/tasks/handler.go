package tasks

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

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
		TaskEditors: func() []uuid.UUID {
			editors := make([]uuid.UUID, len(taskEditors))
			for i, editor := range taskEditors {
				editors[i] = editor
			}
			return editors
		}(),
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
	// Validate the task ID
	if _, err := uuid.Parse(taskID); err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
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

// GetTasks handles the retrieval of tasks based on the tag query parameter and user access level
func (h *handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	// Extract the tag query parameter and validate it
	paramsTag := r.URL.Query().Get("tag")
	tag, err := utils.CheckTag(paramsTag)
	if err != nil {
		http.Error(w, "Invalid tag", http.StatusBadRequest)
		return
	}
	// Extract the user ID from the context (set by the authentication middleware)
	userIDValue := r.Context().Value("userID")
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
	// Handle the retrieval of tasks based on the tag value and user access level
	switch tag {
	// If the tag is "private", retrieve tasks that are private to the user from the database
	case "private":
		// Retrieve tasks that are private to the user from the database
		privateTasks, err := h.service.GetTasksByUserID(r.Context(), userID)
		if err != nil {
			log.Printf("get tasks error: %v", err)
			http.Error(w, "Failed to get tasks: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// If no private tasks are found for the user, return a 404 Not Found response
		if len(privateTasks) == 0 {
			http.Error(w, "No private tasks found", http.StatusNotFound)
			return
		}
		// Prepare the response struct with the retrieved tasks information
		response := make([]Task, len(privateTasks))
		for i, task := range privateTasks {
			// Convert []pgtype.UUID to []uuid.UUID
			var taskEditors []uuid.UUID
			if len(task.TaskEditors) > 0 {
				taskEditors = make([]uuid.UUID, len(task.TaskEditors))
				for j, editor := range task.TaskEditors {
					taskEditors[j] = uuid.UUID(editor.Bytes)
				}
			}
			// Create a response struct for each task with the retrieved information, including user details and task editors
			response[i] = Task{
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
		}
		// Write the retrieved tasks as JSON response
		if err := json.WriteJSON(w, http.StatusOK, response); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	// If the tag is "collaborative", retrieve tasks that are collaborative and associated with the specified parent ID from the database
	case "collaborative":
		// Extract the parent_id query parameter and validate it
		paramsParentID := r.URL.Query().Get("parent_id")
		parentID, err := uuid.Parse(paramsParentID)
		if err != nil {
			http.Error(w, "Invalid parent_id", http.StatusBadRequest)
			return
		}
		//Retrieve collaborative tasks that are associated with the specified parent ID from the database
		collaborativeTasks, err := h.service.GetCollaborativeTasksByParentID(r.Context(), parentID.String())
		if err != nil {
			log.Printf("get collaborative tasks error: %v", err)
			http.Error(w, "Failed to get collaborative tasks: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// If no collaborative tasks are found for the specified parent ID, return a 404 Not Found response
		if len(collaborativeTasks) == 0 {
			http.Error(w, "No collaborative tasks found for the specified parent ID", http.StatusNotFound)
			return
		}
		// Prepare the response struct with the retrieved collaborative tasks information, including user details and task editors

		// Check if the user making the request is the owner of any of the collaborative tasks or has access to them

		hasAccess := false
		var response []taskResponse
		for _, task := range collaborativeTasks {
			if task.UserID.String() == userID {
				hasAccess = true
			} else {
				for _, editor := range task.TaskEditors {
					if editor.String() == userID {
						hasAccess = true
					}
				}
			}
			var taskEditors []uuid.UUID
			if len(task.TaskEditors) > 0 {
				taskEditors = make([]uuid.UUID, len(task.TaskEditors))
				for j, editor := range task.TaskEditors {
					taskEditors[j] = uuid.UUID(editor.Bytes)
				}
			}
			response = append(response, taskResponse{
				ID:          uuid.UUID(task.ID.Bytes),
				CreatedAt:   task.CreatedAt.Time,
				UpdatedAt:   task.UpdatedAt.Time,
				UserID:      uuid.UUID(task.UserID.Bytes),
				Username:    task.Username,
				Email:       task.Email,
				Title:       task.Title,
				EndDate:     task.EndDate.Time,
				Description: task.Description,
				Priority:    task.Priority,
				Tag:         task.Tag,
				State:       task.State,
				TaskEditors: taskEditors,
			})
		}
		// If the user does not have access to any of the collaborative tasks, return a forbidden error
		if !hasAccess {
			http.Error(w, "You do not have access to these collaborative tasks", http.StatusForbidden)
			return
		}
		log.Println("length of response", len(response))
		// Write the retrieved collaborative tasks as JSON response
		if err := json.WriteJSON(w, http.StatusOK, response); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	// If the tag is "public", retrieve tasks that are public from the database
	case "public":
		// Retrieve tasks that are public from the database
		publicTasks, err := h.service.GetPublicTasks(r.Context())
		if err != nil {
			log.Printf("get public tasks error: %v", err)
			http.Error(w, "Failed to get public tasks: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// If no public tasks are found, return a 404 Not Found response
		if len(publicTasks) == 0 {
			http.Error(w, "No public tasks found", http.StatusNotFound)
			return
		}
		// Prepare the response struct with the retrieved public tasks information, including user details and task editors
		type publicTaskResponse struct {
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
			TaskEditors []uuid.UUID `json:"task_editors"`
		}
		// Convert the retrieved public tasks to the response struct format, including user details and task editors
		response := make([]publicTaskResponse, len(publicTasks))
		for i, task := range publicTasks {
			var taskEditors []uuid.UUID
			if len(task.TaskEditors) > 0 {
				taskEditors = make([]uuid.UUID, len(task.TaskEditors))
				for j, editor := range task.TaskEditors {
					taskEditors[j] = uuid.UUID(editor.Bytes)
				}
			}
			response[i] = publicTaskResponse{
				ID:          uuid.UUID(task.ID.Bytes),
				CreatedAt:   task.CreatedAt.Time,
				UpdatedAt:   task.UpdatedAt.Time,
				UserID:      uuid.UUID(task.UserID.Bytes),
				Username:    task.Username,
				Email:       task.Email,
				Title:       task.Title,
				EndDate:     task.EndDate.Time,
				Description: task.Description,
				Priority:    task.Priority,
				Tag:         task.Tag,
				State:       task.State,
				TaskEditors: taskEditors,
			}
		}
		// Write the retrieved public tasks as JSON response
		if err := json.WriteJSON(w, http.StatusOK, response); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	// If the tag value is invalid, return a bad request error
	default:
		http.Error(w, "Invalid tag parameter", http.StatusBadRequest)
		return
	}
}

// GetParentsCollaborativeTasks handles the retrieval of parent tasks that are collaborative and associated with the user making the request from the database
func (h *handler) GetParentsCollaborativeTasks(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the context (set by the authentication middleware)
	userIDValue := r.Context().Value("userID")
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
	// Retrieve parent tasks that are collaborative
	collaborativeTasks, err := h.service.GetParentTasks(r.Context())
	if err != nil {
		log.Printf("get collaborative tasks error: %v", err)
		http.Error(w, "Failed to get collaborative tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(collaborativeTasks) == 0 {
		http.Error(w, "No collaborative tasks found for the user", http.StatusNotFound)
		return
	}
	// Create a slice of taskResponse to hold the response data for parent collaborative tasks
	response := make([]taskResponse, len(collaborativeTasks))
	for i, task := range collaborativeTasks {
		var taskEditors []uuid.UUID
		if len(task.TaskEditors) > 0 {
			taskEditors = make([]uuid.UUID, len(task.TaskEditors))
			for j, editor := range task.TaskEditors {
				taskEditors[j] = uuid.UUID(editor.Bytes)
			}
		}
		response[i] = taskResponse{
			ID:          uuid.UUID(task.ID.Bytes),
			CreatedAt:   task.CreatedAt.Time,
			UpdatedAt:   task.UpdatedAt.Time,
			UserID:      uuid.UUID(task.UserID.Bytes),
			Username:    task.Username,
			Email:       task.Email,
			Title:       task.Title,
			EndDate:     task.EndDate.Time,
			Description: task.Description,
			Priority:    task.Priority,
			Tag:         task.Tag,
			State:       task.State,
			TaskEditors: taskEditors,
		}
	}
	// Write the response as JSON
	if err := json.WriteJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// UpdateTask handles the updating of a task by its ID, allowing only the owner or editors of the task to perform the update
func (h *handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	// Define a struct to hold the parameters for updating the task
	type parameters struct {
		Title       *string      `json:"title,omitempty"`
		EndDate     *time.Time   `json:"end_date,omitempty"`
		Description *string      `json:"description,omitempty"`
		Priority    *string      `json:"priority,omitempty"`
		State       *string      `json:"state,omitempty"`
		Tag         *string      `json:"tag,omitempty"`
		ParentID    *uuid.UUID   `json:"parent_id,omitempty"`
		TaskEditors *[]uuid.UUID `json:"task_editors,omitempty"`
	}
	// Get the task ID from the URL parameters
	taskID := chi.URLParam(r, "taskID")
	// Validate the task ID
	if _, err := uuid.Parse(taskID); err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
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
	// Get the task from the database to check if it exists and to verify the user's access level (owner or editor) for authorization to update the task
	dbTask, err := h.service.GetTaskByID(r.Context(), taskID)
	if err != nil {
		log.Printf("get task error: %v", err)
		http.Error(w, "Failed to get task: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Check if the user making the request is the owner of the task or has access to it (either as an editor or as a collaborator) to determine if they are authorized to update the task
	isAuthorized := dbTask.UserID.String() == userID
	if !isAuthorized && dbTask.Tag == "collaborative" && !dbTask.ParentID.Valid {
		http.Error(w, "You do not have access to update this task", http.StatusForbidden)
		return
	}
	if !isAuthorized && dbTask.Tag == "collaborative" {
		adminTask, err := h.service.GetTaskByID(r.Context(), dbTask.ParentID.String())
		if err != nil {
			log.Printf("Error retrieving admin task: %v", err)
			http.Error(w, "Failed to retrieve admin task", http.StatusInternalServerError)
			return
		}
		if adminTask.UserID.String() == userID {
			isAuthorized = true
		}
	}
	if !isAuthorized {
		for _, editorID := range dbTask.TaskEditors {
			if editorID.String() == userID {
				isAuthorized = true
				break
			}
		}
	}
	if !isAuthorized {
		http.Error(w, "You do not have access to update this task", http.StatusForbidden)
		return
	}
	// Parse the JSON request body into the parameters struct to get the fields that need to be updated for the task
	var params parameters
	if err := json.ReadJSON(r, &params); err != nil {
		log.Printf("Error parsing request body: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Set missing fields to their current values in the database to ensure that only the provided fields are updated while the others remain unchanged
	title := dbTask.Title
	if params.Title != nil {
		title = *params.Title
	}
	endDate := dbTask.EndDate
	if params.EndDate != nil {
		endDate = pgtype.Timestamp{Time: *params.EndDate, Valid: true}
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
	tag := dbTask.Tag
	if params.Tag != nil {
		tag = *params.Tag
	}
	parentID := dbTask.ParentID
	if params.ParentID != nil {
		parentID = pgtype.UUID{Bytes: *params.ParentID, Valid: true}
	}
	taskEditors := dbTask.TaskEditors
	if params.TaskEditors != nil {
		convertedEditors := make([]pgtype.UUID, len(*params.TaskEditors))
		for i, editor := range *params.TaskEditors {
			convertedEditors[i] = pgtype.UUID{Bytes: editor, Valid: true}
		}
		taskEditors = convertedEditors
	}
	// Create a Task with the updated fields and the user ID to pass to the service for updating the task in the database, ensuring that only the provided fields are updated while the others remain unchanged
	task := Task{
		ID:          uuid.MustParse(taskID),
		Title:       title,
		Description: description,
		EndDate:     endDate.Time,
		UserID:      uuid.UUID(dbTask.UserID.Bytes),
		Priority:    priority,
		State:       state,
		Tag:         tag,
		ParentID:    uuid.UUID(parentID.Bytes),
		TaskEditors: func() []uuid.UUID {
			editors := make([]uuid.UUID, len(taskEditors))
			for i, editor := range taskEditors {
				editors[i] = uuid.UUID(editor.Bytes)
			}
			return editors
		}(),
	}
	// Call the service to update the task in the database with the provided fields, ensuring that only the provided fields are updated while the others remain unchanged
	updatedTask, err := h.service.UpdateTask(r.Context(), task)
	if err != nil {
		log.Printf("Error updating task: %v", err)
		http.Error(w, "Failed to update task: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Create a response struct to send back to the client with the updated task information
	response := Task{
		ID:          uuid.UUID(updatedTask.ID.Bytes),
		Title:       updatedTask.Title,
		Description: updatedTask.Description,
		EndDate:     updatedTask.EndDate.Time,
		UserID:      uuid.UUID(updatedTask.UserID.Bytes),
		Priority:    updatedTask.Priority,
		State:       updatedTask.State,
		Tag:         updatedTask.Tag,
		ParentID:    uuid.UUID(updatedTask.ParentID.Bytes),
		TaskEditors: func() []uuid.UUID {
			editors := make([]uuid.UUID, len(updatedTask.TaskEditors))
			for i, editor := range updatedTask.TaskEditors {
				editors[i] = uuid.UUID(editor.Bytes)
			}
			return editors
		}(),
	}
	// Write the response as JSON
	if err := json.WriteJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// DeleteTask handles the deletion of a task by its ID, allowing only the owner or editors of the task to perform the deletion, and if the task is a parent collaborative task, it also deletes all associated child collaborative tasks
func (h *handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	// Get the task ID from the URL parameters
	taskID := chi.URLParam(r, "taskID")
	// Validate the task ID
	if _, err := uuid.Parse(taskID); err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
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
	// Get the task from the database to check if it exists and to verify the user's access level (owner or editor) for authorization to delete the task
	dbTask, err := h.service.GetTaskByID(r.Context(), taskID)
	if err != nil {
		log.Printf("get task error: %v", err)
		http.Error(w, "Failed to get task: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Check if the user making the request is the owner of the task
	isAuthorized := dbTask.UserID.String() == userID
	if !isAuthorized && dbTask.Tag == "collaborative" && !dbTask.ParentID.Valid {
		http.Error(w, "You do not have access to delete this task", http.StatusForbidden)
		return
	}
	// If the task is a collaborative task and the user is not the owner, check if they are the admin of the collaborative task (the owner of the parent task) to determine if they are authorized to delete the task
	if !isAuthorized && dbTask.Tag == "collaborative" {
		adminTask, err := h.service.GetTaskByID(r.Context(), dbTask.ParentID.String())
		if err != nil {
			log.Printf("Error retrieving admin task: %v", err)
			http.Error(w, "Failed to retrieve admin task", http.StatusInternalServerError)
			return
		}
		if adminTask.UserID.String() == userID {
			isAuthorized = true
		}
	}
	// If the user is not the owner and not the admin of the collaborative task, check if they are an editor of the task to determine if they are authorized to delete the task
	if !isAuthorized {
		http.Error(w, "You do not have access to delete this task", http.StatusForbidden)
		return
	}
	// If the task is a parent collaborative task, delete all associated child collaborative tasks as well, ensuring that only the owner or admin of the task can perform the deletion and that if the task is a parent collaborative task, all associated child collaborative tasks are also deleted
	isAdmin := false
	if !dbTask.ParentID.Valid && dbTask.Tag == "collaborative" {
		isAdmin = true
	}
	if isAdmin {
		collaborativeTasks, err := h.service.GetCollaborativeTasksByParentID(r.Context(), dbTask.ID.String())
		if err != nil {
			log.Printf("get collaborative tasks error: %v", err)
			http.Error(w, "Failed to get collaborative tasks: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, task := range collaborativeTasks {
			if err := h.service.DeleteTask(r.Context(), task.ID.String()); err != nil {
				log.Printf("Error deleting collaborative task: %v", err)
				http.Error(w, "Failed to delete collaborative task: "+err.Error(), http.StatusInternalServerError)
				return
			}
			json.WriteJSON(w, http.StatusOK, map[string]string{"message": "Collaborative task with ID " + task.ID.String() + " deleted successfully"})
		}
	} else {
		// Call the service to delete the task from the database
		if err := h.service.DeleteTask(r.Context(), taskID); err != nil {
			log.Printf("Error deleting task: %v", err)
			http.Error(w, "Failed to delete task: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Write a success response as JSON
		json.WriteJSON(w, http.StatusOK, map[string]string{"message": "Task deleted successfully"})
	}

}
