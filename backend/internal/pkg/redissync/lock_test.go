package redissync

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestMutexNilClientDegradation(t *testing.T) {
	ctx := context.Background()
	mutex := NewMutex(nil, "lock:test:1", 5*time.Second)

	if mutex.Key() != "lock:test:1" {
		t.Fatalf("expected key lock:test:1, got %s", mutex.Key())
	}
	locked, err := mutex.TryLock(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !locked {
		t.Fatalf("expected locked true for nil client degradation")
	}

	err = mutex.Unlock(ctx)
	if err != nil {
		t.Fatalf("unexpected error on unlock: %v", err)
	}
}

func TestMutexCompetitionRenewAndOwnership(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	ctx := context.Background()
	first := NewMutex(rdb, "lock:test", 5*time.Second)
	second := NewMutex(rdb, "lock:test", 5*time.Second)

	if ok, err := first.TryLock(ctx); err != nil || !ok {
		t.Fatalf("first lock failed: ok=%v err=%v", ok, err)
	}
	if ok, err := second.TryLock(ctx); err != nil || ok {
		t.Fatalf("second lock should be rejected: ok=%v err=%v", ok, err)
	}
	mr.FastForward(4 * time.Second)
	if ok, err := first.Renew(ctx); err != nil || !ok {
		t.Fatalf("renew failed: ok=%v err=%v", ok, err)
	}
	if ttl := mr.TTL("lock:test"); ttl < 4*time.Second {
		t.Fatalf("renew did not restore TTL: %v", ttl)
	}
	if err := second.Unlock(ctx); err != nil {
		t.Fatal(err)
	}
	if !mr.Exists("lock:test") {
		t.Fatal("non-owner removed lock")
	}
	if err := first.Unlock(ctx); err != nil {
		t.Fatal(err)
	}
	if mr.Exists("lock:test") {
		t.Fatal("owner failed to remove lock")
	}
}

func TestMutexAutoRenew(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mutex := NewMutex(rdb, "lock:auto-renew", 100*time.Millisecond)
	if ok, err := mutex.TryLock(ctx); err != nil || !ok {
		t.Fatalf("lock failed: ok=%v err=%v", ok, err)
	}
	stop := mutex.StartAutoRenew(ctx, 10*time.Millisecond, func(err error) {
		t.Errorf("unexpected renew error: %v", err)
	})
	defer stop()
	time.Sleep(20 * time.Millisecond)
	mr.FastForward(90 * time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	if ttl := mr.TTL("lock:auto-renew"); ttl <= 0 {
		t.Fatalf("auto-renew did not keep lock alive: %v", ttl)
	}
}
