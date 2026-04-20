package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/config"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"
	"github.com/carloscfgos1980/taskSphere-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Task represents a task in the system
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
	ParentID    uuid.UUID   `json:"parent_id"`
	TaskEditors []uuid.UUID `json:"task_editors"`
}

// CreateTaskHandler is the handler for creating a new task in the system
func CreateTaskHandler(cfg *config.Config) gin.HandlerFunc {
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
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Extract the user ID from the context (set by the authentication middleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		// Bind the incoming JSON request to the parameters struct
		var params parameters
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Validate the required fields and the values of priority, state, and tag if provided
		if params.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
			return
		}
		if params.EndDate.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "End date is required"})
			return
		}
		if params.Description == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
			return
		}
		priority, err := utils.CheckPriority(params.Priority)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		state, err := utils.CheckState(params.State)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		tag, err := utils.CheckTag(params.Tag)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create the task in the database using the provided configuration and parameters
		task, err := cfg.DB.CreateTask(c, database.CreateTaskParams{
			UserID:      userID.(uuid.UUID),
			Title:       params.Title,
			EndDate:     params.EndDate,
			Description: params.Description,
			Priority:    priority,
			Tag:         tag,
			State:       state,
			ParentID:    uuid.NullUUID{UUID: params.ParentID, Valid: params.ParentID != uuid.Nil},
			TaskEditors: params.TaskEditors,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
			return
		}
		// Prepare the response struct with the created task information
		response := Task{
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
			TaskEditors: params.TaskEditors,
		}
		// Return the created task in the response with a 201 Created status
		c.JSON(http.StatusCreated, response)
	}
}

// GetTasksByIdHandler is the handler for retrieving a task by its ID
func GetTaskByIdHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Extract the task ID from the URL parameters and validate it
		taskIDString := c.Param("taskID")
		// Parse the task ID string into a UUID format
		taskID, err := uuid.Parse(taskIDString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		// Retrieve the task from the database using the provided configuration and task ID
		dbTask, err := cfg.DB.GetTaskByID(c, taskID)
		if err != nil {
			log.Printf("Error retrieving task: %v", err)
			if err.Error() == "sql: no rows in result set" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task"})
			return
		}
		// Check if the user making the request is the owner of the task or has access to it
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		if dbTask.UserID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this task"})
			return
		}
		// Prepare the response struct with the retrieved task information
		response := Task{
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
		}
		// Return the retrieved task in the response with a 200 OK status
		c.JSON(http.StatusOK, response)
	}
}

// GetTasksHandler is the handler for retrieving tasks based on the provided tag and user access
func GetTasksHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Extract the tag query parameter and validate it
		paramTag := c.Query("tag")
		// Check if the tag parameter is provided and valid
		tag, err := utils.CheckTag(paramTag)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Extract the user ID from the context (set by the authentication middleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		// Retrieve tasks based on the tag and user access level using the provided configuration
		switch tag {
		// For private tasks, retrieve only the tasks that are owned by the user from the database and return them in the response
		case "private":
			// Retrieve tasks that are private to the user from the database
			privateTasks, err := cfg.DB.GetTasksByUserID(c, userID.(uuid.UUID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
				return
			}
			// If no private tasks are found for the user, return a 404 Not Found response
			if len(privateTasks) == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "No tasks found for the user"})
				return
			}
			// Prepare the response struct with the retrieved tasks information
			response := make([]Task, len(privateTasks))
			for i, task := range privateTasks {
				response[i] = Task{
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
				}
			}
			// Return the retrieved tasks in the response with a 200 OK status
			c.JSON(http.StatusOK, response)
			return
		// For collaborative tasks, the user must provide a parent_id query parameter to specify which collaborative tasks to retrieve
		case "collaborative":
			// Extract the parent_id query parameter and validate it
			stringParentID := c.Query("parent_id")
			if stringParentID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "parent_id query parameter is required for collaborative tasks"})
				return
			}
			// Parse the parent_id string into a UUID format
			parentID, err := uuid.Parse(stringParentID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parent_id"})
				return
			}
			// Retrieve collaborative tasks that are associated with the specified parent ID from the database
			collaborativeTasks, err := cfg.DB.GetCollaborativeTasksByParentID(c, uuid.NullUUID{UUID: parentID, Valid: true})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
				return
			}
			// If no collaborative tasks are found for the specified parent ID, return a 404 Not Found response
			if len(collaborativeTasks) == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "No collaborative tasks found for the parent task"})
				return
			}
			// Prepare the response struct with the retrieved collaborative tasks information, including user details and task editors
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
			// Check if the user making the request is the owner of any of the collaborative tasks or has access to them
			isAuthorized := false
			var response []taskResponse
			for _, task := range collaborativeTasks {
				if task.UserID == userID.(uuid.UUID) {
					isAuthorized = true
				}
				response = append(response, taskResponse{
					ID:          task.ID,
					CreatedAt:   task.CreatedAt,
					UpdatedAt:   task.UpdatedAt,
					UserID:      task.UserID,
					Username:    task.Username,
					Email:       task.Email,
					Title:       task.Title,
					EndDate:     task.EndDate,
					Description: task.Description,
					Priority:    task.Priority,
					Tag:         task.Tag,
					State:       task.State,
					ParentID:    task.ParentID.UUID,
					TaskEditors: task.TaskEditors,
				})
			}
			// If the user does not have access to any of the collaborative tasks, return a 403 Forbidden response
			if !isAuthorized {
				c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to these tasks"})
				return
			}
			// Return the retrieved collaborative tasks in the response with a 200 OK status
			c.JSON(http.StatusOK, response)
			return
		// For public tasks, retrieve all tasks that are tagged as public from the database and return them in the response
		case "public":
			// Retrieve public tasks from the database
			publicTasks, err := cfg.DB.GetPublicTasks(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
				return
			}
			// If no public tasks are found, return a 404 Not Found response
			if len(publicTasks) == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "No public tasks found"})
				return
			}
			// Prepare the response struct with the retrieved public tasks information, including user details
			response := make([]Task, len(publicTasks))
			for i, task := range publicTasks {
				response[i] = Task{
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
				}
			}
			// Return the retrieved public tasks in the response with a 200 OK status
			c.JSON(http.StatusOK, response)
			return
		// If the tag value is not valid, return a 400 Bad Request response
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag value. It must be 'private', 'collaborative', or 'public'"})
			return
		}

	}
}

// GetCollaborativeTasksHandler is the handler for retrieving collaborative tasks that are associated with a specified parent ID
func GetCollaborativeTasksHandler(cfg *config.Config) gin.HandlerFunc {
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Extract the user ID from the context (set by the authentication middleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		// Check if the user making the request is valid and exists in the database
		_, err := cfg.DB.GetUserByID(c, userID.(uuid.UUID))
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}
		// Retrieve collaborative tasks that are associated with the specified parent ID from the database
		collaborativeTasks, err := cfg.DB.GetCollaborativeTasks(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
			return
		}
		// If no collaborative tasks are found for the specified parent ID, return a 404 Not Found response
		if len(collaborativeTasks) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No collaborative tasks found"})
			return
		}
		// Prepare the response struct with the retrieved collaborative tasks information, including user details
		type taskResponse struct {
			ID          uuid.UUID `json:"id"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			Username    string    `json:"username"`
			Email       string    `json:"email"`
		}
		// loop through the retrieved collaborative tasks and then prepare the response struct with the task information and user details
		var response []taskResponse
		for _, task := range collaborativeTasks {
			response = append(response, taskResponse{
				ID:          task.ID,
				Title:       task.Title,
				Description: task.Description,
				Username:    task.Username,
				Email:       task.Email,
			})
		}
		// Return the retrieved collaborative tasks in the response with a 200 OK status
		c.JSON(http.StatusOK, response)
	}
}
