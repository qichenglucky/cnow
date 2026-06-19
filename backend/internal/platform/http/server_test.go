package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"cnow/backend/internal/app"
	"cnow/backend/internal/pkg/audit"
	apperr "cnow/backend/internal/pkg/errors"
	"cnow/backend/internal/pkg/middleware"
	"cnow/backend/internal/repo"
	"cnow/backend/internal/workflow"
	"cnow/backend/testutil"
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
	connStr := "postgres://" + user + ":"+user+"@localhost:5433/" + user + "?sslmode=disable"

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
// Mock workflow engine
// ---------------------------------------------------------------------------

type mockWorkflowEngine struct{}

func (m *mockWorkflowEngine) StartRelease(_ context.Context, _ workflow.ReleaseInput) (workflow.WorkflowHandle, error) {
	return workflow.WorkflowHandle{ID: "mock-wf", Status: "running"}, nil
}
func (m *mockWorkflowEngine) StartServiceCreate(_ context.Context, _ workflow.ServiceCreateInput) (workflow.WorkflowHandle, error) {
	return workflow.WorkflowHandle{ID: "mock-wf", Status: "running"}, nil
}
func (m *mockWorkflowEngine) StartRollback(_ context.Context, _ workflow.RollbackInput) (workflow.WorkflowHandle, error) {
	return workflow.WorkflowHandle{ID: "mock-wf", Status: "running"}, nil
}
func (m *mockWorkflowEngine) GetWorkflowStatus(_ context.Context, _ string) (string, error) {
	return "running", nil
}
func (m *mockWorkflowEngine) SignalWorkflow(_ context.Context, _ string, _ string, _ interface{}) error {
	return nil
}

// ---------------------------------------------------------------------------
// Helper: build a test HTTP handler backed by the test database
// ---------------------------------------------------------------------------

func setupTestHandler(t *testing.T) http.Handler {
	t.Helper()
	skipIfNoDB(t)
	testutil.TruncateAll(t, testPool)

	log := zap.NewNop()
	aw := audit.NewWriter(testPool, log, 1024)
	t.Cleanup(func() { aw.Close() })

	svcRepo := repo.NewServiceRepository(testPool)
	relRepo := repo.NewReleaseRepository(testPool)
	evtRepo := repo.NewEventRepository(testPool)
	envRepo := repo.NewEnvironmentRepository(testPool)
	apprRepo := repo.NewApprovalRepository(testPool)
	pipeRepo := repo.NewPipelineRepository(testPool)
	buildRepo := repo.NewBuildRepository(testPool)

	svcApp := app.NewServiceApp(svcRepo, aw, log)
	relApp := app.NewReleaseApp(relRepo, evtRepo, svcRepo, envRepo, aw, log)
	envApp := app.NewEnvironmentApp(envRepo, svcRepo, log)
	apprApp := app.NewApprovalApp(apprRepo, relRepo, evtRepo, log)
	pipeApp := app.NewPipelineApp(pipeRepo, buildRepo, log)

	return NewServer(testPool, svcApp, relApp, envApp, apprApp, pipeApp, &mockWorkflowEngine{}, aw, log)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	handler := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp middleware.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != 0 {
		t.Errorf("response code = %d, want 0", resp.Code)
	}
	if resp.Message != "ok" {
		t.Errorf("message = %q, want %q", resp.Message, "ok")
	}
}

func TestCreateAndGetService(t *testing.T) {
	handler := setupTestHandler(t)

	// Create service
	body := `{"name":"test-svc","displayName":"Test","description":"A test","techStack":"go","ownerId":1}`
	req := httptest.NewRequest("POST", "/api/v1/services", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var createResp struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Code != 0 {
		t.Errorf("create response code = %d, want 0", createResp.Code)
	}

	idFloat, ok := createResp.Data["id"].(float64)
	if !ok || idFloat == 0 {
		t.Fatalf("expected non-zero id in response, got %v", createResp.Data["id"])
	}
	id := int64(idFloat)

	// Get service by ID
	getReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/services/%d", id), nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d; body: %s", getRec.Code, http.StatusOK, getRec.Body.String())
	}

	var getResp struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(getRec.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getResp.Data["name"] != "test-svc" {
		t.Errorf("name = %v, want %q", getResp.Data["name"], "test-svc")
	}
}

func TestListServices(t *testing.T) {
	handler := setupTestHandler(t)

	// Create 3 services
	for i := 0; i < 3; i++ {
		body := fmt.Sprintf(`{"name":"list-svc-%d","techStack":"go","ownerId":1}`, i)
		req := httptest.NewRequest("POST", "/api/v1/services", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create svc %d: status = %d, body: %s", i, rec.Code, rec.Body.String())
		}
	}

	// List with pagination
	req := httptest.NewRequest("GET", "/api/v1/services?offset=0&limit=2", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Items []map[string]any `json:"items"`
			Total int              `json:"total"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if resp.Data.Total != 3 {
		t.Errorf("total = %d, want 3", resp.Data.Total)
	}
	if len(resp.Data.Items) != 2 {
		t.Errorf("len(items) = %d, want 2", len(resp.Data.Items))
	}
}

func TestCreateService_BadJSON(t *testing.T) {
	handler := setupTestHandler(t)

	body := `{invalid json`
	req := httptest.NewRequest("POST", "/api/v1/services", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Code != apperr.ErrInvalidParam.Code {
		t.Errorf("error code = %d, want %d", resp.Code, apperr.ErrInvalidParam.Code)
	}
}

func TestGetService_NotFound(t *testing.T) {
	handler := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/services/99999", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Code != apperr.ErrNotFound.Code {
		t.Errorf("error code = %d, want %d", resp.Code, apperr.ErrNotFound.Code)
	}
}
