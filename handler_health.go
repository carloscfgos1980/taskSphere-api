package main

import (
	"net/http"
)

// handlerHealth responds with a simple health check
func (cfg *apiConfig) handlerHealth(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": "1.0",
		"message": "TaskSphere API is healthy",
	})
}
