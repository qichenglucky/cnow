package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const ensureSQL = `
CREATE TABLE IF NOT EXISTS idempotency_key (
    key          TEXT PRIMARY KEY,
    request_hash TEXT NOT NULL,
    response_json JSONB,
    status       TEXT NOT NULL DEFAULT 'reserved',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_idempotency_expires ON idempotency_key(expires_at);
`

// EnsureTable creates the idempotency table if it doesn't exist.
func EnsureTable(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, ensureSQL)
	return err
}

// HashPayload computes SHA-256 hex of the request body.
func HashPayload(payload []byte) string {
	h := sha256.Sum256(payload)
	return hex.EncodeToString(h[:])
}

// CheckAndReserve tries to reserve a key. Returns:
//   - reserved=true  → caller should proceed
//   - reserved=false, response!="" → return cached response
//   - reserved=false, response="" → request in progress (concurrent)
func CheckAndReserve(ctx context.Context, db *pgxpool.Pool, key, requestHash string) (reserved bool, cachedResponse string, err error) {
	var status, existingHash string
	var response *string

	err = db.QueryRow(ctx,
		`SELECT status, request_hash, response_json FROM idempotency_key WHERE key = $1 AND expires_at > NOW()`,
		key,
	).Scan(&status, &existingHash, &response)

	if err != nil {
		// Key not found — reserve it
		_, insertErr := db.Exec(ctx,
			`INSERT INTO idempotency_key (key, request_hash, status, expires_at) VALUES ($1, $2, 'reserved', $3) ON CONFLICT (key) DO NOTHING`,
			key, requestHash, time.Now().Add(24*time.Hour),
		)
		if insertErr != nil {
			return false, "", insertErr
		}
		return true, "", nil
	}

	if existingHash != requestHash {
		return false, "", fmt.Errorf("idempotency key conflict: different request payload for key %s", key)
	}

	if status == "completed" && response != nil {
		return false, *response, nil
	}

	return false, "", nil // in progress
}

// Complete marks the key as completed with the serialized response.
func Complete(ctx context.Context, db *pgxpool.Pool, key string, responseJSON []byte) error {
	_, err := db.Exec(ctx,
		`UPDATE idempotency_key SET status = 'completed', response_json = $1 WHERE key = $2`,
		responseJSON, key,
	)
	return err
}

// Cleanup removes expired keys. Returns number of deleted rows.
func Cleanup(ctx context.Context, db *pgxpool.Pool) (int64, error) {
	tag, err := db.Exec(ctx, `DELETE FROM idempotency_key WHERE expires_at < NOW()`)
	return tag.RowsAffected(), err
}
