package app

import (
	"context"
	"fmt"
	"os"
	"testing"

	"cnow/backend/internal/domain"
	"cnow/backend/internal/pkg/audit"
	"cnow/backend/internal/pkg/errors"
	"cnow/backend/internal/repo"
	"cnow/backend/testutil"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Test DB setup via TestMain
// ---------------------------------------------------------------------------

var (
	testPool *pgxpool.Pool
	dbAvail  bool
)

func TestMain(m *testing.M) {
	user := string([]byte{'c', 'n', 'o', 'w'})
	connStr := "postgres://" + user + ":" + user + "@localhost:5433/" + user + "?sslmode=disable"

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot connect to test db, skipping integration tests: %v\n", err)
		os.Exit(0)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		fmt.Fprintf(os.Stderr, "cannot ping test db, skipping integration tests: %v\n", err)
		os.Exit(0)
	}

	testPool = pool
	dbAvail = true
	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func skipIfNoDB(t *testing.T) {
	t.Helper()
	if !dbAvail {
		t.Skip("test database not available")
	}
}

// ---------------------------------------------------------------------------
// Helper: build a real ServiceApp backed by the test database
// ---------------------------------------------------------------------------

func setupServiceApp(t *testing.T) *ServiceApp {
	t.Helper()
	skipIfNoDB(t)
	testutil.TruncateAll(t, testPool)

	svcRepo := repo.NewServiceRepository(testPool)
	log := zap.NewNop()
	aw := audit.NewWriter(testPool, log, 1024)
	t.Cleanup(func() { aw.Close() })

	return NewServiceApp(svcRepo, aw, log)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCreateService_Success(t *testing.T) {
	app := setupServiceApp(t)
	ctx := context.Background()

	input := CreateServiceInput{
		Name:        "my-service",
		DisplayName: "My Service",
		Description: "A test service",
		TechStack:   "go",
		OwnerID:     1,
	}

	svc, err := app.CreateService(ctx, input, "req-1")
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}
	if svc.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if svc.Name != "my-service" {
		t.Errorf("Name = %q, want %q", svc.Name, "my-service")
	}
	if svc.DisplayName != "My Service" {
		t.Errorf("DisplayName = %q, want %q", svc.DisplayName, "My Service")
	}
	if svc.Status != domain.ServiceDraft {
		t.Errorf("Status = %q, want %q", svc.Status, domain.ServiceDraft)
	}
	if svc.TechStack != "go" {
		t.Errorf("TechStack = %q, want %q", svc.TechStack, "go")
	}
	if svc.OwnerID != 1 {
		t.Errorf("OwnerID = %d, want 1", svc.OwnerID)
	}
}

func TestCreateService_DuplicateName(t *testing.T) {
	app := setupServiceApp(t)
	ctx := context.Background()

	input := CreateServiceInput{Name: "dup-svc", TechStack: "go", OwnerID: 1}
	_, err := app.CreateService(ctx, input, "req-1")
	if err != nil {
		t.Fatalf("first CreateService failed: %v", err)
	}

	_, err = app.CreateService(ctx, input, "req-2")
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}
	appErr, ok := errors.IsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Code != errors.ErrAlreadyExists.Code {
		t.Errorf("error code = %d, want %d", appErr.Code, errors.ErrAlreadyExists.Code)
	}
}

func TestCreateService_EmptyName(t *testing.T) {
	app := setupServiceApp(t)
	ctx := context.Background()

	input := CreateServiceInput{Name: "", TechStack: "go", OwnerID: 1}
	_, err := app.CreateService(ctx, input, "req-1")
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
	appErr, ok := errors.IsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Code != errors.ErrInvalidParam.Code {
		t.Errorf("error code = %d, want %d", appErr.Code, errors.ErrInvalidParam.Code)
	}
}

func TestGetService_Success(t *testing.T) {
	app := setupServiceApp(t)
	ctx := context.Background()

	input := CreateServiceInput{Name: "get-svc", TechStack: "go", OwnerID: 1}
	created, err := app.CreateService(ctx, input, "req-1")
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	got, err := app.GetService(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
	if got.Name != "get-svc" {
		t.Errorf("Name = %q, want %q", got.Name, "get-svc")
	}
	if got.Status != domain.ServiceDraft {
		t.Errorf("Status = %q, want %q", got.Status, domain.ServiceDraft)
	}
}

func TestGetService_NotFound(t *testing.T) {
	app := setupServiceApp(t)
	ctx := context.Background()

	_, err := app.GetService(ctx, 99999)
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
	appErr, ok := errors.IsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Code != errors.ErrNotFound.Code {
		t.Errorf("error code = %d, want %d", appErr.Code, errors.ErrNotFound.Code)
	}
}

func TestListServices(t *testing.T) {
	app := setupServiceApp(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		input := CreateServiceInput{
			Name:      fmt.Sprintf("svc-%d", i),
			TechStack: "go",
			OwnerID:   1,
		}
		if _, err := app.CreateService(ctx, input, "req"); err != nil {
			t.Fatalf("CreateService %d failed: %v", i, err)
		}
	}

	// First page: limit 2
	result, err := app.ListServices(ctx, ServiceFilter{
		Pagination: domain.Pagination{Offset: 0, Limit: 2},
	})
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("Total = %d, want 3", result.Total)
	}
	if len(result.Items) != 2 {
		t.Errorf("len(Items) = %d, want 2", len(result.Items))
	}

	// Second page: offset 2, limit 2
	result2, err := app.ListServices(ctx, ServiceFilter{
		Pagination: domain.Pagination{Offset: 2, Limit: 2},
	})
	if err != nil {
		t.Fatalf("ListServices page 2 failed: %v", err)
	}
	if len(result2.Items) != 1 {
		t.Errorf("page 2 len(Items) = %d, want 1", len(result2.Items))
	}
}

func TestUpdateServiceStatus_Valid(t *testing.T) {
	app := setupServiceApp(t)
	ctx := context.Background()

	input := CreateServiceInput{Name: "trans-svc", TechStack: "go", OwnerID: 1}
	svc, err := app.CreateService(ctx, input, "req-1")
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	// draft -> creating
	err = app.UpdateServiceStatus(ctx, svc.ID, domain.ServiceCreating, "1", "req-2")
	if err != nil {
		t.Fatalf("draft -> creating failed: %v", err)
	}

	// Verify status was updated
	got, err := app.GetService(ctx, svc.ID)
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}
	if got.Status != domain.ServiceCreating {
		t.Errorf("Status = %q, want %q", got.Status, domain.ServiceCreating)
	}
}

func TestUpdateServiceStatus_Invalid(t *testing.T) {
	app := setupServiceApp(t)
	ctx := context.Background()

	input := CreateServiceInput{Name: "bad-trans", TechStack: "go", OwnerID: 1}
	svc, err := app.CreateService(ctx, input, "req-1")
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	// draft -> ready (skip creating) should fail
	err = app.UpdateServiceStatus(ctx, svc.ID, domain.ServiceReady, "1", "req-2")
	if err == nil {
		t.Fatal("expected error for invalid transition, got nil")
	}
	appErr, ok := errors.IsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Code != errors.ErrConflict.Code {
		t.Errorf("error code = %d, want %d", appErr.Code, errors.ErrConflict.Code)
	}
}
