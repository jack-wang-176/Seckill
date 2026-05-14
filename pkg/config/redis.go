package config

import (
	"strconv"
)

// ============================================================
// Redis 配置
// ============================================================

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Redis 默认配置常量
const (
	DefaultRedisAddr     = "127.0.0.1:6379"
	DefaultRedisPassword = ""
	DefaultRedisDB       = 0
)

// LoadRedisConfig 从环境变量加载 Redis 配置
func LoadRedisConfig() RedisConfig {
	db, _ := strconv.Atoi(getEnvOrDefault("REDIS_DB", strconv.Itoa(DefaultRedisDB)))

	return RedisConfig{
		Addr:     getEnvOrDefault("REDIS_ADDR", DefaultRedisAddr),
		Password: getEnvOrDefault("REDIS_PASSWORD", DefaultRedisPassword),
		DB:       db,
	}
}
