package workflow

import "context"

// Engine is the unified interface for starting and managing workflows.
type Engine interface {
	// StartRelease begins a new release workflow.
	StartRelease(ctx context.Context, input ReleaseInput) (WorkflowHandle, error)
	// StartServiceCreate begins a new service creation workflow.
	StartServiceCreate(ctx context.Context, input ServiceCreateInput) (WorkflowHandle, error)
	// StartRollback begins a new rollback workflow.
	StartRollback(ctx context.Context, input RollbackInput) (WorkflowHandle, error)
	// GetWorkflowStatus returns the current status of a workflow by its ID.
	GetWorkflowStatus(ctx context.Context, workflowID string) (string, error)
	// SignalWorkflow sends a signal with an optional payload to a running workflow.
	SignalWorkflow(ctx context.Context, workflowID string, signalName string, payload interface{}) error
}

// WorkflowHandle is the reference returned when a workflow is started.
type WorkflowHandle struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ReleaseInput describes the parameters for starting a release workflow.
type ReleaseInput struct {
	ServiceID     int64  `json:"serviceId"`
	EnvironmentID int64  `json:"environmentId"`
	Version       string `json:"version"`
	Strategy      string `json:"strategy"` // direct, canary, blue_green
	TriggeredBy   string `json:"triggeredBy"`
	IdempotencyKey string `json:"idempotencyKey"`
}

// ServiceCreateInput describes the parameters for starting a service creation workflow.
type ServiceCreateInput struct {
	Name           string `json:"name"`
	DisplayName    string `json:"displayName"`
	TechStack      string `json:"techStack"`
	RepoMode       string `json:"repoMode"`
	IdempotencyKey string `json:"idempotencyKey"`
}

// RollbackInput describes the parameters for starting a rollback workflow.
type RollbackInput struct {
	ReleaseID   int64  `json:"releaseId"`
	Reason      string `json:"reason"`
	TriggeredBy string `json:"triggeredBy"`
}
