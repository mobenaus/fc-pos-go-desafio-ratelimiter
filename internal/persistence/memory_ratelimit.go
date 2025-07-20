package persistence

import (
	"time"
)

type MemoryRateLimitPersistence struct {
	capacity int
	period   time.Duration
	buckets  map[string]*Bucket
}

func NewMemoryRateLimitPersistence(capacity int, period time.Duration) *MemoryRateLimitPersistence {
	return &MemoryRateLimitPersistence{
		capacity: capacity,
		period:   period,
		buckets:  make(map[string]*Bucket),
	}
}

func (p *MemoryRateLimitPersistence) Refill(bucket *Bucket) {
	bucket.LastReffil = time.Now()
	bucket.Tokens = p.capacity
}

func (p *MemoryRateLimitPersistence) GetBucket(key string) (*Bucket, error) {
	bucket, ok := (p.buckets)[key]
	if !ok {
		bucket = &Bucket{
			LastReffil: time.Now(),
			Tokens:     p.capacity,
		}
	}

	return bucket, nil
}

func (p *MemoryRateLimitPersistence) CheckRefill(bucket *Bucket) bool {
	last := time.Since(bucket.LastReffil)
	return last.Milliseconds() > p.period.Milliseconds()
}

func (p *MemoryRateLimitPersistence) SaveBucket(key string, bucket *Bucket) error {
	(p.buckets)[key] = bucket
	return nil
}
