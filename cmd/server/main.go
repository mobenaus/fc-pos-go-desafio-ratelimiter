package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/middleware"
	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/persistence"
	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/ratelimit"
)

type HandlerResult struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	result := HandlerResult{
		Message: "Hello, World!",
		Data:    map[string]string{"info": "Sample data"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	fmt.Println("Server listening on :8080")

	ipconfig := ratelimit.NewRateLimitConfig("IP", 5, time.Second, persistence.NewMemoryRateLimitPersistence())
	tokenconfig := ratelimit.NewRateLimitConfig("API", 5, time.Second, persistence.NewMemoryRateLimitPersistence())
	config := middleware.NewRateLimitConfig(ipconfig, tokenconfig)

	rateLimitMiddleWare := config.RateLimitMiddleware()

	err := http.ListenAndServe(":8080",
		middleware.LoggingMiddleware(
			rateLimitMiddleWare(
				mux)))
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
