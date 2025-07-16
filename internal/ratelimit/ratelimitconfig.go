package ratelimit

import (
	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/persistence"
)

type RateLimit struct {
	Persistence persistence.RateLimitPersistence
}

func NewRateLimitConfig(persistence persistence.RateLimitPersistence) RateLimit {
	rl := RateLimit{
		Persistence: persistence,
	}
	return rl
}

func (rl *RateLimit) UseToken(key string) error {
	error := rl.Persistence.UseToken(key)
	return error
}
