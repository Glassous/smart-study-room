// Package redissync 基于 Redis 的分布式锁与并发互斥支持
package redissync

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrLockAcquireFailed = errors.New("获取分布式锁失败，资源正在被占用")
	ErrLockNilClient     = errors.New("redis 客户端未配置")
)

// releaseLuaScript 安全释放锁的 Lua 脚本，原子比对 token 并删除
const releaseLuaScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`

const renewLuaScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("pexpire", KEYS[1], ARGV[2])
else
    return 0
end
`

// Mutex Redis 分布式互斥锁
type Mutex struct {
	rdb   *redis.Client
	key   string
	token string
	ttl   time.Duration
}

// NewMutex 构建一个互斥锁对象
func NewMutex(rdb *redis.Client, key string, ttl time.Duration) *Mutex {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	token := hex.EncodeToString(buf)
	return &Mutex{
		rdb:   rdb,
		key:   key,
		token: token,
		ttl:   ttl,
	}
}

// TryLock 尝试获取锁，非阻塞模式；若获取成功返回 true，若被他人持有返回 false
func (m *Mutex) TryLock(ctx context.Context) (bool, error) {
	if m.rdb == nil {
		// 客户端为空时视为降级通过（由下层 DB 约束兜底）
		return true, nil
	}
	success, err := m.rdb.SetNX(ctx, m.key, m.token, m.ttl).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

// Unlock 安全释放锁
func (m *Mutex) Unlock(ctx context.Context) error {
	if m.rdb == nil {
		return nil
	}
	res, err := m.rdb.Eval(ctx, releaseLuaScript, []string{m.key}, m.token).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	_ = res
	return nil
}

// Renew 仅在当前实例仍持有锁时原子续租。
func (m *Mutex) Renew(ctx context.Context) (bool, error) {
	if m.rdb == nil {
		return true, nil
	}
	res, err := m.rdb.Eval(ctx, renewLuaScript, []string{m.key}, m.token, m.ttl.Milliseconds()).Int64()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

// StartAutoRenew 周期续租，返回停止函数；续租失败后停止并调用 onError。
func (m *Mutex) StartAutoRenew(ctx context.Context, interval time.Duration, onError func(error)) func() {
	renewCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-renewCtx.Done():
				return
			case <-ticker.C:
				ok, err := m.Renew(renewCtx)
				if renewCtx.Err() != nil {
					return
				}
				if err == nil && !ok {
					err = ErrLockAcquireFailed
				}
				if err != nil {
					if onError != nil {
						onError(err)
					}
					return
				}
			}
		}
	}()
	return cancel
}

// Key 返回锁键，便于诊断与测试。
func (m *Mutex) Key() string { return m.key }
