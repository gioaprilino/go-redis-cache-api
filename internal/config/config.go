package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort  string
	RedisAddr   string
	RedisPass   string
	DatabaseURL string
	CacheTTL    int
}

func Load() *Config {
	ttl, _ := strconv.Atoi(getEnv("CACHE_TTL", "60"))
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASS", ""),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/cache_demo?sslmode=disable"),
		CacheTTL:    ttl,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
