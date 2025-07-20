package ratelimit

import (
	"errors"
	"sync"

	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/persistence"
)

type RateLimit struct {
	Persistence persistence.RateLimitPersistence
	mutex       sync.Mutex
}

func NewRateLimit(persistence persistence.RateLimitPersistence) *RateLimit {
	rl := &RateLimit{
		Persistence: persistence,
	}
	return rl
}

func (rl *RateLimit) UseToken(key string) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	p := rl.Persistence

	bucket, error := p.GetBucket(key)
	if error != nil {
		return errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
	}

	if p.CheckRefill(bucket) {
		p.Refill(bucket)
	}

	if bucket.Tokens < 1 {
		return errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
	}

	bucket.Tokens--

	error = p.SaveBucket(key, bucket)
	if error != nil {
		return errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
	}
	return nil
}
