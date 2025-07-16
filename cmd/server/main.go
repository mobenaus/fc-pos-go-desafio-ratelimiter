package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

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

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password set
		DB:       0,                // Default DB
	})
	defer rdb.Close()

	ctx := context.Background()

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Error connecting to Redis: %v", err))
	}
	fmt.Println("Connected to Redis:", pong)

	// implementação com mapa em memoria
	//ipconfig := ratelimit.NewRateLimitConfig(persistence.NewMemoryRateLimitPersistence(5, time.Second))
	//tokenconfig := ratelimit.NewRateLimitConfig(persistence.NewMemoryRateLimitPersistence(5, time.Second))

	// implementação com mapa no REDIS
	ipconfig := ratelimit.NewRateLimitConfig(persistence.NewRedisRateLimitPersistence(ctx, rdb, "IP", 5, time.Second))
	tokenconfig := ratelimit.NewRateLimitConfig(persistence.NewRedisRateLimitPersistence(ctx, rdb, "TOKEN", 5, time.Second))

	config := middleware.NewRateLimitConfig(ipconfig, tokenconfig)

	rateLimitMiddleWare := config.RateLimitMiddleware()

	err = http.ListenAndServe(":8080",
		middleware.LoggingMiddleware(
			rateLimitMiddleWare(
				mux)))
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
