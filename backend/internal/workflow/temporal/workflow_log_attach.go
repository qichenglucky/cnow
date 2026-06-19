package temporal

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// LogAttachResult is the workflow return value.
type LogAttachResult struct {
	LogstoreName  string `json:"logstoreName"`
	DashboardURL  string `json:"dashboardUrl"`
	Status        string `json:"status"`
}

// LogAttachWorkflow orchestrates logstore creation, agent config, dashboard, and verification.
func LogAttachWorkflow(ctx workflow.Context, input AttachLogInput) (LogAttachResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("LogAttachWorkflow started", "serviceId", input.ServiceID, "envId", input.EnvID)

	result := LogAttachResult{}

	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         StandardRetry,
	}
	ctx = workflow.WithActivityOptions(ctx, actOpts)

	// 1. Create logstore
	var logstoreOut CreateLogstoreOutput
	if err := workflow.ExecuteActivity(ctx, ActivityCreateLogstore, CreateLogstoreInput{
		ServiceID:     input.ServiceID,
		EnvironmentID: input.EnvID,
	}).Get(ctx, &logstoreOut); err != nil {
		return result, err
	}
	result.LogstoreName = logstoreOut.LogstoreName

	// 2. Configure log agent
	var agentOut ConfigLogAgentOutput
	if err := workflow.ExecuteActivity(ctx, ActivityConfigLogAgent, ConfigLogAgentInput{
		ServiceID:    input.ServiceID,
		LogstoreName: logstoreOut.LogstoreName,
	}).Get(ctx, &agentOut); err != nil {
		return result, err
	}

	// 3. Create log dashboard
	var dashOut CreateLogDashboardOutput
	if err := workflow.ExecuteActivity(ctx, ActivityCreateLogDashboard, CreateLogDashboardInput{
		ServiceID:     input.ServiceID,
		EnvironmentID: input.EnvID,
	}).Get(ctx, &dashOut); err != nil {
		return result, err
	}
	result.DashboardURL = dashOut.DashboardURL

	// 4. Verify log ingestion
	var verifyOut VerifyLogOutput
	if err := workflow.ExecuteActivity(ctx, ActivityVerifyLog, VerifyLogInput{
		ServiceID:    input.ServiceID,
		LogstoreName: logstoreOut.LogstoreName,
	}).Get(ctx, &verifyOut); err != nil {
		return result, err
	}

	result.Status = "attached"
	return result, nil
}
