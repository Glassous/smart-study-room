// Package rediscache 提供通用的 Redis 缓存辅助、序列化及防雪崩机制
package rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

// Helper 缓存操作辅助对象
type Helper struct {
	rdb *redis.Client
}

const statsPrefix = "studyroom:cache:stats:"

// NewHelper 实例化缓存辅助器
func NewHelper(rdb *redis.Client) *Helper {
	return &Helper{rdb: rdb}
}

// Client 获取底层 Redis Client
func (h *Helper) Client() *redis.Client {
	return h.rdb
}

// WithJitter 为指定 TTL 增加 ±10% 的随机抖动，杜绝缓存同时大面积过期导致的雪崩
func WithJitter(baseTTL time.Duration) time.Duration {
	if baseTTL <= 0 {
		return baseTTL
	}
	// 浮动范围 [-10%, +10%]
	deltaPercent := (rand.Float64()*0.2 - 0.1) // -0.1 ~ +0.1
	jitter := time.Duration(float64(baseTTL) * deltaPercent)
	return baseTTL + jitter
}

// Get 从 Redis 获取缓存并反序列化至 dest (指针)
func (h *Helper) Get(ctx context.Context, key string, dest any) (bool, error) {
	if h == nil || h.rdb == nil {
		return false, nil
	}
	data, err := h.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // 缓存未命中
		}
		log.Printf("[cache] 读取失败 key=%s: %v", key, err)
		return false, err // 网络或集群异常
	}
	if err := json.Unmarshal(data, dest); err != nil {
		log.Printf("[cache] 反序列化失败 key=%s: %v", key, err)
		return false, err
	}
	return true, nil
}

// Set 将对象 JSON 序列化并存入 Redis，带 TTL 抖动
func (h *Helper) Set(ctx context.Context, key string, value any, baseTTL time.Duration) error {
	if h == nil || h.rdb == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = h.rdb.Set(ctx, key, data, WithJitter(baseTTL)).Err()
	if err != nil {
		log.Printf("[cache] 写入失败 key=%s: %v", key, err)
	}
	return err
}

// Del 删除指定 key
func (h *Helper) Del(ctx context.Context, keys ...string) error {
	if h == nil || h.rdb == nil || len(keys) == 0 {
		return nil
	}
	err := h.rdb.Del(ctx, keys...).Err()
	if err != nil {
		log.Printf("[cache] 删除失败 keys=%v: %v", keys, err)
	}
	return err
}

// DelPrefix 按前缀匹配删除缓存（使用 Scan 避免阻塞单线程）
func (h *Helper) DelPrefix(ctx context.Context, prefix string) error {
	if h == nil || h.rdb == nil || prefix == "" {
		return nil
	}
	iter := h.rdb.Scan(ctx, 0, prefix+"*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			if err := h.Del(ctx, keys...); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if len(keys) > 0 {
		if err := h.Del(ctx, keys...); err != nil {
			return err
		}
	}
	err := iter.Err()
	if err != nil {
		log.Printf("[cache] 扫描失败 prefix=%s: %v", prefix, err)
	}
	return err
}

// InvalidateStats 清理所有统计缓存；失败仅记录并返回，不影响主业务。
func (h *Helper) InvalidateStats(ctx context.Context) error {
	return h.DelPrefix(ctx, statsPrefix)
}
