package persistence

import (
	"time"
)

type Bucket struct {
	Tokens     int
	LastReffil time.Time
}

type RateLimitPersistence interface {
	GetBucket(key string) (*Bucket, error)
	CheckRefill(lastReffil time.Time) bool
	Refill(bucket *Bucket)
	SaveBucket(key string, bucket *Bucket) error
}
