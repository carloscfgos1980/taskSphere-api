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
			log.Printf("len collaborative tasks for this parent id: %+v", collaborativeTasks)
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
		collaborativeTasks, err := cfg.DB.GetParentTasks(c)
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

// UpdateTaskHandler is the handler for updating an existing task in the system
func UpdateTaskHandler(cfg *config.Config) gin.HandlerFunc {
	// Define a struct to hold the parameters for updating the task
	type parameters struct {
		Title       *string    `json:"title,omitempty"`
		EndDate     *time.Time `json:"end_date,omitempty"`
		Description *string    `json:"description,omitempty"`
		Priority    *string    `json:"priority,omitempty"`
		State       *string    `json:"state,omitempty"`
	}
	// Return a handler function that can be used in the Gin router
	return func(c *gin.Context) {
		// Extract the task ID from the URL parameters and validate it
		taskIDString := c.Param("taskID")
		taskID, err := uuid.Parse(taskIDString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		// Extract the user ID from the context (set by the authentication middleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}

		// Retrieve the task from the database using the provided configuration and task ID to check if the user has access to update the task
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

		// Check if the user making the request is the owner of the task or has access to it (either as an editor or as a collaborator) to determine if they are authorized to update the task
		isAuthorized := dbTask.UserID == userID || dbTask.ID == taskID

		if !isAuthorized {
			for _, editorID := range dbTask.TaskEditors {
				if editorID == userID {
					isAuthorized = true
					break
				}
			}
		}
		// If the user is not authorized, respond with a 403 Forbidden error
		if !isAuthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to update this task"})
			return
		}
		// Bind the incoming JSON request to the parameters struct
		var params parameters
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Set missing fields to their current values in the database to ensure that only the provided fields are updated while the others remain unchanged
		title := dbTask.Title
		if params.Title != nil {
			title = *params.Title
		}
		endDate := dbTask.EndDate
		if params.EndDate != nil {
			endDate = *params.EndDate
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
		// Update the task in the database using the provided configuration and the updated fields, and then return the updated task information in the response
		updatedTaks, err := cfg.DB.UpdateTask(c, database.UpdateTaskParams{
			ID:          taskID,
			Title:       title,
			EndDate:     endDate,
			Description: description,
			Priority:    priority,
			State:       state,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
			return
		}
		// Prepare the response struct with the updated task information
		response := Task{
			ID:          updatedTaks.ID,
			CreatedAt:   updatedTaks.CreatedAt,
			UpdatedAt:   updatedTaks.UpdatedAt,
			UserID:      updatedTaks.UserID,
			Title:       updatedTaks.Title,
			EndDate:     updatedTaks.EndDate,
			Description: updatedTaks.Description,
			Priority:    updatedTaks.Priority,
			Tag:         updatedTaks.Tag,
			State:       updatedTaks.State,
		}
		// Return the updated task in the response with a 200 OK status
		c.JSON(http.StatusOK, response)
	}
}

// DeleteTaskHandler is the handler for deleting an existing task from the system
func DeleteTaskHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the task ID from the URL parameters and validate it
		taskIDString := c.Param("taskID")
		taskID, err := uuid.Parse(taskIDString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		// Extract the user ID from the context (set by the authentication middleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		// Retrieve the task from the database using the provided configuration and task ID to check if the user has access to delete the task
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
		// Check if the user making the request is the owner of the task or has access to it to determine if they are authorized to delete the task
		isAuthorized := dbTask.UserID == userID || dbTask.ID == taskID
		if !isAuthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to delete this task"})
			return
		}
		// Check if the task is a parent task of a collaborative task by checking if the ParentID is null and the tag is "collaborative". If it is a parent task, then delete all the associated collaborative tasks as well. If it is not a parent task, then just delete the task itself.
		isAdmin := false
		if dbTask.ParentID.UUID == uuid.Nil && dbTask.Tag == "collaborative" {
			isAdmin = true
		}
		// If the task is a parent task of a collaborative task, retrieve all the associated collaborative tasks from the database and then delete them along with the parent task.
		if isAdmin {
			collaborativeTasks, err := cfg.DB.GetCollaborativeTasksByParentID(c, uuid.NullUUID{UUID: taskID, Valid: true})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve collaborative tasks"})
				return
			}
			// Delete the parent task and all its associated collaborative tasks from the database using the provided configuration and task ID, and then return a success message in the response
			for _, task := range collaborativeTasks {
				log.Printf("deleting task with id: %s", task.ID)
				err = cfg.DB.DeleteTask(c, task.ID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
					return
				}
			}
			// respond with a success message in the response with a 200 OK status after deleting the parent task and all its associated collaborative tasks
			c.JSON(http.StatusOK, gin.H{"message": "Parent task and its collaborative tasks deleted successfully"})
			return
			// If the task is not a parent task of a collaborative task, then just delete the task itself from the database using the provided configuration and task ID, and then return a success message in the response
		} else {
			// Delete the task from the database using the provided configuration and task ID
			err = cfg.DB.DeleteTask(c, taskID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
				return
			}
			// Return a success message in the response with a 200 OK status
			c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
		}

	}
}
