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

func JSON(w http.ResponseWriter, status int, success bool, message string, data any) {
	resp := ApiResponse{
		Status:  status,
		Success: success,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, `{"status": 500, "success": false, "message": "internal
        server error"}`, http.StatusInternalServerError)
	}
}
