package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Config 持有服务运行所需的全部配置。
type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	CORSOrigins []string

	Database   DatabaseConfig
	DBRequired bool

	Redis  RedisConfig
	Cache  CacheConfig
}

// DatabaseConfig 描述 MySQL 连接参数。
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	Username string
	Password string
}

// DSN 生成 go-sql-driver/mysql 格式的 DSN。
func (d DatabaseConfig) DSN() string {
	cfg := mysql.Config{
		User:                 d.Username,
		Passwd:               d.Password,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(d.Host, d.Port),
		DBName:               d.Name,
		ParseTime:            true,
		Loc:                  time.Local,
		AllowNativePasswords: true,
		Params:               map[string]string{"charset": "utf8mb4"},
	}
	return cfg.FormatDSN()
}

// RedisConfig 描述 Redis 连接参数。
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Enabled  bool
}

// CacheConfig 描述缓存行为。
type CacheConfig struct {
	TTL time.Duration
}

// Load 从 .env 文件和环境变量加载配置。
func Load() (*Config, error) {
	loadDotEnv(".env")

	cfg := &Config{
		Addr:         getEnv("KG_ADDR", ":10171"),
		ReadTimeout:  getDurationEnv("KG_READ_TIMEOUT", 15*time.Second),
		WriteTimeout: getDurationEnv("KG_WRITE_TIMEOUT", 30*time.Second),
		CORSOrigins:  getCSVEnv("KG_CORS_ORIGINS", []string{"*"}),
		DBRequired:   getBoolEnv("KG_DB_REQUIRED", false),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "3306"),
			Name:     getEnv("DB_NAME", "ptadatabase"),
			Username: getEnv("DB_USERNAME", "root"),
			Password: firstNonEmpty(os.Getenv("DB_PASSWORD"), os.Getenv("DB_PASS")),
		},
		Redis: RedisConfig{
			Addr:     getEnv("KG_REDIS_ADDR", "127.0.0.1:6379"),
			Password: firstNonEmpty(os.Getenv("KG_REDIS_PASSWORD"), os.Getenv("REDIS_PASSWORD")),
			DB:       getIntEnv("KG_REDIS_DB", 0),
			Enabled:  getBoolEnv("KG_REDIS_ENABLED", false),
		},
		Cache: CacheConfig{
			TTL: getDurationEnv("KG_CACHE_TTL", 300*time.Second),
		},
	}

	return cfg, nil
}

// ---- .env 解析（手写，不依赖第三方库，与 LeetCodeClaw 一致）----

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return // 文件不存在视为正常
	}
	lines := strings.Split(string(data), "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if !validEnvKey(key) {
			continue
		}
		value = parseEnvValue(strings.TrimSpace(value))
		// 不覆盖已存在的环境变量
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func validEnvKey(key string) bool {
	if key == "" {
		return false
	}
	for i, c := range key {
		if i == 0 {
			if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_') {
				return false
			}
		} else {
			if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
				return false
			}
		}
	}
	return true
}

func parseEnvValue(value string) string {
	if len(value) >= 2 {
		if value[0] == '"' && value[len(value)-1] == '"' {
			if unquoted, err := strconv.Unquote(value); err == nil {
				return unquoted
			}
			return value[1 : len(value)-1]
		}
		if value[0] == '\'' && value[len(value)-1] == '\'' {
			return value[1 : len(value)-1]
		}
	}
	return value
}

// ---- 环境变量辅助函数 ----

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("config: invalid duration for %s=%q, using default %v", key, v, fallback)
		return fallback
	}
	return d
}

func getIntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getBoolEnv(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getCSVEnv(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// String 返回配置的可读摘要（不含密码）。
func (c *Config) String() string {
	return fmt.Sprintf("Config{Addr:%s, DB:%s:%s/%s, DBRequired:%v, Redis:%s(enabled:%v), CacheTTL:%v, CORS:%v}",
		c.Addr, c.Database.Host, c.Database.Port, c.Database.Name,
		c.DBRequired, c.Redis.Addr, c.Redis.Enabled, c.Cache.TTL, c.CORSOrigins)
}
