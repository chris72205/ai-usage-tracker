package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port             string
	ServiceName      string // used as Redis key prefix to avoid collisions on shared instances
	RedisURL         string
	DedupWindow      time.Duration // captures within this window are discarded after the first
	RabbitMQURL      string
	RabbitMQExchange string
	InfluxURL        string
	InfluxToken      string
	InfluxOrg        string
	InfluxBucket     string
}

func Load() Config {
	return Config{
		Port:             getEnv("PORT", "8080"),
		ServiceName:      getEnv("SERVICE_NAME", "ai-usage-svc"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379/0"),
		DedupWindow:      time.Duration(getEnvInt("DEDUP_WINDOW_SECONDS", 30)) * time.Second,
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "usage"),
		InfluxURL:        getEnv("INFLUX_URL", "http://localhost:8086"),
		InfluxToken:      getEnv("INFLUX_TOKEN", ""),
		InfluxOrg:        getEnv("INFLUX_ORG", ""),
		InfluxBucket:     getEnv("INFLUX_BUCKET", "ai-agent-usages"),
	}
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
