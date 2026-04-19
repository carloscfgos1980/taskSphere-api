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
func GetTasksByIdHandler(cfg *config.Config) gin.HandlerFunc {
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
