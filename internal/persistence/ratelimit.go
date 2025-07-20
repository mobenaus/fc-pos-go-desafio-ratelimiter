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
	CheckRefill(bucket *Bucket) bool
	Refill(bucket *Bucket)
	SaveBucket(key string, bucket *Bucket) error
}
