package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimitPersistence struct {
	ctx      context.Context
	rdb      *redis.Client
	prefix   string
	capacity int
	period   time.Duration
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

func (p *RedisRateLimitPersistence) Refill(bucket *Bucket) {
	bucket.LastReffil = time.Now()
	bucket.Tokens = p.capacity
}

func (p *RedisRateLimitPersistence) GetBucket(key string) (*Bucket, error) {
	var bucket Bucket
	refillKey := fmt.Sprintf("%s:refill:%s", p.prefix, key)
	tokensKey := fmt.Sprintf("%s:tokens:%s", p.prefix, key)

	exists, error := p.rdb.Exists(p.ctx, refillKey).Result()
	if error != nil {
		return nil, error
	}
	if exists == 0 {
		bucket = Bucket{
			LastReffil: time.Now(),
			Tokens:     p.capacity,
		}
	} else {
		bucket, error = p.fill(refillKey, tokensKey)
		if error != nil {
			return nil, error
		}
	}

	return &bucket, nil
}

func (p *RedisRateLimitPersistence) CheckRefill(bucket *Bucket) bool {
	last := time.Since(bucket.LastReffil)
	return last.Milliseconds() > p.period.Milliseconds()
}

func (p *RedisRateLimitPersistence) SaveBucket(key string, bucket *Bucket) error {

	refillKey := fmt.Sprintf("%s:refill:%s", p.prefix, key)
	tokensKey := fmt.Sprintf("%s:tokens:%s", p.prefix, key)

	pipe := p.rdb.TxPipeline()

	pipe.Set(p.ctx, refillKey, bucket.LastReffil.Format(time.RFC3339Nano), p.period)
	pipe.Set(p.ctx, tokensKey, bucket.Tokens, p.period)

	_, error := pipe.Exec(p.ctx)

	if error != nil {
		return error
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
