package main

import (
	"log"
	"net/http"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"

	"github.com/google/uuid"
)

// handlerTasksGet handles the retrieval of tasks based on the provided tag (private or collaborative) and, for collaborative tasks, the parent ID. It validates the user's authorization using JWT tokens and returns the appropriate tasks in JSON format.
func (cfg *apiConfig) handlerTasksGet(w http.ResponseWriter, r *http.Request) {
	// Extract the "tag" query parameter from the URL and validate it to determine whether to retrieve private or collaborative tasks
	paramsTag := r.URL.Query().Get("tag")
	tag, err := auth.CheckTag(paramsTag)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
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
	// Based on the value of the "tag" query parameter, retrieve either private or collaborative tasks from the database and prepare the response accordingly
	switch tag {
	// If the tag is "private", retrieve the private tasks associated with the user ID from the database and prepare the response with the retrieved tasks information
	case "private":
		// Retrieve private tasks associated with the user ID from the database
		privateTasks, err := cfg.db.GetTasksByUserID(r.Context(), userID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get tasks", err)
			return
		}
		if len(privateTasks) == 0 {
			respondWithError(w, http.StatusNotFound, "No private tasks found for this user", nil)
			return
		}
		// Prepare the response struct with the retrieved tasks information
		var response []Task
		for _, dbTask := range privateTasks {
			response = append(response, Task{
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
				TaskEditors: dbTask.TaskEditors,
			})
		}
		// Respond with the array of private tasks in JSON format
		respondWithJSON(w, http.StatusOK, response)
	case "collaborative":
		// Extract the parent_id query parameter and validate it
		paramsParentID := r.URL.Query().Get("parent_id")
		parentID, err := uuid.Parse(paramsParentID)
		if err != nil {
			http.Error(w, "Invalid parent_id", http.StatusBadRequest)
			return
		}
		//Retrieve collaborative tasks that are associated with the specified parent ID from the database
		collaborativeTasks, err := cfg.db.GetCollaborativeTasksByParentID(r.Context(), uuid.NullUUID{UUID: parentID, Valid: true})
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
		// Check if the user has permission to access the collaborative tasks by verifying if they are either the owner or an editor of any of the retrieved tasks
		hasPersmission := false
		for _, task := range collaborativeTasks {
			if task.UserID == userID {
				hasPersmission = true
				break
			}
			for _, editorID := range task.TaskEditors {
				if editorID == userID {
					hasPersmission = true
					break
				}
			}
			if hasPersmission {
				break
			}
		}
		if !hasPersmission {
			http.Error(w, "You don't have permission to access these collaborative tasks", http.StatusForbidden)
			return
		}
		// Prepare the response struct with the retrieved collaborative tasks information
		var response []Task
		for _, dbTask := range collaborativeTasks {
			response = append(response, Task{
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
				TaskEditors: dbTask.TaskEditors,
			})
		}
		// Respond with the array of collaborative tasks in JSON format
		respondWithJSON(w, http.StatusOK, response)
	default:
		http.Error(w, "Invalid tag value. Must be 'private' or 'collaborative'", http.StatusBadRequest)
	}

}

//		// Extract the task ID from the URL path and validate it
//		taskIDString := r.PathValue("taskID")
//		taskID, err := uuid.Parse(taskIDString)
//		if err != nil {
//			respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
//			return
//		}
//		// Retrieve the task from the database using the provided ID
//		dbTask, err := cfg.db.GetTaskByID(r.Context(), taskID)
//		if err != nil {
//			respondWithError(w, http.StatusNotFound, "Couldn't get task", err)
//			return
//		}
//		// If the task is marked as "private", validate the user's authorization to access it
//		if dbTask.Tag == "private" {
//			// Extract the Bearer token from the Authorization header
//			token, err := auth.GetBearerToken(r.Header)
//			if err != nil {
//				respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
//				return
//			}
//			// Verify JWT token and extract the user ID
//			userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
//			if err != nil {
//				respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
//				return
//			}
//			// Check if the user ID from the token matches the user ID associated with the task
//			if dbTask.UserID != userID {
//				respondWithError(w, http.StatusForbidden, "You don't have access to this task", nil)
//				return
//			}
//		}
//		// Respond with the task details in JSON format
//		respondWithJSON(w, http.StatusOK, dbTask)
//	}
func (cfg *apiConfig) handlerGetTaskByID(w http.ResponseWriter, r *http.Request) {
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
	// Extract the task ID from the route path and validate it
	taskIDString := r.PathValue("taskID")
	if taskIDString == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}
	taskID, err := uuid.Parse(taskIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}
	// Retrieve the task from the database using the provided ID
	dbTask, err := cfg.db.GetTaskByID(r.Context(), taskID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get task", err)
		return
	}
	// If the task is marked as "private", check if the user has permission to access it by verifying if they are either the owner or an editor of the task
	if dbTask.Tag == "private" {
		hasPersmission := false
		if dbTask.UserID == userID {
			hasPersmission = true
		}
		for _, editorID := range dbTask.TaskEditors {
			if editorID == userID {
				hasPersmission = true
				break
			}
		}
		if !hasPersmission {
			http.Error(w, "You don't have access to this task", http.StatusForbidden)
			return
		}
	}
	if dbTask.Tag == "collaborative" {
		hasPersmission := false
		if dbTask.UserID == userID {
			hasPersmission = true
		}
		for _, editorID := range dbTask.TaskEditors {
			if editorID == userID {
				hasPersmission = true
				break
			}
		}
		if !hasPersmission {
			http.Error(w, "You don't have access to this task", http.StatusForbidden)
			return
		}
	}
	// prepare the response struct with the retrieved task information
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
		TaskEditors: dbTask.TaskEditors,
	}
	// Respond with the task details in JSON format
	respondWithJSON(w, http.StatusOK, response)
}
