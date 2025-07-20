package main

import (
	"context"

	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/configs"
	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/handler"
	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/middleware"
	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/persistence"
	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/ratelimit"
)

func main() {
	configs := loadConfigurations()
	IPRatePeriod, err := time.ParseDuration(configs.IPRatePeriod)
	if err != nil {
		panic(fmt.Sprintf("Error load config IP Rate Period: %v", err))
	}
	TOKENRatePeriod, err := time.ParseDuration(configs.TOKENRatePeriod)
	if err != nil {
		panic(fmt.Sprintf("Error load config TOKEN Rate Period: %v", err))
	}

	var rdb *redis.Client
	ctx := context.Background()

	defer func() {
		if rdb != nil {
			rdb.Close()
		}
	}()

	var ipconfig *ratelimit.RateLimit
	var tokenconfig *ratelimit.RateLimit

	if configs.RateLimitStrategy == "REDIS" {
		rdb = connectRedis(ctx, configs)
		// implementação com mapa no REDIS
		ipconfig = ratelimit.NewRateLimit(persistence.NewRedisRateLimitPersistence(ctx, rdb, "IP", configs.IPRateLimit, IPRatePeriod))
		tokenconfig = ratelimit.NewRateLimit(persistence.NewRedisRateLimitPersistence(ctx, rdb, "TOKEN", configs.TOKENRateLimit, TOKENRatePeriod))
	} else {
		// implementação com mapa em memoria
		ipconfig = ratelimit.NewRateLimit(persistence.NewMemoryRateLimitPersistence(configs.IPRateLimit, IPRatePeriod))
		tokenconfig = ratelimit.NewRateLimit(persistence.NewMemoryRateLimitPersistence(configs.TOKENRateLimit, TOKENRatePeriod))
	}

	config := middleware.NewRateLimit(ipconfig, tokenconfig)

	rateLimitMiddleWare := config.RateLimitMiddleware()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Handler)
	fmt.Printf("Server listening on :%s\n", configs.WebServerPort)
	err = http.ListenAndServe(fmt.Sprintf(":%s", configs.WebServerPort),
		middleware.LoggingMiddleware(
			rateLimitMiddleWare(
				mux)))
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func connectRedis(ctx context.Context, configs *configs.Conf) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     configs.REDISAddr,
		Password: configs.REDISPassword,
		DB:       configs.REDISDefaultDB,
	})

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Error connecting to Redis: %v", err))
	}
	fmt.Println("Connected to Redis:", pong)
	return rdb
}

func loadConfigurations() *configs.Conf {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}
	return configs
}
