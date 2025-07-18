package handler

import (
	"encoding/json"
	"net/http"
)

type HandlerResult struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	result := HandlerResult{
		Message: "Hello, World!",
		Data:    map[string]string{"info": "Sample data"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
