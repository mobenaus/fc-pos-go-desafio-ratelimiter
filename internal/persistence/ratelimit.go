package persistence

import (
	"time"
)

type Bucket struct {
	Tokens     int
	LastReffil time.Time
}

type RateLimitPersistence interface {
	UseToken(key string) error
}
