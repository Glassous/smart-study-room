package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRateLimitIsAtomicAndExpires(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	ctx := context.Background()
	key := "ratelimit:test"

	if !checkRateLimit(ctx, rdb, key, 2, time.Minute) || !checkRateLimit(ctx, rdb, key, 2, time.Minute) {
		t.Fatal("first two requests should pass")
	}
	if checkRateLimit(ctx, rdb, key, 2, time.Minute) {
		t.Fatal("third request should be limited")
	}
	if ttl := mr.TTL(key); ttl <= 0 {
		t.Fatalf("counter must have TTL, got %v", ttl)
	}
	mr.FastForward(time.Minute)
	if !checkRateLimit(ctx, rdb, key, 2, time.Minute) {
		t.Fatal("request should pass after window expires")
	}
}

func TestRateLimitDegradesOnRedisFailure(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	mr.Close()
	if !checkRateLimit(context.Background(), rdb, "key", 1, time.Minute) {
		t.Fatal("Redis failure should degrade to allow")
	}
}
