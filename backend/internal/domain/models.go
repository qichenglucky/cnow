package domain

import "time"

// --- Enums ---

type ServiceStatus string

const (
	ServiceDraft    ServiceStatus = "draft"
	ServiceCreating ServiceStatus = "creating"
	ServiceReady    ServiceStatus = "ready"
	ServiceDegraded ServiceStatus = "degraded"
	ServiceArchived ServiceStatus = "archived"
)

var ServiceTransitions = map[ServiceStatus][]ServiceStatus{
	ServiceDraft:    {ServiceCreating, ServiceArchived},
	ServiceCreating: {ServiceReady, ServiceDegraded, ServiceArchived},
	ServiceReady:    {ServiceDegraded, ServiceArchived},
	ServiceDegraded: {ServiceReady, ServiceArchived},
}

type ReleaseStatus string

const (
	ReleaseCreated        ReleaseStatus = "created"
	ReleaseReviewing      ReleaseStatus = "reviewing"
	ReleaseApproved       ReleaseStatus = "approved"
	ReleaseDeploying      ReleaseStatus = "deploying"
	ReleaseVerifying      ReleaseStatus = "verifying"
	ReleaseObserving      ReleaseStatus = "observing"
	ReleaseSucceeded      ReleaseStatus = "succeeded"
	ReleaseFailed         ReleaseStatus = "failed"
	ReleaseRollbackPending ReleaseStatus = "rollback_pending"
	ReleaseRollingBack    ReleaseStatus = "rolling_back"
	ReleaseRolledBack     ReleaseStatus = "rolled_back"
)

var ReleaseTransitions = map[ReleaseStatus][]ReleaseStatus{
	ReleaseCreated:        {ReleaseReviewing, ReleaseDeploying},
	ReleaseReviewing:      {ReleaseApproved, ReleaseFailed},
	ReleaseApproved:       {ReleaseDeploying, ReleaseFailed},
	ReleaseDeploying:      {ReleaseVerifying, ReleaseFailed, ReleaseRollbackPending},
	ReleaseVerifying:      {ReleaseObserving, ReleaseFailed, ReleaseRollbackPending},
	ReleaseObserving:      {ReleaseSucceeded, ReleaseRollbackPending},
	ReleaseSucceeded:      {ReleaseRollbackPending},
	ReleaseRollbackPending: {ReleaseRollingBack},
	ReleaseRollingBack:    {ReleaseRolledBack, ReleaseFailed},
}

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
	ApprovalExpired  ApprovalStatus = "expired"
)

type BuildStatus string

const (
	BuildPending   BuildStatus = "pending"
	BuildRunning   BuildStatus = "running"
	BuildSucceeded BuildStatus = "succeeded"
	BuildFailed    BuildStatus = "failed"
	BuildCancelled BuildStatus = "cancelled"
)

type IncidentSeverity string

const (
	SeverityLow      IncidentSeverity = "low"
	SeverityMedium   IncidentSeverity = "medium"
	SeverityHigh     IncidentSeverity = "high"
	SeverityCritical IncidentSeverity = "critical"
)

// --- Models ---

type Service struct {
	ID          int64         `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"`
	DisplayName string        `json:"displayName" db:"display_name"`
	Description string        `json:"description" db:"description"`
	TechStack   string        `json:"techStack" db:"tech_stack"`
	Status      ServiceStatus `json:"status" db:"status"`
	OwnerID     int64        `json:"ownerId" db:"owner_id"`
	CreatedAt   time.Time     `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time     `json:"updatedAt" db:"updated_at"`
}

type Repo struct {
	ID           int64     `json:"id" db:"id"`
	ServiceID    int64     `json:"serviceId" db:"service_id"`
	Provider     string    `json:"provider" db:"provider"`
	URL          string    `json:"url" db:"url"`
	DefaultBranch string   `json:"defaultBranch" db:"default_branch"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

type Pipeline struct {
	ID         int64     `json:"id" db:"id"`
	RepoID     int64     `json:"repoId" db:"repo_id"`
	ServiceID  int64     `json:"serviceId" db:"service_id"`
	ConfigRef  string    `json:"configRef" db:"config_ref"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

type Build struct {
	ID          int64       `json:"id" db:"id"`
	PipelineID  int64       `json:"pipelineId" db:"pipeline_id"`
	CommitSHA   string      `json:"commitSha" db:"commit_sha"`
	Branch      string      `json:"branch" db:"branch"`
	Status      BuildStatus `json:"status" db:"status"`
	StartedAt   *time.Time  `json:"startedAt,omitempty" db:"started_at"`
	FinishedAt  *time.Time  `json:"finishedAt,omitempty" db:"finished_at"`
	ArtifactURL string      `json:"artifactUrl,omitempty" db:"artifact_url"`
}

type Environment struct {
	ID            int64     `json:"id" db:"id"`
	ServiceID     int64     `json:"serviceId" db:"service_id"`
	Name          string    `json:"name" db:"name"`
	Type          string    `json:"type" db:"type"`
	Version       string    `json:"version" db:"version"`
	Status        string    `json:"status" db:"status"`
	DomainID      *int64    `json:"domainId,omitempty" db:"domain_id"`
	LogSourceID   *int64    `json:"logSourceId,omitempty" db:"log_source_id"`
	MetricPanelID *int64    `json:"metricPanelId,omitempty" db:"metric_panel_id"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

type Domain struct {
	ID            int64     `json:"id" db:"id"`
	EnvironmentID int64     `json:"environmentId" db:"environment_id"`
	FQDN          string    `json:"fqdn" db:"fqdn"`
	CertID        *int64    `json:"certId,omitempty" db:"cert_id"`
	Status        string    `json:"status" db:"status"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

type Certificate struct {
	ID        int64      `json:"id" db:"id"`
	DomainID  int64      `json:"domainId" db:"domain_id"`
	Issuer    string     `json:"issuer" db:"issuer"`
	Status    string     `json:"status" db:"status"`
	NotBefore *time.Time `json:"notBefore,omitempty" db:"not_before"`
	NotAfter  *time.Time `json:"notAfter,omitempty" db:"not_after"`
}

type LogSource struct {
	ID            int64  `json:"id" db:"id"`
	ServiceID     int64  `json:"serviceId" db:"service_id"`
	EnvironmentID int64  `json:"environmentId" db:"environment_id"`
	Provider      string `json:"provider" db:"provider"`
	Config        string `json:"config" db:"config"`
	Status        string `json:"status" db:"status"`
}

type MetricPanel struct {
	ID            int64  `json:"id" db:"id"`
	ServiceID     int64  `json:"serviceId" db:"service_id"`
	EnvironmentID int64  `json:"environmentId" db:"environment_id"`
	DashboardURL  string `json:"dashboardUrl" db:"dashboard_url"`
	Status        string `json:"status" db:"status"`
}

type AlertRule struct {
	ID            int64  `json:"id" db:"id"`
	ServiceID     int64  `json:"serviceId" db:"service_id"`
	EnvironmentID int64  `json:"environmentId" db:"environment_id"`
	Name          string `json:"name" db:"name"`
	Condition     string `json:"condition" db:"condition"`
	Severity      string `json:"severity" db:"severity"`
	Status        string `json:"status" db:"status"`
}

type Release struct {
	ID            int64         `json:"id" db:"id"`
	ServiceID     int64         `json:"serviceId" db:"service_id"`
	EnvironmentID int64         `json:"environmentId" db:"environment_id"`
	Version       string        `json:"version" db:"version"`
	CommitSHA     string        `json:"commitSha" db:"commit_sha"`
	ImageTag      string        `json:"imageTag" db:"image_tag"`
	Strategy      string        `json:"strategy" db:"strategy"`
	Status        ReleaseStatus `json:"status" db:"status"`
	TriggeredBy   int64         `json:"triggeredBy" db:"triggered_by"`
	ApprovedBy    *int64        `json:"approvedBy,omitempty" db:"approved_by"`
	StartedAt     *time.Time    `json:"startedAt,omitempty" db:"started_at"`
	FinishedAt    *time.Time    `json:"finishedAt,omitempty" db:"finished_at"`
	Summary       string        `json:"summary" db:"summary"`
	RiskLevel     string        `json:"riskLevel" db:"risk_level"`
}

type Approval struct {
	ID         int64          `json:"id" db:"id"`
	ReleaseID  int64          `json:"releaseId" db:"release_id"`
	ApproverID string         `json:"approverId" db:"approver_id"`
	Status     ApprovalStatus `json:"status" db:"status"`
	Comment    string         `json:"comment" db:"comment"`
	CreatedAt  time.Time      `json:"createdAt" db:"created_at"`
}

type RollbackRecord struct {
	ID          int64     `json:"id" db:"id"`
	ReleaseID   int64     `json:"releaseId" db:"release_id"`
	FromVersion string    `json:"fromVersion" db:"from_version"`
	ToVersion   string    `json:"toVersion" db:"to_version"`
	Reason      string    `json:"reason" db:"reason"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

type ReleaseEvent struct {
	ID          int64     `json:"id" db:"id"`
	ReleaseID   int64     `json:"releaseId" db:"release_id"`
	EventType   string    `json:"eventType" db:"event_type"`
	StatusBefore string   `json:"statusBefore" db:"status_before"`
	StatusAfter string    `json:"statusAfter" db:"status_after"`
	Payload     any       `json:"payload,omitempty" db:"payload"`
	CreatedBy   int64     `json:"createdBy" db:"created_by"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

type Incident struct {
	ID            int64             `json:"id" db:"id"`
	ServiceID     int64             `json:"serviceId" db:"service_id"`
	EnvironmentID int64             `json:"environmentId" db:"environment_id"`
	ReleaseID     *int64            `json:"releaseId,omitempty" db:"release_id"`
	Title         string            `json:"title" db:"title"`
	Severity      IncidentSeverity  `json:"severity" db:"severity"`
	Status        string            `json:"status" db:"status"`
	CreatedAt     time.Time         `json:"createdAt" db:"created_at"`
}

type AuditLog struct {
	ID           int64      `json:"id" db:"id"`
	ActorID      int64      `json:"actorId" db:"actor_id"`
	ActorRole    string     `json:"actorRole" db:"actor_role"`
	Action       string     `json:"action" db:"action"`
	ResourceType string     `json:"resourceType" db:"resource_type"`
	ResourceID   int64      `json:"resourceId" db:"resource_id"`
	RequestID    string     `json:"requestId" db:"request_id"`
	Detail       any        `json:"detail,omitempty" db:"detail"`
	Result       string     `json:"result" db:"result"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
}

type AIRun struct {
	ID          int64     `json:"id" db:"id"`
	ServiceID   int64     `json:"serviceId" db:"service_id"`
	Type        string    `json:"type" db:"type"`
	InputRef    string    `json:"inputRef" db:"input_ref"`
	OutputRef   string    `json:"outputRef" db:"output_ref"`
	RiskLevel   string    `json:"riskLevel" db:"risk_level"`
	ModelUsed   string    `json:"modelUsed" db:"model_used"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

// --- Validation helpers ---

// CanTransitionTo checks if a service status transition is valid.
func (s ServiceStatus) CanTransitionTo(target ServiceStatus) bool {
	allowed, ok := ServiceTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}

// CanTransitionTo checks if a release status transition is valid.
func (r ReleaseStatus) CanTransitionTo(target ReleaseStatus) bool {
	allowed, ok := ReleaseTransitions[r]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}

// --- Pagination ---

type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

func (p *Pagination) Normalize() {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
}

type PagedResult[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
