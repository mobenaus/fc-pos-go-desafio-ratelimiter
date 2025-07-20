package ratelimit_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/persistence"
	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/ratelimit"
)

type mockPersistence struct {
	bucket       *persistence.Bucket
	getErr       error
	saveErr      error
	refilled     bool
	shouldrefill bool
}

func (m *mockPersistence) GetBucket(key string) (*persistence.Bucket, error) {
	return m.bucket, m.getErr
}
func (m *mockPersistence) SaveBucket(key string, bucket *persistence.Bucket) error {
	m.bucket = bucket
	return m.saveErr
}
func (m *mockPersistence) CheckRefill(bucket *persistence.Bucket) bool {
	return m.shouldrefill
}
func (m *mockPersistence) Refill(bucket *persistence.Bucket) {
	m.refilled = true
	bucket.Tokens = 5
}

func TestUseToken_Success(t *testing.T) {
	bucket := &persistence.Bucket{Tokens: 2, LastReffil: time.Now()}
	mock := &mockPersistence{bucket: bucket, shouldrefill: false}
	rl := ratelimit.NewRateLimit(mock)
	err := rl.UseToken("key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.bucket.Tokens != 1 {
		t.Errorf("expected 1 token left, got %d", mock.bucket.Tokens)
	}
}

func TestUseToken_Refill(t *testing.T) {
	bucket := &persistence.Bucket{Tokens: 0, LastReffil: time.Now()}
	mock := &mockPersistence{bucket: bucket, shouldrefill: true}
	rl := ratelimit.NewRateLimit(mock)
	err := rl.UseToken("key")
	if err != nil {
		t.Fatalf("expected no error after refill, got %v", err)
	}
	if !mock.refilled {
		t.Error("expected bucket to be refilled")
	}
}

func TestUseToken_GetBucketError(t *testing.T) {
	mock := &mockPersistence{getErr: errors.New("fail")}
	rl := ratelimit.NewRateLimit(mock)
	err := rl.UseToken("key")
	if err == nil {
		t.Error("expected error when GetBucket fails")
	}
}

func TestUseToken_SaveBucketError(t *testing.T) {
	bucket := &persistence.Bucket{Tokens: 2, LastReffil: time.Now()}
	mock := &mockPersistence{bucket: bucket, saveErr: errors.New("fail")}
	rl := ratelimit.NewRateLimit(mock)
	err := rl.UseToken("key")
	if err == nil {
		t.Error("expected error when SaveBucket fails")
	}
}

func TestUseToken_NoTokens(t *testing.T) {
	bucket := &persistence.Bucket{Tokens: 0, LastReffil: time.Now()}
	mock := &mockPersistence{bucket: bucket}
	rl := ratelimit.NewRateLimit(mock)
	err := rl.UseToken("key")
	if err == nil {
		t.Error("expected error when no tokens left")
	}
}

func TestRateLimitFiveRequestsMemoryPersistence(t *testing.T) {
	persistence := persistence.NewMemoryRateLimitPersistence(5, time.Second)
	rl := ratelimit.NewRateLimit(persistence)
	countsuccess := 0
	counterrors := 0
	for range 10 {
		errors := rl.UseToken("key")
		if errors != nil {
			counterrors++
		} else {
			countsuccess++
		}
	}
	if countsuccess != 5 {
		t.Error("expected 5 success count")
	}
	if counterrors != 5 {
		t.Error("expected 5 error count")
	}

}
