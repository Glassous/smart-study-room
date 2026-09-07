// Package config 加载服务运行配置(环境变量优先,提供默认值)
package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Port HTTP 监听端口
	Port string
	// DBUrl PostgreSQL 连接串(用户级集群默认走 /tmp socket)
	DBUrl string
	// JWTSecret 令牌签名密钥
	JWTSecret string
	// TokenTTL 令牌有效期
	TokenTTL time.Duration
	// RedisAddr Redis 服务地址
	RedisAddr string
	// RedisPassword Redis 访问密码
	RedisPassword string
	// RedisDB Redis 数据库号
	RedisDB int
}

func Load() *Config {
	return &Config{
		Port:          getenv("STUDYROOM_PORT", "8080"),
		DBUrl:         getenv("STUDYROOM_DB_URL", "host=/tmp port=5432 dbname=studyroom"),
		JWTSecret:     getenv("STUDYROOM_JWT_SECRET", "studyroom-dev-secret"),
		TokenTTL:      24 * time.Hour,
		RedisAddr:     getenv("STUDYROOM_REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: getenv("STUDYROOM_REDIS_PASSWORD", "studyroom_redis_pass"),
		RedisDB:       getenvInt("STUDYROOM_REDIS_DB", 0),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
		log.Printf("[config] %s=%q 非法，使用默认值 %d", key, v, def)
	}
	return def
}
