package persistence

import (
	"errors"
	"sync"
	"time"
)

type MemoryRateLimitPersistence struct {
	capacity int
	period   time.Duration
	buckets  map[string]*Bucket
	mutex    sync.Mutex
}

func NewMemoryRateLimitPersistence(capacity int, period time.Duration) *MemoryRateLimitPersistence {
	return &MemoryRateLimitPersistence{
		capacity: capacity,
		period:   period,
		buckets:  make(map[string]*Bucket),
	}
}

func (p *MemoryRateLimitPersistence) refill(key string) *Bucket {
	bucket, ok := (p.buckets)[key]
	if !ok {
		bucket = &Bucket{
			Tokens:     p.capacity,
			LastReffil: time.Now(),
		}
		(p.buckets)[key] = bucket
	} else {
		bucket.LastReffil = time.Now()
		bucket.Tokens = p.capacity
	}
	return bucket
}

func (p *MemoryRateLimitPersistence) UseToken(key string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	bucket, ok := (p.buckets)[key]
	if !ok {
		bucket = p.refill(key)
	}

	last := time.Since(bucket.LastReffil)
	if last.Milliseconds() > p.period.Milliseconds() {
		bucket = p.refill(key)
	}

	if bucket.Tokens < 1 {
		return errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
	}

	bucket.Tokens -= 1

	return nil
}
