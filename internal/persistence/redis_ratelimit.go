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
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var bucket Bucket
	refillKey := fmt.Sprintf("%s:refill:%s", p.prefix, key)
	tokensKey := fmt.Sprintf("%s:tokens:%s", p.prefix, key)

	pipe := p.rdb.TxPipeline()

	exists, _ := p.rdb.Exists(p.ctx, refillKey).Result()
	if exists == 0 {
		bucket = Bucket{
			LastReffil: time.Now(),
			Tokens:     p.capacity,
		}
	} else {

		rs := p.rdb.Get(p.ctx, refillKey).Val()
		refill, error := time.Parse(time.RFC3339Nano, rs)
		if error != nil {
			return errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
		}
		ts, error := p.rdb.Get(p.ctx, tokensKey).Int()
		if error != nil {
			return errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
		}

		bucket = Bucket{
			LastReffil: refill,
			Tokens:     ts,
		}
	}

	last := time.Since(bucket.LastReffil)
	if last.Milliseconds() > p.period.Milliseconds() {
		bucket = Bucket{
			LastReffil: time.Now(),
			Tokens:     p.capacity,
		}
	}

	if bucket.Tokens < 1 {
		return errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
	}

	bucket.Tokens -= 1

	pipe.Set(p.ctx, refillKey, bucket.LastReffil.Format(time.RFC3339Nano), p.period)
	pipe.Set(p.ctx, tokensKey, bucket.Tokens, p.period)
	_, error := pipe.Exec(p.ctx)
	if error != nil {
		return errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
	}
	return nil
}
