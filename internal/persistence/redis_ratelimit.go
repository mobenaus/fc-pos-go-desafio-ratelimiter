package persistence

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimitPersistence struct {
	ctx      context.Context
	rdb      *redis.Client
	prefix   string
	capacity int
	period   time.Duration
	mutex    sync.Mutex
}

func NewRedisRateLimitPersistence(ctx context.Context, rdb *redis.Client, prefix string, capacity int, period time.Duration) *RedisRateLimitPersistence {
	return &RedisRateLimitPersistence{
		ctx:      ctx,
		rdb:      rdb,
		prefix:   prefix,
		capacity: capacity,
		period:   period,
	}
}

func (p *RedisRateLimitPersistence) UseToken(key string) error {

	maxrequests := errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var bucket Bucket
	refillKey := fmt.Sprintf("%s:refill:%s", p.prefix, key)
	tokensKey := fmt.Sprintf("%s:tokens:%s", p.prefix, key)

	pipe := p.rdb.TxPipeline()

	exists, error := p.rdb.Exists(p.ctx, refillKey).Result()
	if error != nil {
		return maxrequests
	}
	if exists == 0 {
		bucket = p.refill()
	} else {
		bucket, error = p.fill(refillKey, tokensKey)
		if error != nil {
			return maxrequests
		}
	}

	last := time.Since(bucket.LastReffil)
	if last.Milliseconds() > p.period.Milliseconds() {
		bucket = p.refill()
	}

	if bucket.Tokens < 1 {
		return maxrequests
	}

	bucket.Tokens -= 1

	pipe.Set(p.ctx, refillKey, bucket.LastReffil.Format(time.RFC3339Nano), p.period)
	pipe.Set(p.ctx, tokensKey, bucket.Tokens, p.period)

	_, error = pipe.Exec(p.ctx)
	if error != nil {
		return maxrequests
	}

	return nil
}

func (p *RedisRateLimitPersistence) fill(refillKey string, tokensKey string) (Bucket, error) {
	rs := p.rdb.Get(p.ctx, refillKey).Val()
	refill, error := time.Parse(time.RFC3339Nano, rs)
	if error != nil {
		return Bucket{}, error
	}
	ts, error := p.rdb.Get(p.ctx, tokensKey).Int()
	if error != nil {
		return Bucket{}, error
	}

	return Bucket{
		LastReffil: refill,
		Tokens:     ts,
	}, nil
}

func (p *RedisRateLimitPersistence) refill() Bucket {
	return Bucket{
		LastReffil: time.Now(),
		Tokens:     p.capacity,
	}
}
