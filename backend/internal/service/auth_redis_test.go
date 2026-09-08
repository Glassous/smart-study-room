package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"github.com/imicola/smart-study-room/backend/internal/config"
	"github.com/imicola/smart-study-room/backend/internal/model"
)

func newRedisTestAuth(t *testing.T) (*AuthService, *miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewAuthService(nil, &config.Config{JWTSecret: "test-secret", TokenTTL: time.Hour}, rdb), mr, rdb
}

func signedTestToken(t *testing.T, auth *AuthService, expires time.Time) string {
	t.Helper()
	claims := &model.Claims{RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expires)}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(auth.cfg.JWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestLogoutBlacklistsTokenUntilExpiry(t *testing.T) {
	auth, mr, _ := newRedisTestAuth(t)
	token := signedTestToken(t, auth, time.Now().Add(time.Hour))
	if err := auth.Logout(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	blocked, err := auth.IsTokenBlacklisted(context.Background(), token)
	if err != nil || !blocked {
		t.Fatalf("blocked=%v err=%v", blocked, err)
	}
	ttl := mr.TTL("studyroom:auth:blacklist:" + TokenHash(token))
	if ttl <= 0 || ttl > time.Hour {
		t.Fatalf("unexpected blacklist TTL: %v", ttl)
	}
}

func TestLogoutRequiresRevocationStore(t *testing.T) {
	auth := NewAuthService(nil, &config.Config{JWTSecret: "test-secret"}, nil)
	err := auth.Logout(context.Background(), "token")
	if !errors.Is(err, ErrRevocationUnavailable) {
		t.Fatalf("expected ErrRevocationUnavailable, got %v", err)
	}
}

func TestBlacklistReadFailureIsReported(t *testing.T) {
	auth, mr, _ := newRedisTestAuth(t)
	mr.Close()
	if _, err := auth.IsTokenBlacklisted(context.Background(), "token"); err == nil {
		t.Fatal("expected Redis read error")
	}
}
