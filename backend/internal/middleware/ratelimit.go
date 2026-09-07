// Package middleware 中间件层: 访问限流与安全防护
package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const rateLimitLuaScript = `
local count = redis.call("incr", KEYS[1])
if count == 1 then
    redis.call("pexpire", KEYS[1], ARGV[1])
end
return count
`

// RateLimitByIP 基于客户端 IP 的 Redis 计数器限流中间件
func RateLimitByIP(rdb *redis.Client, action string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}
		ip := c.ClientIP()
		key := fmt.Sprintf("studyroom:ratelimit:ip:%s:%s", action, ip)
		if !checkRateLimit(c.Request.Context(), rdb, key, limit, window) {
			abort(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			return
		}
		c.Next()
	}
}

// RateLimitByUser 基于当前登录用户的 UID 限流
func RateLimitByUser(rdb *redis.Client, action string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}
		uid := CurrentUID(c)
		if uid == 0 {
			c.Next()
			return
		}
		key := fmt.Sprintf("studyroom:ratelimit:user:%d:%s", uid, action)
		if !checkRateLimit(c.Request.Context(), rdb, key, limit, window) {
			abort(c, http.StatusTooManyRequests, "操作过于频繁，请稍后再试")
			return
		}
		c.Next()
	}
}

func checkRateLimit(ctx context.Context, rdb *redis.Client, key string, limit int64, window time.Duration) bool {
	count, err := rdb.Eval(ctx, rateLimitLuaScript, []string{key}, window.Milliseconds()).Int64()
	if err != nil {
		// Redis 异常时降级放行
		log.Printf("[ratelimit] Redis 限流失败，降级放行 key=%s: %v", key, err)
		return true
	}
	return count <= limit
}
