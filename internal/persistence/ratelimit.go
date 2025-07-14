package persistence

import (
	"errors"
	"time"
)

type Bucket struct {
	Tokens     int
	LastReffil time.Time
}

type RateLimitPersistence interface {
	SetRules(capacity int, period time.Duration)
	GetStatus(key string) *Bucket
	Refill(key string) *Bucket
	UseToken(key string) error
}

type MemoryRateLimitPersistence struct {
	capacity int
	period   time.Duration
	buckets  map[string]*Bucket
}

func NewMemoryRateLimitPersistence() *MemoryRateLimitPersistence {
	return &MemoryRateLimitPersistence{
		buckets: make(map[string]*Bucket),
	}
}

func (p *MemoryRateLimitPersistence) SetRules(capacity int, period time.Duration) {
	p.capacity = capacity
	p.period = period
}

func (p *MemoryRateLimitPersistence) GetStatus(key string) *Bucket {
	bucket, ok := (p.buckets)[key]
	if !ok {
		bucket = &Bucket{
			Tokens:     p.capacity,
			LastReffil: time.Now(),
		}
	}
	return bucket
}

func (p *MemoryRateLimitPersistence) Refill(key string) *Bucket {
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
	bucket, ok := (p.buckets)[key]
	if !ok {
		bucket = p.Refill(key)
	}

	last := time.Since(bucket.LastReffil)
	if last.Milliseconds() > p.period.Milliseconds() {
		bucket = p.Refill(key)
	}

	if bucket.Tokens < 1 {
		return errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
	}

	bucket.Tokens -= 1

	return nil
}
