package temporal

import (
	"context"
	"fmt"

	"cnow/backend/internal/workflow"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

// Adapter implements workflow.Engine using the Temporal Go SDK.
type Adapter struct {
	client    client.Client
	taskQueue string
	logger    *zap.Logger
}

// New creates a new Temporal-backed workflow engine adapter.
func New(c client.Client, taskQueue string, logger *zap.Logger) *Adapter {
	return &Adapter{
		client:    c,
		taskQueue: taskQueue,
		logger:    logger,
	}
}

// StartRelease starts the ReleaseWorkflow.
func (a *Adapter) StartRelease(ctx context.Context, input workflow.ReleaseInput) (workflow.WorkflowHandle, error) {
	workflowID := fmt.Sprintf("release-%d-%s", input.ServiceID, input.Version)
	if input.IdempotencyKey != "" {
		workflowID = fmt.Sprintf("release-%s", input.IdempotencyKey)
	}

	opts := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: a.taskQueue,
	}

	wfInput := ValidateReleaseInput{
		ServiceID:     input.ServiceID,
		EnvironmentID: input.EnvironmentID,
		Version:       input.Version,
	}

	run, err := a.client.ExecuteWorkflow(ctx, opts, ReleaseWorkflow, wfInput)
	if err != nil {
		return workflow.WorkflowHandle{}, fmt.Errorf("start release workflow: %w", err)
	}

	a.logger.Info("release workflow started",
		zap.String("workflowId", run.GetID()),
		zap.String("runId", run.GetRunID()),
	)

	return workflow.WorkflowHandle{
		ID:     run.GetID(),
		Status: "running",
	}, nil
}

// StartServiceCreate starts the ServiceCreateWorkflow.
func (a *Adapter) StartServiceCreate(ctx context.Context, input workflow.ServiceCreateInput) (workflow.WorkflowHandle, error) {
	workflowID := fmt.Sprintf("service-create-%s", input.Name)
	if input.IdempotencyKey != "" {
		workflowID = fmt.Sprintf("service-create-%s", input.IdempotencyKey)
	}

	opts := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: a.taskQueue,
	}

	wfInput := CreateServiceInput{
		Name:        input.Name,
		DisplayName: input.DisplayName,
		TechStack:   input.TechStack,
		RepoMode:    input.RepoMode,
	}

	run, err := a.client.ExecuteWorkflow(ctx, opts, ServiceCreateWorkflow, wfInput)
	if err != nil {
		return workflow.WorkflowHandle{}, fmt.Errorf("start service-create workflow: %w", err)
	}

	a.logger.Info("service-create workflow started",
		zap.String("workflowId", run.GetID()),
		zap.String("runId", run.GetRunID()),
	)

	return workflow.WorkflowHandle{
		ID:     run.GetID(),
		Status: "running",
	}, nil
}

// StartRollback starts the RollbackWorkflow.
func (a *Adapter) StartRollback(ctx context.Context, input workflow.RollbackInput) (workflow.WorkflowHandle, error) {
	workflowID := fmt.Sprintf("rollback-%d", input.ReleaseID)

	opts := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: a.taskQueue,
	}

	wfInput := ValidateRollbackInput{
		ReleaseID: input.ReleaseID,
	}

	run, err := a.client.ExecuteWorkflow(ctx, opts, RollbackWorkflow, wfInput)
	if err != nil {
		return workflow.WorkflowHandle{}, fmt.Errorf("start rollback workflow: %w", err)
	}

	a.logger.Info("rollback workflow started",
		zap.String("workflowId", run.GetID()),
		zap.String("runId", run.GetRunID()),
	)

	return workflow.WorkflowHandle{
		ID:     run.GetID(),
		Status: "running",
	}, nil
}

// GetWorkflowStatus queries Temporal for the current status of a workflow.
func (a *Adapter) GetWorkflowStatus(ctx context.Context, workflowID string) (string, error) {
	resp, err := a.client.DescribeWorkflowExecution(ctx, workflowID, "")
	if err != nil {
		return "", fmt.Errorf("describe workflow %s: %w", workflowID, err)
	}

	status := resp.GetWorkflowExecutionInfo().GetStatus()
	return mapTemporalStatus(status), nil
}

// SignalWorkflow sends a signal to a running workflow.
func (a *Adapter) SignalWorkflow(ctx context.Context, workflowID string, signalName string, payload interface{}) error {
	err := a.client.SignalWorkflow(ctx, workflowID, "", signalName, payload)
	if err != nil {
		return fmt.Errorf("signal workflow %s (%s): %w", workflowID, signalName, err)
	}

	a.logger.Info("signal sent to workflow",
		zap.String("workflowId", workflowID),
		zap.String("signal", signalName),
	)
	return nil
}

// mapTemporalStatus converts a Temporal execution status to a human-readable string.
func mapTemporalStatus(status enums.WorkflowExecutionStatus) string {
	switch status {
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		return "running"
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return "completed"
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		return "failed"
	case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
		return "cancelled"
	case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		return "terminated"
	case enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return "continued_as_new"
	case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return "timed_out"
	default:
		return "unknown"
	}
}

// Ensure Adapter satisfies the workflow.Engine interface at compile time.
var _ workflow.Engine = (*Adapter)(nil)
