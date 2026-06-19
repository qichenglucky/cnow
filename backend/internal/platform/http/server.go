package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"cnow/backend/internal/app"
	"cnow/backend/internal/domain"
	"cnow/backend/internal/pkg/audit"
	apperr "cnow/backend/internal/pkg/errors"
	"cnow/backend/internal/pkg/idempotency"
	"cnow/backend/internal/pkg/middleware"
	"cnow/backend/internal/workflow"
)

// idempotencyWriter captures the response for caching.
type idempotencyWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *idempotencyWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *idempotencyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// server holds all application dependencies and implements http.Handler.
type server struct {
	pool           *pgxpool.Pool
	serviceApp     *app.ServiceApp
	releaseApp     *app.ReleaseApp
	envApp         *app.EnvironmentApp
	approvalApp    *app.ApprovalApp
	pipelineApp    *app.PipelineApp
	workflowEngine workflow.Engine
	auditWriter    *audit.Writer
	log            *zap.Logger
}

// NewServer wires all application layers and returns a fully configured http.Handler
// with middleware chain: CORS → Recovery → Logging → RequestID.
func NewServer(
	pool *pgxpool.Pool,
	serviceApp *app.ServiceApp,
	releaseApp *app.ReleaseApp,
	envApp *app.EnvironmentApp,
	approvalApp *app.ApprovalApp,
	pipelineApp *app.PipelineApp,
	workflowEngine workflow.Engine,
	auditWriter *audit.Writer,
	log *zap.Logger,
) http.Handler {
	s := &server{
		pool:           pool,
		serviceApp:     serviceApp,
		releaseApp:     releaseApp,
		envApp:         envApp,
		approvalApp:    approvalApp,
		pipelineApp:    pipelineApp,
		workflowEngine: workflowEngine,
		auditWriter:    auditWriter,
		log:            log,
	}

	mux := http.NewServeMux()

	// Health / readiness
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readyz", s.handleReadyz)

	// Services
	mux.HandleFunc("GET /api/v1/services", s.handleListServices)
	mux.HandleFunc("POST /api/v1/services", s.handleCreateService)
	mux.HandleFunc("GET /api/v1/services/{id}", s.handleGetService)
	mux.HandleFunc("PUT /api/v1/services/{id}", s.handleUpdateService)
	mux.HandleFunc("DELETE /api/v1/services/{id}", s.handleDeleteService)

	// Releases
	mux.HandleFunc("GET /api/v1/releases", s.handleListReleases)
	mux.HandleFunc("POST /api/v1/releases", s.handleCreateRelease)
	mux.HandleFunc("GET /api/v1/releases/{id}", s.handleGetRelease)
	mux.HandleFunc("POST /api/v1/releases/{id}/rollback", s.handleRollbackRelease)

	// Environments
	mux.HandleFunc("GET /api/v1/environments", s.handleListEnvironments)
	mux.HandleFunc("POST /api/v1/environments", s.handleCreateEnvironment)

	// Approvals
	mux.HandleFunc("POST /api/v1/approvals", s.handleCreateApproval)
	mux.HandleFunc("GET /api/v1/approvals", s.handleListApprovals)
	mux.HandleFunc("PUT /api/v1/approvals/{id}/approve", s.handleApprove)
	mux.HandleFunc("PUT /api/v1/approvals/{id}/reject", s.handleReject)

	// AI
	mux.HandleFunc("POST /api/v1/ai/plan", s.handleAIPlan)
	mux.HandleFunc("POST /api/v1/ai/risk", s.handleAIRisk)

	// Observability
	mux.HandleFunc("GET /api/v1/observability/logs", s.handleObsLogs)
	mux.HandleFunc("GET /api/v1/observability/metrics", s.handleObsMetrics)
	mux.HandleFunc("GET /api/v1/observability/alerts", s.handleObsAlerts)

	// Pipelines
	mux.HandleFunc("POST /api/v1/pipelines", s.handleCreatePipeline)
	mux.HandleFunc("GET /api/v1/pipelines/{id}", s.handleGetPipeline)

	// Builds
	mux.HandleFunc("POST /api/v1/builds", s.handleCreateBuild)
	mux.HandleFunc("PUT /api/v1/builds/{id}", s.handleUpdateBuildStatus)
	mux.HandleFunc("GET /api/v1/builds", s.handleListBuilds)

	// Webhooks
	mux.HandleFunc("POST /api/v1/webhooks/github", s.handleGitHubWebhook)

	return middleware.Chain(
		mux,
		middleware.CORS([]string{"*"}),
		middleware.Recovery(log),
		middleware.Logging(log),
		middleware.RequestID,
		middleware.InjectIdentity,
		middleware.ContentTypeJSON,
	)
}

// ---------------------------------------------------------------------------
// Health / Readiness
// ---------------------------------------------------------------------------

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	if err := s.pool.Ping(r.Context()); err != nil {
		middleware.WriteError(w, http.StatusServiceUnavailable, 4001, "db ping failed", rid)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "ok", map[string]string{"status": "ok"}, rid)
}

func (s *server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	if err := s.pool.Ping(r.Context()); err != nil {
		middleware.WriteError(w, http.StatusServiceUnavailable, 4001, "db not ready", rid)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "ok", map[string]string{"status": "ready"}, rid)
}

// ---------------------------------------------------------------------------
// Services
// ---------------------------------------------------------------------------

func (s *server) handleListServices(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	filter := app.ServiceFilter{}
	filter.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))

	result, err := s.serviceApp.ListServices(r.Context(), filter)
	if err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "ok", result, rid)
}

func (s *server) handleCreateService(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	var input app.CreateServiceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", rid)
		return
	}

	// Idempotency check
	idemKey := r.Header.Get("X-Idempotency-Key")
	if idemKey != "" {
		body, _ := json.Marshal(input)
		reqHash := idempotency.HashPayload(body)
		reserved, cached, err := idempotency.CheckAndReserve(r.Context(), s.pool, idemKey, reqHash)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, 4001, "idempotency check failed", rid)
			return
		}
		if !reserved {
			if cached != "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(cached))
				return
			}
			middleware.WriteError(w, http.StatusConflict, 1003, "request in progress", rid)
			return
		}
	}

	iw := &idempotencyWriter{ResponseWriter: w, status: http.StatusOK}

	svc, err := s.serviceApp.CreateService(r.Context(), input, rid)
	if err != nil {
		writeAppError(iw, err, rid)
		if idemKey != "" {
			idempotency.Complete(r.Context(), s.pool, idemKey, iw.body.Bytes())
		}
		return
	}
	middleware.WriteJSON(iw, http.StatusCreated, 0, "created", svc, rid)
	if idemKey != "" {
		idempotency.Complete(r.Context(), s.pool, idemKey, iw.body.Bytes())
	}
}

func (s *server) handleGetService(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid service id", rid)
		return
	}

	svc, err := s.serviceApp.GetService(r.Context(), id)
	if err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "ok", svc, rid)
}

// ---------------------------------------------------------------------------
// Releases
// ---------------------------------------------------------------------------

func (s *server) handleListReleases(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	filter := app.ReleaseFilter{}
	filter.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))

	if v := r.URL.Query().Get("service_id"); v != "" {
		sid, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid service_id", rid)
			return
		}
		filter.ServiceID = &sid
	}
	if v := r.URL.Query().Get("status"); v != "" {
		filter.Status = &v
	}

	result, err := s.releaseApp.ListReleases(r.Context(), filter)
	if err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "ok", result, rid)
}

func (s *server) handleCreateRelease(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	var input app.CreateReleaseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", rid)
		return
	}

	// Idempotency check
	idemKey := r.Header.Get("X-Idempotency-Key")
	if idemKey != "" {
		body, _ := json.Marshal(input)
		reqHash := idempotency.HashPayload(body)
		reserved, cached, err := idempotency.CheckAndReserve(r.Context(), s.pool, idemKey, reqHash)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, 4001, "idempotency check failed", rid)
			return
		}
		if !reserved {
			if cached != "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(cached))
				return
			}
			middleware.WriteError(w, http.StatusConflict, 1003, "request in progress", rid)
			return
		}
	}

	iw := &idempotencyWriter{ResponseWriter: w, status: http.StatusOK}

	rel, err := s.releaseApp.CreateRelease(r.Context(), input, rid)
	if err != nil {
		writeAppError(iw, err, rid)
		if idemKey != "" {
			idempotency.Complete(r.Context(), s.pool, idemKey, iw.body.Bytes())
		}
		return
	}
	middleware.WriteJSON(iw, http.StatusCreated, 0, "created", rel, rid)
	if idemKey != "" {
		idempotency.Complete(r.Context(), s.pool, idemKey, iw.body.Bytes())
	}
}

func (s *server) handleGetRelease(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid release id", rid)
		return
	}

	rel, events, err := s.releaseApp.GetRelease(r.Context(), id)
	if err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "ok", map[string]any{
		"release": rel,
		"events":  events,
	}, rid)
}

// ---------------------------------------------------------------------------
// Environments
// ---------------------------------------------------------------------------

func (s *server) handleListEnvironments(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	sidStr := r.URL.Query().Get("service_id")
	if sidStr == "" {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "service_id is required", rid)
		return
	}
	sid, err := strconv.ParseInt(sidStr, 10, 64)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid service_id", rid)
		return
	}

	envs, err := s.envApp.ListByService(r.Context(), sid)
	if err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "ok", map[string]any{"items": envs}, rid)
}

func (s *server) handleCreateEnvironment(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	var input app.CreateEnvironmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", rid)
		return
	}

	env, err := s.envApp.CreateEnvironment(r.Context(), input)
	if err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, 0, "created", env, rid)
}

// ---------------------------------------------------------------------------
// AI endpoints
// ---------------------------------------------------------------------------

func (s *server) handleAIPlan(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	var req struct {
		Prompt      string `json:"prompt"`
		ServiceID   string `json:"service_id"`
		Environment string `json:"environment"`
		Version     string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", rid)
		return
	}
	if req.Prompt == "" {
		req.Prompt = "plan deployment"
	}

	steps := []map[string]any{
		{"order": 1, "action": "build", "description": "Build Docker image from source", "estimated_duration": "2m", "status": "pending"},
		{"order": 2, "action": "test", "description": "Run unit and integration tests", "estimated_duration": "5m", "status": "pending"},
		{"order": 3, "action": "deploy_canary", "description": "Deploy canary (10% traffic)", "estimated_duration": "3m", "status": "pending"},
		{"order": 4, "action": "observe", "description": "Monitor canary for 5 minutes", "estimated_duration": "5m", "status": "pending"},
		{"order": 5, "action": "promote", "description": "Promote to 100% traffic", "estimated_duration": "2m", "status": "pending"},
	}

	middleware.WriteJSON(w, http.StatusOK, 0, "ok", map[string]any{
		"risk_level":          "low",
		"confidence":          0.92,
		"editable":            true,
		"estimated_duration":  "17m",
		"steps":               steps,
		"recommendations":     []string{"Enable canary deployment for safer rollout", "Ensure monitoring dashboards are active before deploy"},
		"summary":             "Deployment plan generated for prompt: " + req.Prompt,
	}, rid)
}

func (s *server) handleAIRisk(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	var req struct {
		Prompt      string `json:"prompt"`
		ServiceID   string `json:"service_id"`
		Environment string `json:"environment"`
		Version     string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", rid)
		return
	}

	riskLevel := "medium"
	riskScore := 55
	if req.Environment == "production" || req.Environment == "生产环境" {
		riskLevel = "high"
		riskScore = 78
	}

	middleware.WriteJSON(w, http.StatusOK, 0, "ok", map[string]any{
		"risk_level":      riskLevel,
		"risk_score":      riskScore,
		"confidence":      0.87,
		"reason":          "Deployment includes infrastructure config changes in " + req.Environment,
		"factors":         []string{"infrastructure_changes", "config_modification", "dependency_updates"},
		"recommendations": []string{"Run full regression test suite", "Prepare rollback plan", "Notify on-call engineer before deploy"},
	}, rid)
}

// ---------------------------------------------------------------------------
// Observability endpoints
// ---------------------------------------------------------------------------

func (s *server) handleObsLogs(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	services := []string{"user-center", "order-service", "payment-gateway", "msg-push", "data-platform"}
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	messages := []string{
		"Request processed successfully",
		"Database connection pool exhausted, waiting for available connection",
		"Cache miss for key user:12345, fetching from DB",
		"Health check passed",
		"Timeout waiting for downstream service response",
		"Rate limit exceeded for client api-key-xxx",
		"Deployment v2.3.1 rolled out to canary environment",
		"JWT token expired, rejecting request",
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	items := make([]map[string]any, limit)
	for i := 0; i < limit; i++ {
		items[i] = map[string]any{
			"id":        fmt.Sprintf("log-%d", i),
			"timestamp": time.Now().Add(-time.Duration(rnd.Intn(3600)) * time.Second).UTC().Format(time.RFC3339),
			"level":     levels[rnd.Intn(len(levels))],
			"service":   services[rnd.Intn(len(services))],
			"message":   messages[rnd.Intn(len(messages))],
			"traceId":   fmt.Sprintf("trace-%06d", rnd.Intn(1000000)),
		}
	}

	middleware.WriteJSON(w, http.StatusOK, 0, "ok", map[string]any{
		"items": items,
		"total": limit,
	}, rid)
}

func (s *server) handleObsMetrics(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	latency := make([]map[string]any, 24)
	errorRate := make([]map[string]any, 24)
	throughput := make([]map[string]any, 24)
	for i := 0; i < 24; i++ {
		ts := fmt.Sprintf("%02d:00", i)
		latency[i] = map[string]any{
			"time": ts,
			"p50":  math.Round((20+rnd.Float64()*30)*100) / 100,
			"p95":  math.Round((80+rnd.Float64()*60)*100) / 100,
			"p99":  math.Round((150+rnd.Float64()*100)*100) / 100,
		}
		errorRate[i] = map[string]any{
			"time":  ts,
			"rate":  math.Round((0.1+rnd.Float64()*0.8)*100) / 100,
			"count": rnd.Intn(50),
		}
		throughput[i] = map[string]any{
			"time":       ts,
			"requests":   500 + rnd.Intn(2000),
			"success":    480 + rnd.Intn(1900),
		}
	}

	middleware.WriteJSON(w, http.StatusOK, 0, "ok", map[string]any{
		"latency":    latency,
		"error_rate": errorRate,
		"throughput": throughput,
	}, rid)
}

func (s *server) handleObsAlerts(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	alerts := []map[string]any{
		{"id": "alert-1", "level": "critical", "service": "payment-gateway", "message": "P99 latency exceeds 500ms", "time": "2026-06-12T08:15:00Z", "status": "firing"},
		{"id": "alert-2", "level": "warning", "service": "order-service", "message": "Error rate > 1%", "time": "2026-06-12T07:30:00Z", "status": "acknowledged"},
		{"id": "alert-3", "level": "info", "service": "user-center", "message": "CPU usage exceeds 80%", "time": "2026-06-12T06:00:00Z", "status": "resolved"},
		{"id": "alert-4", "level": "critical", "service": "search-service", "message": "Pod restart count excessive", "time": "2026-06-11T23:00:00Z", "status": "firing"},
		{"id": "alert-5", "level": "warning", "service": "data-platform", "message": "Disk usage exceeds 90%", "time": "2026-06-11T20:00:00Z", "status": "acknowledged"},
	}

	middleware.WriteJSON(w, http.StatusOK, 0, "ok", map[string]any{
		"items": alerts,
		"total": len(alerts),
	}, rid)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// writeAppError maps an application error to the appropriate HTTP status.
func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func writeAppError(w http.ResponseWriter, err error, requestID string) {
	if appErr, ok := apperr.IsAppError(err); ok {
		status := http.StatusInternalServerError
		switch appErr.Code {
		case 1001: // not found
			status = http.StatusNotFound
		case 1002: // already exists
			status = http.StatusConflict
		case 1003: // conflict
			status = http.StatusConflict
		case 2001: // invalid param
			status = http.StatusBadRequest
		case 2002: // unauthorized
			status = http.StatusUnauthorized
		case 2003: // forbidden
			status = http.StatusForbidden
		}
		msg := appErr.Message
		if appErr.Details != "" {
			msg = appErr.Details
		}
		middleware.WriteError(w, status, appErr.Code, msg, requestID)
		return
	}
	middleware.WriteError(w, http.StatusInternalServerError, 4001, "internal server error", requestID)
}

// ---------------------------------------------------------------------------
// Service Edit / Delete
// ---------------------------------------------------------------------------

func (s *server) handleUpdateService(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid service id", requestID)
		return
	}
	var input struct {
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
		TechStack   string `json:"techStack"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", requestID)
		return
	}
	svc, err := s.serviceApp.GetService(r.Context(), id)
	if err != nil {
		writeAppError(w, err, requestID)
		return
	}
	if input.DisplayName != "" {
		svc.DisplayName = input.DisplayName
	}
	if input.Description != "" {
		svc.Description = input.Description
	}
	if input.TechStack != "" {
		svc.TechStack = input.TechStack
	}
	// Update via repo directly (simple update)
	if err := s.serviceApp.UpdateService(r.Context(), svc); err != nil {
		writeAppError(w, err, requestID)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "updated", svc, requestID)
}

func (s *server) handleDeleteService(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid service id", requestID)
		return
	}
	if err := s.serviceApp.DeleteService(r.Context(), id); err != nil {
		writeAppError(w, err, requestID)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "deleted", nil, requestID)
}

// ---------------------------------------------------------------------------
// Release Rollback
// ---------------------------------------------------------------------------

func (s *server) handleRollbackRelease(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid release id", requestID)
		return
	}
	var input struct {
		Reason      string `json:"reason"`
		TriggeredBy int64  `json:"triggeredBy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", requestID)
		return
	}
	// Get current release to check status
	rel, _, err := s.releaseApp.GetRelease(r.Context(), id)
	if err != nil {
		writeAppError(w, err, requestID)
		return
	}
	// Rollback is only valid from deploying/verifying/observing/succeeded
	validRollback := map[domain.ReleaseStatus]bool{
		domain.ReleaseDeploying: true, domain.ReleaseVerifying: true,
		domain.ReleaseObserving: true, domain.ReleaseSucceeded: true,
	}
	if !validRollback[rel.Status] {
		middleware.WriteError(w, http.StatusConflict, 1003,
			fmt.Sprintf("cannot rollback release in '%s' status", rel.Status), requestID)
		return
	}
	if err := s.releaseApp.UpdateReleaseStatus(r.Context(), id, domain.ReleaseRollbackPending, input.TriggeredBy, requestID); err != nil {
		writeAppError(w, err, requestID)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "rollback initiated", map[string]interface{}{"releaseId": id, "reason": input.Reason}, requestID)
}

// ---------------------------------------------------------------------------
// Approvals
// ---------------------------------------------------------------------------

func (s *server) handleCreateApproval(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	var input struct {
		ReleaseID  int64 `json:"releaseId"`
		ApproverID int64 `json:"approverId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", requestID)
		return
	}
	approval, err := s.approvalApp.CreateApproval(r.Context(), input.ReleaseID, input.ApproverID)
	if err != nil {
		writeAppError(w, err, requestID)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, 0, "created", approval, requestID)
}

func (s *server) handleListApprovals(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	releaseIDStr := r.URL.Query().Get("release_id")
	if releaseIDStr == "" {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "release_id is required", requestID)
		return
	}
	releaseID, err := parseID(releaseIDStr)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid release_id", requestID)
		return
	}
	items, err := s.approvalApp.ListByRelease(r.Context(), releaseID)
	if err != nil {
		writeAppError(w, err, requestID)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "ok", items, requestID)
}

func (s *server) handleApprove(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid approval id", requestID)
		return
	}
	var input struct {
		ApproverID int64  `json:"approverId"`
		Comment    string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&input)
	if err := s.approvalApp.Approve(r.Context(), id, input.ApproverID, input.Comment); err != nil {
		writeAppError(w, err, requestID)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "approved", nil, requestID)
}

func (s *server) handleReject(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid approval id", requestID)
		return
	}
	var input struct {
		ApproverID int64  `json:"approverId"`
		Comment    string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&input)
	if err := s.approvalApp.Reject(r.Context(), id, input.ApproverID, input.Comment); err != nil {
		writeAppError(w, err, requestID)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "rejected", nil, requestID)
}

// ---------------------------------------------------------------------------
// Pipelines
// ---------------------------------------------------------------------------

func (s *server) handleCreatePipeline(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	var input app.CreatePipelineInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", rid)
		return
	}

	p, err := s.pipelineApp.CreatePipeline(r.Context(), input)
	if err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, 0, "created", p, rid)
}

func (s *server) handleGetPipeline(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	id, err := parseID(r.PathValue("id"))
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid pipeline id", rid)
		return
	}

	p, err := s.pipelineApp.GetPipeline(r.Context(), id)
	if err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "ok", p, rid)
}

// ---------------------------------------------------------------------------
// Builds
// ---------------------------------------------------------------------------

func (s *server) handleCreateBuild(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	var input app.CreateBuildInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", rid)
		return
	}

	b, err := s.pipelineApp.CreateBuild(r.Context(), input)
	if err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, 0, "created", b, rid)
}

func (s *server) handleUpdateBuildStatus(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	id, err := parseID(r.PathValue("id"))
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid build id", rid)
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", rid)
		return
	}

	if err := s.pipelineApp.UpdateBuildStatus(r.Context(), id, domain.BuildStatus(input.Status)); err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "updated", nil, rid)
}

func (s *server) handleListBuilds(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	filter := app.BuildFilter{}
	filter.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))

	if v := r.URL.Query().Get("pipeline_id"); v != "" {
		pid, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid pipeline_id", rid)
			return
		}
		filter.PipelineID = &pid
	}

	result, err := s.pipelineApp.ListBuilds(r.Context(), filter)
	if err != nil {
		writeAppError(w, err, rid)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, 0, "ok", result, rid)
}

// ---------------------------------------------------------------------------
// GitHub Webhook
// ---------------------------------------------------------------------------

func (s *server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())

	// Validate API token
	apiToken := r.Header.Get("X-API-Token")
	if apiToken == "" {
		middleware.WriteError(w, http.StatusUnauthorized, 2002, "missing X-API-Token header", rid)
		return
	}
	// TODO: validate token against configured value (MVP: accept any non-empty token)

	var payload struct {
		BuildID int64  `json:"build_id"`
		Status  string `json:"status"`
		Action  string `json:"action"`
		Workflow string `json:"workflow"`
		Branch  string `json:"branch"`
		Commit  string `json:"commit_sha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "invalid JSON body", rid)
		return
	}

	if payload.BuildID == 0 || payload.Status == "" {
		middleware.WriteError(w, http.StatusBadRequest, 2001, "build_id and status are required", rid)
		return
	}

	buildStatus := domain.BuildStatus(payload.Status)
	if err := s.pipelineApp.UpdateBuildStatus(r.Context(), payload.BuildID, buildStatus); err != nil {
		writeAppError(w, err, rid)
		return
	}

	s.log.Info("github webhook processed",
		zap.Int64("build_id", payload.BuildID),
		zap.String("status", payload.Status),
		zap.String("workflow", payload.Workflow),
	)

	middleware.WriteJSON(w, http.StatusOK, 0, "ok", map[string]string{"status": "processed"}, rid)
}
