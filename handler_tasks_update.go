package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerTasksUpdate(w http.ResponseWriter, r *http.Request) {
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
	// Extract the task ID from the URL path
	taskIDString := r.PathValue("taskID")
	taskID, err := uuid.Parse(taskIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}
	// Get the user ID from the JWT token in the Authorization header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No authorization token included", err)
		return
	}
	// Validate the JWT token and extract the user ID
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	// Retrieve the task from the database using the provided ID
	dbTask, err := cfg.db.GetTaskByID(r.Context(), taskID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get task", err)
		return
	}
	isAuthorized := false
	switch dbTask.Tag {
	case "private":
		// If the task is marked as "private", check if the user ID from the token matches the user ID associated with the task
		if dbTask.UserID == userID {
			isAuthorized = true
		}
		if !isAuthorized {
			respondWithError(w, http.StatusForbidden, "You don't have access to this task", nil)
			return
		}

	case "collaborative":
		// If the task is marked as "collaborative", check if the user is either the owner or a task editor
		isAuthorized = dbTask.UserID == userID
		// check if user id is admin
		if !isAuthorized {
			admin, err := cfg.db.GetTaskByID(r.Context(), dbTask.ParentID.UUID)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't get parent task", err)
				return
			}
			if admin.UserID == userID {
				isAuthorized = true
			}
		}
		// If the user is not the owner, check if they are an editor of the task by checking if their user ID is in the list of task editors associated with the task
		if !isAuthorized {
			for _, editorID := range dbTask.TaskEditors {
				if editorID == userID {
					isAuthorized = true
					break
				}
			}
		}
		if !isAuthorized {
			respondWithError(w, http.StatusForbidden, "You don't have access to this task", nil)
			return
		}
	case "public":
		// If the task is marked as "public", allow any authenticated user to access it, so no additional checks are needed
		isAuthorized = true
	default:
		// If the task has an unrecognized tag, respond with a 500 Internal Server Error
		respondWithError(w, http.StatusInternalServerError, "Unrecognized task tag", errors.New("unrecognized task tag"))
		return
	}

	// Decode the request body to get the new task parameters
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// Prepare the parameters for updating the task, using the existing values from the database if the corresponding parameter is not provided in the request body
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
	tag := dbTask.Tag
	if params.Tag != nil {
		tag = *params.Tag
	}
	parentID := dbTask.ParentID
	if params.ParentID != nil {
		parentID = uuid.NullUUID{UUID: *params.ParentID, Valid: true}
	}
	taskEditors := dbTask.TaskEditors
	if params.TaskEditors != nil {
		taskEditors = *params.TaskEditors
	}

	// Validate the parameters (e.g., check date format, priority and state values)

	updatedTask, err := cfg.db.UpdateTask(r.Context(), database.UpdateTaskParams{
		ID:          taskID,
		Title:       title,
		EndDate:     endDate,
		Description: description,
		Priority:    priority,
		State:       state,
		Tag:         tag,
		ParentID:    parentID,
		TaskEditors: taskEditors,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update task", err)
		return
	}
	//prepare the response struct with the updated task information
	updatedTaskResponse := Task{
		ID:          updatedTask.ID,
		CreatedAt:   updatedTask.CreatedAt,
		UpdatedAt:   updatedTask.UpdatedAt,
		UserID:      updatedTask.UserID,
		Title:       updatedTask.Title,
		EndDate:     updatedTask.EndDate,
		Description: updatedTask.Description,
		Priority:    updatedTask.Priority,
		Tag:         updatedTask.Tag,
		State:       updatedTask.State,
		ParentID:    updatedTask.ParentID.UUID,
		TaskEditors: updatedTask.TaskEditors,
	}

	// Respond with the updated task details
	respondWithJSON(w, http.StatusOK, updatedTaskResponse)
}
