package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/auth"
	"github.com/carloscfgos1980/taskSphere-api/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerTasksCreateIntegration(t *testing.T) {
	// Mark this test as a helper function so that errors are reported at the correct line in the test code
	t.Helper()

	_ = godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		t.Skip("DB_URL not set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, dbConn.Close())
	})

	queries := database.New(dbConn)
	jwtSecret := "integration-test-secret"
	apiCfg := apiConfig{
		db:        queries,
		jwtSecret: jwtSecret,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/tasks", apiCfg.handlerTasksCreate)

	username := fmt.Sprintf("task-owner-%d", time.Now().UnixNano())
	email := fmt.Sprintf("%s@example.com", username)
	hashedPassword, err := auth.HashPassword("StrongP@ssw0rd")
	require.NoError(t, err)

	createdUser, err := queries.CreateUser(context.Background(), database.CreateUserParams{
		Username: username,
		Email:    email,
		Password: hashedPassword,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, cleanupTasksErr := dbConn.ExecContext(context.Background(), "DELETE FROM tasks WHERE user_id = $1", createdUser.ID)
		assert.NoError(t, cleanupTasksErr)
		_, cleanupUserErr := dbConn.ExecContext(context.Background(), "DELETE FROM users WHERE id = $1", createdUser.ID)
		assert.NoError(t, cleanupUserErr)
	})

	token, err := auth.MakeJWT(createdUser.ID, jwtSecret, time.Hour)
	require.NoError(t, err)

	endDate := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	body, err := json.Marshal(map[string]any{
		"title":        "Write integration test",
		"end_date":     endDate,
		"description":  "Verify POST /api/tasks persists a task",
		"priority":     "high",
		"tag":          "private",
		"state":        "pending",
		"parent_id":    nil,
		"task_editors": []uuid.UUID{},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))

	var response Task
	err = json.NewDecoder(res.Body).Decode(&response)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, response.ID)
	assert.Equal(t, createdUser.ID, response.UserID)
	assert.Equal(t, "Write integration test", response.Title)
	assert.WithinDuration(t, endDate, response.EndDate, time.Second)
	assert.Equal(t, "Verify POST /api/tasks persists a task", response.Description)
	assert.Equal(t, "high", response.Priority)
	assert.Equal(t, "private", response.Tag)
	assert.Equal(t, "pending", response.State)
	assert.Empty(t, response.TaskEditors)

	persistedTask, err := queries.GetTaskByID(context.Background(), response.ID)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, persistedTask.UserID)
	assert.Equal(t, "Write integration test", persistedTask.Title)
	assert.WithinDuration(t, endDate, persistedTask.EndDate, time.Second)
	assert.Equal(t, "Verify POST /api/tasks persists a task", persistedTask.Description)
	assert.Equal(t, "high", persistedTask.Priority)
	assert.Equal(t, "private", persistedTask.Tag)
	assert.Equal(t, "pending", persistedTask.State)
	assert.Empty(t, persistedTask.TaskEditors)
}
