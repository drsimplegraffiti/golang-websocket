package utils

import (
	"encoding/json"
	"net/http"
)

type ApiResponse struct {
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// JSON here is a utility function that sends a JSON response to the client. It
// takes an http.ResponseWriter, status code, success flag, message, and data as
// parameters. It constructs an ApiResponse struct with these values, sets the
// Content-Type header to application/json, and writes the response with the
// specified status code. If encoding the response fails, it sends a 500
// Internal Server Error response.
func JSON(w http.ResponseWriter, status int, success bool, message string, data any) {
	resp := ApiResponse{
		Status:  status,
		Success: success,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// the json.NewEncoder(w).Encode(resp) function encodes the resp struct into
	// JSON format and writes it to the http.ResponseWriter. If an error occurs
	// during encoding, it sends a 500 Internal Server Error response with a
	// JSON error message.
	err := json.NewEncoder(w).Encode(resp) // converts golang struct to json and
	// writes it to the response writer
	if err != nil {
		http.Error(w, `{"status": 500, "success": false, "message": "internal
        server error"}`, http.StatusInternalServerError)
	}
}
