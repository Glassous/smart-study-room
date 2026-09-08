// Package config 加载服务运行配置(环境变量优先,提供默认值)
package config

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
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
	RedisDB           int
	AIBaseURL         string
	AIAPIKey          string
	AIModel           string
	AIConnectTimeout  time.Duration
	AIResponseTimeout time.Duration
	AIContextMessages int
}

func Load() *Config {
	// 支持从仓库根目录执行 go run ./backend/cmd/server，也支持先 cd backend 再启动。
	// backend/.env 优先于根目录 .env；已注入的系统环境变量始终具有最高优先级。
	loadDotEnv("backend/.env")
	loadDotEnv(".env")
	return &Config{
		Port:              getenv("STUDYROOM_PORT", "8080"),
		DBUrl:             getenv("STUDYROOM_DB_URL", "host=/tmp port=5432 dbname=studyroom"),
		JWTSecret:         getenv("STUDYROOM_JWT_SECRET", "studyroom-dev-secret"),
		TokenTTL:          24 * time.Hour,
		RedisAddr:         getenv("STUDYROOM_REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:     getenv("STUDYROOM_REDIS_PASSWORD", "studyroom_redis_pass"),
		RedisDB:           getenvInt("STUDYROOM_REDIS_DB", 0),
		AIBaseURL:         getenv("STUDYROOM_AI_BASE_URL", ""),
		AIAPIKey:          getenv("STUDYROOM_AI_API_KEY", ""),
		AIModel:           getenv("STUDYROOM_AI_MODEL", "gpt-4o-mini"),
		AIConnectTimeout:  time.Duration(getenvInt("STUDYROOM_AI_CONNECT_TIMEOUT_SECONDS", 10)) * time.Second,
		AIResponseTimeout: time.Duration(getenvInt("STUDYROOM_AI_RESPONSE_TIMEOUT_SECONDS", 60)) * time.Second,
		AIContextMessages: getenvInt("STUDYROOM_AI_CONTEXT_MESSAGES", 20),
	}
}

// loadDotEnv 读取后端目录的本地 .env。已有系统环境变量优先，避免覆盖 Docker/部署配置。
// 仅支持 KEY=VALUE、空行和 # 注释，满足本项目的运行配置需要且不引入额外依赖。
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" || os.Getenv(key) != "" {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), "\"'")
		if err := os.Setenv(key, value); err != nil {
			log.Printf("[config] 无法加载 .env 中的 %s", key)
		}
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
