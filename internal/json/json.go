package json

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes the given data as JSON to the response writer with the specified status code
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

// ReadJSON reads the JSON from the request body and decodes it into the given data structure
func ReadJSON(r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}
