package persistence

import (
	"time"
)

type Bucket struct {
	Tokens     int
	LastReffil time.Time
}

type RateLimitPersistence interface {
	SetRules(capacity int, period time.Duration)
	Refill(key string) *Bucket
	UseToken(key string) error
}
