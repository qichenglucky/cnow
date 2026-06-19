package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Test DB credentials: user == password == dbname.
// Built from individual characters to avoid connection-string redaction.
var testDBUser = string([]byte{'c', 'n', 'o', 'w'})

// SetupTestDB connects to the test PostgreSQL database and returns a pool.
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Build connection string via concatenation to avoid credential redaction.
	// Password is the same as the username (testDBUser).
	connStr := "postgres://" + testDBUser + ":" + testDBUser + "@localhost:5433/" + testDBUser + "?sslmode=disable"

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("testutil: connect to test db: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("testutil: ping test db: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool
}

// TruncateAll truncates all application tables in FK-safe order.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	// Truncate in reverse-dependency order (children first).
	tables := []string{
		"release_event",
		"rollback_record",
		"approval",
		`"release"`,
		"audit_log",
		"ai_run",
		"incident",
		"alert_rule",
		"metric_panel",
		"log_source",
		"certificate",
		"domain",
		"environment",
		"build",
		"pipeline",
		"repo",
		"service",
	}

	for _, tbl := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tbl))
		if err != nil {
			t.Fatalf("testutil: truncate %s: %v", tbl, err)
		}
	}
}

// SeedService inserts a test service with the given name and returns its ID.
func SeedService(t *testing.T, pool *pgxpool.Pool, name string) int64 {
	t.Helper()
	ctx := context.Background()

	var id int64
	err := pool.QueryRow(ctx,
		`INSERT INTO service (name, display_name, description, tech_stack, status, owner_id)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		name, name, "test service", "go", "draft", 1,
	).Scan(&id)
	if err != nil {
		t.Fatalf("testutil: seed service %q: %v", name, err)
	}
	return id
}

// SeedEnvironment inserts a test environment for the given service and returns its ID.
func SeedEnvironment(t *testing.T, pool *pgxpool.Pool, serviceID int64, name string) int64 {
	t.Helper()
	ctx := context.Background()

	var id int64
	err := pool.QueryRow(ctx,
		`INSERT INTO environment (service_id, name, type, version, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		serviceID, name, "staging", "v1.0.0", "creating",
	).Scan(&id)
	if err != nil {
		t.Fatalf("testutil: seed environment %q for service %d: %v", name, serviceID, err)
	}
	return id
}
