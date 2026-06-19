package temporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
)

// QuickRetry is for fast, idempotent operations (health checks, lookups).
var QuickRetry = &temporal.RetryPolicy{
	InitialInterval:    time.Second,
	BackoffCoefficient: 2.0,
	MaximumInterval:    5 * time.Second,
	MaximumAttempts:    3,
}

// StandardRetry is the default for most activities.
var StandardRetry = &temporal.RetryPolicy{
	InitialInterval:    5 * time.Second,
	BackoffCoefficient: 2.0,
	MaximumInterval:    30 * time.Second,
	MaximumAttempts:    3,
}

// LongRetry is for slow operations (environment provisioning, builds).
var LongRetry = &temporal.RetryPolicy{
	InitialInterval:    30 * time.Second,
	BackoffCoefficient: 2.0,
	MaximumInterval:    5 * time.Minute,
	MaximumAttempts:    5,
}

// ExternalRetry is for calls to third-party APIs with rate limits.
// Non-retryable on 4xx (client errors).
var ExternalRetry = &temporal.RetryPolicy{
	InitialInterval:    10 * time.Second,
	BackoffCoefficient: 2.0,
	MaximumInterval:    2 * time.Minute,
	MaximumAttempts:    5,
	NonRetryableErrorTypes: []string{
		"BadRequest",
		"Unauthorized",
		"Forbidden",
		"NotFound",
		"Conflict",
	},
}

// NoRetry disables retries entirely.
var NoRetry = &temporal.RetryPolicy{
	MaximumAttempts: 1,
}
