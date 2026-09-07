// Package service 业务逻辑层
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/imicola/smart-study-room/backend/internal/config"
	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
)

// 认证相关业务错误
var (
	ErrUsernameTaken         = errors.New("用户名已被注册")
	ErrBadCredentials        = errors.New("用户名或口令错误")
	ErrUserDisabled          = errors.New("账号已被禁用")
	ErrRevocationUnavailable = errors.New("令牌注销服务暂时不可用")
)

// AuthService 认证服务
type AuthService struct {
	users *repository.UserRepo
	cfg   *config.Config
	rdb   *redis.Client
}

func NewAuthService(users *repository.UserRepo, cfg *config.Config, rdb *redis.Client) *AuthService {
	return &AuthService{users: users, cfg: cfg, rdb: rdb}
}

// Register 注册新用户(默认 student 角色, 信用分 100)
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.UserPublic, error) {
	taken, err := s.users.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("口令哈希失败: %w", err)
	}

	u := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		RealName:     req.RealName,
		Role:         model.RoleStudent,
	}
	// 学号可空, 非空时转指针
	if req.StudentNo != "" {
		u.StudentNo = &req.StudentNo
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u.ToPublic(), nil
}

// Login 校验口令并签发 JWT
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	u, err := s.users.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrBadCredentials
		}
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		return nil, ErrBadCredentials
	}
	if u.Status != model.UserActive {
		return nil, ErrUserDisabled
	}

	now := time.Now()
	expiresAt := now.Add(s.cfg.TokenTTL)
	claims := model.NewClaims(u, now, expiresAt)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("令牌签发失败: %w", err)
	}

	return &model.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      u.ToPublic(),
	}, nil
}

// Profile 查询当前用户信息
func (s *AuthService) Profile(ctx context.Context, uid int64) (*model.UserPublic, error) {
	u, err := s.users.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return u.ToPublic(), nil
}

// ParseToken 校验并解析 JWT 载荷
func (s *AuthService) ParseToken(tokenStr string) (*model.Claims, error) {
	claims := &model.Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名算法: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// TokenHash 计算令牌的 SHA-256 摘要作为 Redis 键的一部分
func TokenHash(tokenStr string) string {
	sum := sha256.Sum256([]byte(tokenStr))
	return hex.EncodeToString(sum[:])
}

// Logout 用户登出: 将 Token 加入 Redis 黑名单直至其自然过期
func (s *AuthService) Logout(ctx context.Context, tokenStr string) error {
	if tokenStr == "" {
		return nil
	}
	if s.rdb == nil {
		return ErrRevocationUnavailable
	}
	claims, err := s.ParseToken(tokenStr)
	if err != nil {
		// 已经无效的 token 直接返回
		return nil
	}
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining <= 0 {
		return nil
	}
	key := "studyroom:auth:blacklist:" + TokenHash(tokenStr)
	if err := s.rdb.Set(ctx, key, "1", remaining).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrRevocationUnavailable, err)
	}
	return nil
}

// IsTokenBlacklisted 检查令牌是否在注销黑名单中
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, tokenStr string) (bool, error) {
	if s.rdb == nil || tokenStr == "" {
		return false, nil
	}
	key := "studyroom:auth:blacklist:" + TokenHash(tokenStr)
	res, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}
