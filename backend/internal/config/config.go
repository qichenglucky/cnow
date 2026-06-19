package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	Env      string
	HTTPAddr string

	DB       DBConfig
	Temporal TemporalConfig
	Log      LogConfig
	CORS     CORSConfig
}

// DBConfig holds database connection parameters.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	MaxConns int32
}

// TemporalConfig holds Temporal client configuration.
type TemporalConfig struct {
	HostPort  string
	Namespace string
	TaskQueue string
}

// LogConfig holds logger configuration.
type LogConfig struct {
	Level  string
	Format string
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowedOrigins []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	maxConns := int32(20)
	if v := os.Getenv("CNOW_DB_MAX_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxConns = int32(n)
		}
	}

	origins := "*"
	if v := os.Getenv("CNOW_CORS_ALLOWED_ORIGINS"); v != "" {
		origins = v
	}

	return &Config{
		Env:      envOrDefault("CNOW_ENV", "dev"),
		HTTPAddr: envOrDefault("CNOW_HTTP_ADDR", ":8080"),

		DB: DBConfig{
			Host:     envOrDefault("CNOW_DB_HOST", "localhost"),
			Port:     envOrDefault("CNOW_DB_PORT", "5432"),
			User:     envOrDefault("CNOW_DB_USER", "cnow"),
			Password: envOrDefault("CNOW_DB_PASSWORD", "cnow"),
			Name:     envOrDefault("CNOW_DB_NAME", "cnow"),
			MaxConns: maxConns,
		},

		Temporal: TemporalConfig{
			HostPort:  envOrDefault("CNOW_TEMPORAL_HOST_PORT", "localhost:7233"),
			Namespace: envOrDefault("CNOW_TEMPORAL_NAMESPACE", "default"),
			TaskQueue: envOrDefault("CNOW_TEMPORAL_TASK_QUEUE", "cnow-main"),
		},

		Log: LogConfig{
			Level:  envOrDefault("CNOW_LOG_LEVEL", "info"),
			Format: envOrDefault("CNOW_LOG_FORMAT", "json"),
		},

		CORS: CORSConfig{
			AllowedOrigins: strings.Split(origins, ","),
		},
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
