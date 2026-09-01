// Package config 加载服务运行配置(环境变量优先,提供默认值)
package config

import (
	"os"
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
}

func Load() *Config {
	return &Config{
		Port:      getenv("STUDYROOM_PORT", "8080"),
		DBUrl:     getenv("STUDYROOM_DB_URL", "host=/tmp port=5432 dbname=studyroom"),
		JWTSecret: getenv("STUDYROOM_JWT_SECRET", "studyroom-dev-secret"),
		TokenTTL:  24 * time.Hour,
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
