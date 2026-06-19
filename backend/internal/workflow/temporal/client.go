package temporal

import (
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

// TemporalConfig holds connection parameters.
type TemporalConfig struct {
	HostPort  string
	Namespace string
	TaskQueue string
}

// LoadTemporalConfig reads Temporal connection settings from env with defaults.
func LoadTemporalConfig() TemporalConfig {
	cfg := TemporalConfig{
		HostPort:  os.Getenv("CNOW_TEMPORAL_HOSTPORT"),
		Namespace: os.Getenv("CNOW_TEMPORAL_NAMESPACE"),
		TaskQueue: os.Getenv("CNOW_TEMPORAL_TASK_QUEUE"),
	}
	if cfg.HostPort == "" {
		cfg.HostPort = "localhost:7233"
	}
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}
	if cfg.TaskQueue == "" {
		cfg.TaskQueue = DefaultTaskQueue
	}
	return cfg
}

// NewTemporalClient creates a Temporal client from the given configuration.
func NewTemporalClient(cfg TemporalConfig, logger *zap.Logger) (client.Client, error) {
	opts := client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
		Logger:    newTemporalZapAdapter(logger),
	}

	c, err := client.Dial(opts)
	if err != nil {
		return nil, fmt.Errorf("temporal: dial %s/%s: %w", cfg.HostPort, cfg.Namespace, err)
	}

	logger.Info("temporal client connected",
		zap.String("hostPort", cfg.HostPort),
		zap.String("namespace", cfg.Namespace),
	)
	return c, nil
}

// CloseTemporalClient closes the Temporal client connection.
func CloseTemporalClient(c client.Client) {
	if c != nil {
		c.Close()
	}
}

// temporalZapAdapter adapts zap.Logger to the Temporal log.Logger interface.
type temporalZapAdapter struct {
	l *zap.Logger
}

func newTemporalZapAdapter(l *zap.Logger) *temporalZapAdapter {
	return &temporalZapAdapter{l: l.Named("temporal")}
}

func (a *temporalZapAdapter) Debug(msg string, keyvals ...interface{}) {
	a.l.Sugar().Debugw(msg, keyvals...)
}

func (a *temporalZapAdapter) Info(msg string, keyvals ...interface{}) {
	a.l.Sugar().Infow(msg, keyvals...)
}

func (a *temporalZapAdapter) Warn(msg string, keyvals ...interface{}) {
	a.l.Sugar().Warnw(msg, keyvals...)
}

func (a *temporalZapAdapter) Error(msg string, keyvals ...interface{}) {
	a.l.Sugar().Errorw(msg, keyvals...)
}
