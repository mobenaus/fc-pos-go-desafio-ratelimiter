package ratelimit

import (
	"fmt"
	"time"

	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/persistence"
)

type RateLimitConfig struct {
	Prefix      string
	Capacity    int
	Period      time.Duration
	Persistence persistence.RateLimitPersistence
}

type RateLimit struct {
	config RateLimitConfig
	key    string
}

func NewRateLimitConfig(prefix string, capacity int, period time.Duration, persistence persistence.RateLimitPersistence) RateLimitConfig {
	rlc := RateLimitConfig{
		Prefix:      prefix,
		Capacity:    capacity,
		Period:      period,
		Persistence: persistence,
	}
	rlc.Persistence.SetRules(capacity, period)
	return rlc
}

func (r *RateLimitConfig) GetRateLimit(key string) *RateLimit {
	return &RateLimit{
		config: *r,
		key:    key,
	}
}

func (rl *RateLimit) UseToken() error {
	fmt.Printf("check limit: %s %s\n", rl.config.Prefix, rl.key)
	error := rl.config.Persistence.UseToken(rl.key)
	return error
}
