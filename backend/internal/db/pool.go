package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	MaxConns int32
}

func ConfigFromEnv() Config {
	return Config{
		Host:     envOrDefault("CNOW_DB_HOST", "localhost"),
		Port:     envOrDefault("CNOW_DB_PORT", "5432"),
		User:     envOrDefault("CNOW_DB_USER", "cnow"),
		Password: envOrDefault("CNOW_DB_PASSWORD", "cnow"),
		Name:     envOrDefault("CNOW_DB_NAME", "cnow"),
		MaxConns: 20,
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (c Config) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

func NewPool(ctx context.Context, cfg Config, log *zap.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.ConnString())
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	log.Info("database connected", zap.String("host", cfg.Host), zap.String("port", cfg.Port), zap.String("database", cfg.Name))
	return pool, nil
}

func Health(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
