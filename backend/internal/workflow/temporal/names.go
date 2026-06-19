package temporal

// ─── Workflow names ──────────────────────────────────────────────────────────

const (
	WorkflowServiceCreate = "ServiceCreateWorkflow"
	WorkflowRelease       = "ReleaseWorkflow"
	WorkflowRollback      = "RollbackWorkflow"
	WorkflowDomainBind    = "DomainBindWorkflow"
	WorkflowCertIssue     = "CertificateIssueWorkflow"
	WorkflowLogAttach     = "LogAttachWorkflow"
)

// ─── Activity names — service creation ───────────────────────────────────────

const (
	ActivityCreateService   = "CreateServiceRecord"
	ActivityCreateRepo      = "CreateRepo"
	ActivityGenerateCI      = "GenerateCIConfig"
	ActivityCreateTestEnv   = "CreateTestEnvironment"
	ActivityBindDomain      = "BindDomain"
	ActivityRequestCert     = "RequestCertificate"
	ActivityAttachLog       = "AttachLogSource"
	ActivityAttachMetric    = "AttachMetricPanel"
	ActivityFinalizeService = "FinalizeService"
)

// ─── Activity names — release ────────────────────────────────────────────────

const (
	ActivityValidateRelease = "ValidateReleasePrecondition"
	ActivityRiskAnalysis    = "RunRiskAnalysis"
	ActivityCreateBuild     = "CreateBuild"
	ActivityWaitForBuild    = "WaitForBuildComplete"
	ActivityDeploy          = "Deploy"
	ActivityHealthCheck     = "HealthCheck"
	ActivityObserve         = "ObserveWindow"
	ActivityAutoRollback    = "AutoRollback"
	ActivityFinalizeRelease = "FinalizeRelease"
)

// ─── Activity names — rollback ───────────────────────────────────────────────

const (
	ActivityValidateRollback    = "ValidateRollback"
	ActivityGetPrevVersion      = "GetPreviousVersion"
	ActivityExecuteRollback     = "ExecuteRollback"
	ActivityVerifyRollback      = "VerifyRollback"
	ActivityCreateRollbackRecord = "CreateRollbackRecord"
)

// ─── Activity names — domain binding ─────────────────────────────────────────

const (
	ActivityCreateDNS = "CreateDNSEntry"
	ActivityWaitDNS   = "WaitForDNSPropagation"
	ActivityWaitCert  = "WaitForCertReady"
)

// ─── Activity names — certificate storage ────────────────────────────────────

const (
	ActivityStoreCert = "StoreCertSecret"
)

// ─── Activity names — log attachment ─────────────────────────────────────────

const (
	ActivityCreateLogstore     = "CreateLogstore"
	ActivityConfigLogAgent     = "ConfigureLogAgent"
	ActivityCreateLogDashboard = "CreateLogDashboard"
	ActivityVerifyLog          = "VerifyLogIngestion"
)

// ─── Signal names ────────────────────────────────────────────────────────────

const (
	SignalApproval = "approval_signal"
)

// ─── Task queue ──────────────────────────────────────────────────────────────

const (
	DefaultTaskQueue = "cnow-main"
)
