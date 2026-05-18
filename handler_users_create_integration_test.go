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

func TestHandlerUsersCreateIntegration(t *testing.T) {
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
	apiCfg := apiConfig{db: queries}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", apiCfg.handlerUsersCreate)

	username := fmt.Sprintf("integration-user-%d", time.Now().UnixNano())
	email := fmt.Sprintf("%s@example.com", username)
	password := "StrongP@ssw0rd"

	t.Cleanup(func() {
		_, cleanupErr := dbConn.ExecContext(context.Background(), "DELETE FROM users WHERE email = $1", email)
		assert.NoError(t, cleanupErr)
	})

	body, err := json.Marshal(map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))

	var response User
	err = json.NewDecoder(res.Body).Decode(&response)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, response.ID)
	assert.Equal(t, username, response.Username)
	assert.Equal(t, email, response.Email)
	assert.Empty(t, response.Password)

	createdUser, err := queries.GetUserByEmail(context.Background(), email)
	require.NoError(t, err)
	assert.Equal(t, username, createdUser.Username)
	assert.Equal(t, email, createdUser.Email)
	assert.NotEqual(t, password, createdUser.Password)

	passwordMatches, err := auth.CheckPasswordHash(password, createdUser.Password)
	require.NoError(t, err)
	assert.True(t, passwordMatches)
}
