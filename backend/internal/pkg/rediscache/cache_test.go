package rediscache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestWithJitter(t *testing.T) {
	base := 10 * time.Minute
	for i := 0; i < 20; i++ {
		j := WithJitter(base)
		if j < 9*time.Minute || j > 11*time.Minute {
			t.Fatalf("jittered TTL %v out of range [9m, 11m]", j)
		}
	}
}

func TestCacheRoundTripCorruptionAndStatsInvalidation(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	h := NewHelper(rdb)
	ctx := context.Background()

	if err := h.Set(ctx, "cache:test", map[string]int{"value": 7}, time.Minute); err != nil {
		t.Fatal(err)
	}
	var got map[string]int
	if hit, err := h.Get(ctx, "cache:test", &got); err != nil || !hit || got["value"] != 7 {
		t.Fatalf("hit=%v got=%v err=%v", hit, got, err)
	}
	mr.Set("cache:broken", "not-json")
	if hit, err := h.Get(ctx, "cache:broken", &got); err == nil || hit {
		t.Fatalf("corrupt cache should miss with error: hit=%v err=%v", hit, err)
	}
	mr.Set("studyroom:cache:stats:overview", "value")
	mr.Set("studyroom:cache:rooms", "value")
	if err := h.InvalidateStats(ctx); err != nil {
		t.Fatal(err)
	}
	if mr.Exists("studyroom:cache:stats:overview") || !mr.Exists("studyroom:cache:rooms") {
		t.Fatal("stats invalidation removed wrong keys")
	}
}

func TestNilHelperDegradation(t *testing.T) {
	var h *Helper
	ctx := context.Background()

	var val string
	hit, err := h.Get(ctx, "key", &val)
	if hit || err != nil {
		t.Fatalf("expected false, nil on nil helper, got hit=%v, err=%v", hit, err)
	}

	if err := h.Set(ctx, "key", "val", time.Minute); err != nil {
		t.Fatalf("expected nil error on nil helper Set, got %v", err)
	}

	if err := h.Del(ctx, "key"); err != nil {
		t.Fatalf("expected nil error on nil helper Del, got %v", err)
	}
}
