package main

import (
	"log"
	"net/http"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerTasksGet(w http.ResponseWriter, r *http.Request) {
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
	switch tag {
	case "private":
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

// func (cfg *apiConfig) handlerTasksGet(w http.ResponseWriter, r *http.Request) {
// 	// Extract the task ID from the URL path and validate it
// 	taskIDString := r.PathValue("taskID")
// 	taskID, err := uuid.Parse(taskIDString)
// 	if err != nil {
// 		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
// 		return
// 	}
// 	// Retrieve the task from the database using the provided ID
// 	dbTask, err := cfg.db.GetTaskByID(r.Context(), taskID)
// 	if err != nil {
// 		respondWithError(w, http.StatusNotFound, "Couldn't get task", err)
// 		return
// 	}
// 	// If the task is marked as "private", validate the user's authorization to access it
// 	if dbTask.Tag == "private" {
// 		// Extract the Bearer token from the Authorization header
// 		token, err := auth.GetBearerToken(r.Header)
// 		if err != nil {
// 			respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
// 			return
// 		}
// 		// Verify JWT token and extract the user ID
// 		userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
// 		if err != nil {
// 			respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
// 			return
// 		}
// 		// Check if the user ID from the token matches the user ID associated with the task
// 		if dbTask.UserID != userID {
// 			respondWithError(w, http.StatusForbidden, "You don't have access to this task", nil)
// 			return
// 		}
// 	}
// 	// Respond with the task details in JSON format
// 	respondWithJSON(w, http.StatusOK, dbTask)
// }

// func (cfg *apiConfig) handlerTasksGetPersonal(w http.ResponseWriter, r *http.Request) {
// 	// Extract the Bearer token from the Authorization header and validate it to get the user ID
// 	token, err := auth.GetBearerToken(r.Header)
// 	if err != nil {
// 		respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
// 		return
// 	}
// 	// Validate the JWT token and extract the user ID
// 	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
// 	if err != nil {
// 		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
// 		return
// 	}

// 	// Retrieve the tasks associated with the user ID from the database
// 	dbTasks, err := cfg.db.GetTasksByUserID(r.Context(), userID)
// 	if err != nil {
// 		respondWithError(w, http.StatusInternalServerError, "Couldn't get tasks", err)
// 		return
// 	}

// 	// Iterate through the retrieved tasks, get their editors, and build the response array
// 	var response []Task

// 	for _, dbTask := range dbTasks {
// 		response = append(response, Task{
// 			ID:          dbTask.ID,
// 			CreatedAt:   dbTask.CreatedAt,
// 			UpdatedAt:   dbTask.UpdatedAt,
// 			UserID:      dbTask.UserID,
// 			Title:       dbTask.Title,
// 			EndDate:     dbTask.EndDate,
// 			Description: dbTask.Description,
// 			Priority:    dbTask.Priority,
// 			Tag:         dbTask.Tag,
// 			State:       dbTask.State,
// 			ParentID:    dbTask.ParentID.UUID,
// 			TaskEditors: dbTask.TaskEditors,
// 		})
// 	}
// 	// Respond with the array of tasks in JSON format
// 	respondWithJSON(w, http.StatusOK, response)
// }

// func (cfg *apiConfig) handlerTasksGetCollaborative(w http.ResponseWriter, r *http.Request) {
// 	// Extract the parent task ID from the URL path and validate it
// 	parentIDString := r.PathValue("parentID")
// 	ParentID, err := uuid.Parse(parentIDString)
// 	if err != nil {
// 		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
// 		return
// 	}
// 	// Retrieve the collaborative tasks associated with the parent task ID from the database

// 	dbGroupTasks, err := cfg.db.GetCollaborativeTasksByParentID(r.Context(), uuid.NullUUID{UUID: ParentID, Valid: true})
// 	// If there's an error retrieving the collaborative tasks, respond with an internal server error
// 	if err != nil {
// 		respondWithError(w, http.StatusInternalServerError, "Couldn't get collaborative tasks", err)
// 		return
// 	}
// 	// If no collaborative tasks are found for the parent task, respond with a not found error
// 	if len(dbGroupTasks) == 0 {
// 		respondWithError(w, http.StatusNotFound, "No collaborative tasks found for this parent task", nil)
// 		return
// 	}

// 	// Extract the Bearer token from the Authorization header
// 	token, err := auth.GetBearerToken(r.Header)
// 	if err != nil {
// 		respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
// 		return
// 	}
// 	// Verify JWT token and extract the user ID
// 	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
// 	if err != nil {
// 		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
// 		return
// 	}
// 	// Define the response structure for a single collaborative task
// 	type taskResponse struct {
// 		ID          uuid.UUID   `json:"id"`
// 		CreatedAt   time.Time   `json:"created_at"`
// 		UpdatedAt   time.Time   `json:"updated_at"`
// 		UserID      uuid.UUID   `json:"user_id"`
// 		Username    string      `json:"username"`
// 		Email       string      `json:"email"`
// 		Title       string      `json:"title"`
// 		EndDate     time.Time   `json:"end_date"`
// 		Description string      `json:"description"`
// 		Priority    string      `json:"priority"`
// 		Tag         string      `json:"tag"`
// 		State       string      `json:"state"`
// 		ParentID    uuid.UUID   `json:"parent_id,omitempty"`
// 		TaskEditors []uuid.UUID `json:"task_editors"`
// 	}
// 	// Iterate through the retrieved collaborative tasks, check if the user is part of the work group, and build the response array
// 	isAuthorized := false
// 	var response []taskResponse
// 	for _, dbGroupTask := range dbGroupTasks {
// 		if dbGroupTask.UserID == userID {
// 			isAuthorized = true
// 		}
// 		// Append the collaborative task details along with its editors to the response array
// 		response = append(response, taskResponse{
// 			ID:          dbGroupTask.ID,
// 			CreatedAt:   dbGroupTask.CreatedAt,
// 			UpdatedAt:   dbGroupTask.UpdatedAt,
// 			UserID:      dbGroupTask.UserID,
// 			Username:    dbGroupTask.Username,
// 			Email:       dbGroupTask.Email,
// 			Title:       dbGroupTask.Title,
// 			EndDate:     dbGroupTask.EndDate,
// 			Description: dbGroupTask.Description,
// 			Priority:    dbGroupTask.Priority,
// 			Tag:         dbGroupTask.Tag,
// 			State:       dbGroupTask.State,
// 			ParentID:    dbGroupTask.ParentID.UUID,
// 			TaskEditors: dbGroupTask.TaskEditors,
// 		})
// 	}
// 	// If the user is not authorized, respond with an unauthorized error
// 	if !isAuthorized {
// 		respondWithError(w, http.StatusUnauthorized, "you are not part of this work group", nil)
// 		return
// 	}
// 	// Respond with the array of collaborative tasks in JSON format
// 	respondWithJSON(w, http.StatusOK, response)

// }
