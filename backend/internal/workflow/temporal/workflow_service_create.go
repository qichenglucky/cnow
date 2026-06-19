package temporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ServiceCreateState tracks completed steps for saga compensation.
type ServiceCreateState struct {
	ServiceID     int64
	RepoURL       string
	ConfigRef     string
	EnvID         int64
	DomainFQDN    string
	CertID        int64
	LogSourceID   int64
	PanelID       int64
	Finalized     bool
}

// ServiceCreateResult is the workflow return value.
type ServiceCreateResult struct {
	ServiceID  int64  `json:"serviceId"`
	RepoURL    string `json:"repoUrl"`
	DomainFQDN string `json:"domainFqdn"`
	Status     string `json:"status"`
}

// ServiceCreateWorkflow orchestrates the full service creation saga.
func ServiceCreateWorkflow(ctx workflow.Context, input CreateServiceInput) (ServiceCreateResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ServiceCreateWorkflow started", "name", input.Name)

	state := &ServiceCreateState{}

	// Activity options for service-creation steps
	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         StandardRetry,
	}
	ctx = workflow.WithActivityOptions(ctx, actOpts)

	var result ServiceCreateResult

	// ── Step 1: Create service record ────────────────────────────────────
	var svcOut CreateServiceOutput
	if err := workflow.ExecuteActivity(ctx, ActivityCreateService, CreateServiceInput{
		Name:        input.Name,
		DisplayName: input.DisplayName,
		TechStack:   input.TechStack,
		RepoMode:    input.RepoMode,
	}).Get(ctx, &svcOut); err != nil {
		return result, temporal.NewApplicationError("create service failed", "CreateServiceError", err)
	}
	state.ServiceID = svcOut.ServiceID

	// ── Step 2: Create repo ──────────────────────────────────────────────
	var repoOut CreateRepoOutput
	if err := workflow.ExecuteActivity(ctx, ActivityCreateRepo, CreateRepoInput{
		ServiceID: state.ServiceID,
		RepoMode:  input.RepoMode,
		Name:      input.Name,
	}).Get(ctx, &repoOut); err != nil {
		compensateServiceCreate(ctx, state)
		return result, temporal.NewApplicationError("create repo failed", "CreateRepoError", err)
	}
	state.RepoURL = repoOut.RepoURL

	// ── Step 3: Generate CI config ───────────────────────────────────────
	var ciOut GenerateCIOutput
	if err := workflow.ExecuteActivity(ctx, ActivityGenerateCI, GenerateCIInput{
		ServiceID: state.ServiceID,
		RepoURL:   state.RepoURL,
		TechStack: input.TechStack,
	}).Get(ctx, &ciOut); err != nil {
		compensateServiceCreate(ctx, state)
		return result, temporal.NewApplicationError("generate CI failed", "GenerateCIError", err)
	}
	state.ConfigRef = ciOut.ConfigRef

	// ── Step 4: Create test environment ──────────────────────────────────
	var envOut CreateTestEnvOutput
	if err := workflow.ExecuteActivity(ctx, ActivityCreateTestEnv, CreateTestEnvInput{
		ServiceID: state.ServiceID,
	}).Get(ctx, &envOut); err != nil {
		compensateServiceCreate(ctx, state)
		return result, temporal.NewApplicationError("create test env failed", "CreateTestEnvError", err)
	}
	state.EnvID = envOut.EnvID

	// ── Step 5: Bind domain ──────────────────────────────────────────────
	var domOut BindDomainOutput
	if err := workflow.ExecuteActivity(ctx, ActivityBindDomain, BindDomainInput{
		ServiceID: state.ServiceID,
		EnvID:     state.EnvID,
	}).Get(ctx, &domOut); err != nil {
		compensateServiceCreate(ctx, state)
		return result, temporal.NewApplicationError("bind domain failed", "BindDomainError", err)
	}
	state.DomainFQDN = domOut.DomainFQDN

	// ── Step 6: Request certificate ──────────────────────────────────────
	var certOut RequestCertOutput
	if err := workflow.ExecuteActivity(ctx, ActivityRequestCert, RequestCertInput{
		DomainFQDN: state.DomainFQDN,
	}).Get(ctx, &certOut); err != nil {
		compensateServiceCreate(ctx, state)
		return result, temporal.NewApplicationError("request cert failed", "RequestCertError", err)
	}
	state.CertID = certOut.CertID

	// ── Step 7: Attach log source ────────────────────────────────────────
	var logOut AttachLogOutput
	if err := workflow.ExecuteActivity(ctx, ActivityAttachLog, AttachLogInput{
		ServiceID: state.ServiceID,
		EnvID:     state.EnvID,
	}).Get(ctx, &logOut); err != nil {
		compensateServiceCreate(ctx, state)
		return result, temporal.NewApplicationError("attach log failed", "AttachLogError", err)
	}
	state.LogSourceID = logOut.LogSourceID

	// ── Step 8: Attach metric panel ──────────────────────────────────────
	var metricOut AttachMetricOutput
	if err := workflow.ExecuteActivity(ctx, ActivityAttachMetric, AttachMetricInput{
		ServiceID: state.ServiceID,
		EnvID:     state.EnvID,
	}).Get(ctx, &metricOut); err != nil {
		compensateServiceCreate(ctx, state)
		return result, temporal.NewApplicationError("attach metric failed", "AttachMetricError", err)
	}
	state.PanelID = metricOut.PanelID

	// ── Step 9: Finalize service ─────────────────────────────────────────
	if err := workflow.ExecuteActivity(ctx, ActivityFinalizeService, FinalizeServiceInput{
		ServiceID: state.ServiceID,
	}).Get(ctx, nil); err != nil {
		compensateServiceCreate(ctx, state)
		return result, temporal.NewApplicationError("finalize service failed", "FinalizeServiceError", err)
	}
	state.Finalized = true

	logger.Info("ServiceCreateWorkflow completed", "serviceId", state.ServiceID)
	return ServiceCreateResult{
		ServiceID:  state.ServiceID,
		RepoURL:    state.RepoURL,
		DomainFQDN: state.DomainFQDN,
		Status:     "ready",
	}, nil
}

// compensateServiceCreate performs best-effort rollback of completed steps.
// In production this would call dedicated compensation activities.
func compensateServiceCreate(ctx workflow.Context, state *ServiceCreateState) {
	logger := workflow.GetLogger(ctx)
	logger.Warn("compensating service creation",
		"serviceId", state.ServiceID,
		"repoURL", state.RepoURL,
		"envId", state.EnvID,
		"domainFqdn", state.DomainFQDN,
		"certId", state.CertID,
		"logSourceId", state.LogSourceID,
		"panelId", state.PanelID,
	)
	// TODO: execute compensation activities in reverse order:
	// 1. Detach metric panel  (if panelID > 0)
	// 2. Detach log source    (if logSourceID > 0)
	// 3. Revoke certificate   (if certID > 0)
	// 4. Unbind domain        (if domainFQDN != "")
	// 5. Destroy test env     (if envID > 0)
	// 6. Delete repo          (if repoURL != "")
	// 7. Mark service failed  (if serviceID > 0)
}
