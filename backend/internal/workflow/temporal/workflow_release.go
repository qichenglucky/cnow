package temporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ReleaseResult is the workflow return value.
type ReleaseResult struct {
	ReleaseID int64  `json:"releaseId"`
	Status    string `json:"status"`
	RiskLevel string `json:"riskLevel"`
}

// ReleaseWorkflow orchestrates the full release pipeline.
func ReleaseWorkflow(ctx workflow.Context, input ValidateReleaseInput) (ReleaseResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ReleaseWorkflow started",
		"serviceId", input.ServiceID,
		"version", input.Version,
	)

	result := ReleaseResult{}

	// Common activity options
	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         StandardRetry,
	}
	ctx = workflow.WithActivityOptions(ctx, actOpts)

	// ── Step 1: Validate release precondition ────────────────────────────
	var validateOut ValidateReleaseOutput
	if err := workflow.ExecuteActivity(ctx, ActivityValidateRelease, input).Get(ctx, &validateOut); err != nil {
		return result, temporal.NewApplicationError("validate release failed", "ValidateReleaseError", err)
	}
	if !validateOut.Valid {
		return result, temporal.NewApplicationError("release precondition not met", "PreconditionFailed", validateOut.Msg)
	}

	// ── Step 2: Risk analysis ────────────────────────────────────────────
	riskInput := RiskAnalysisInput{
		ServiceID:     input.ServiceID,
		EnvironmentID: input.EnvironmentID,
		Version:       input.Version,
	}
	var riskOut RiskAnalysisOutput
	if err := workflow.ExecuteActivity(ctx, ActivityRiskAnalysis, riskInput).Get(ctx, &riskOut); err != nil {
		return result, temporal.NewApplicationError("risk analysis failed", "RiskAnalysisError", err)
	}
	result.RiskLevel = riskOut.RiskLevel
	logger.Info("risk analysis complete", "riskLevel", riskOut.RiskLevel)

	// ── Step 3: Wait for approval if high risk ───────────────────────────
	if riskOut.RiskLevel == "high" {
		logger.Info("high risk release — waiting for approval signal")
		approvalCh := workflow.GetSignalChannel(ctx, SignalApproval)

		// Wait up to 24 hours for approval
		timerCtx, cancelTimer := workflow.WithCancel(ctx)
		deadlineTimer := workflow.NewTimer(timerCtx, 24*time.Hour)

		selector := workflow.NewSelector(ctx)
		var approved bool
		selector.AddReceive(approvalCh, func(ch workflow.ReceiveChannel, more bool) {
			var approvalPayload struct {
				Approved bool   `json:"approved"`
				By       string `json:"by"`
			}
			ch.Receive(ctx, &approvalPayload)
			approved = approvalPayload.Approved
			logger.Info("approval signal received", "approved", approved, "by", approvalPayload.By)
			cancelTimer()
		})
		selector.AddFuture(deadlineTimer, func(f workflow.Future) {
			logger.Warn("approval timed out after 24h")
			cancelTimer()
		})
		selector.Select(ctx)

		if !approved {
			result.Status = "failed"
			_ = workflow.ExecuteActivity(ctx, ActivityFinalizeRelease, FinalizeReleaseInput{
				ReleaseID: input.ServiceID, // will be actual release ID in production
				Status:    "failed",
			}).Get(ctx, nil)
			return result, temporal.NewApplicationError("release not approved", "ApprovalTimeout")
		}
	}

	// ── Step 4: Create build ─────────────────────────────────────────────
	buildInput := CreateBuildInput{
		ServiceID: input.ServiceID,
		Version:   input.Version,
	}
	var buildOut CreateBuildOutput
	if err := workflow.ExecuteActivity(ctx, ActivityCreateBuild, buildInput).Get(ctx, &buildOut); err != nil {
		return result, temporal.NewApplicationError("create build failed", "CreateBuildError", err)
	}

	// ── Step 5: Wait for build complete ──────────────────────────────────
	waitInput := WaitForBuildInput{BuildID: buildOut.BuildID}
	longActOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    1 * time.Minute,
		RetryPolicy:         LongRetry,
	}
	waitCtx := workflow.WithActivityOptions(ctx, longActOpts)
	var buildResult WaitForBuildOutput
	if err := workflow.ExecuteActivity(waitCtx, ActivityWaitForBuild, waitInput).Get(ctx, &buildResult); err != nil {
		return result, temporal.NewApplicationError("build failed", "BuildFailedError", err)
	}

	// ── Step 6: Deploy ───────────────────────────────────────────────────
	deployInput := DeployInput{
		ServiceID:     input.ServiceID,
		EnvironmentID: input.EnvironmentID,
		Version:       input.Version,
		Strategy:      "direct", // TODO: derive from release input
		ArtifactURL:   buildResult.ArtifactURL,
	}
	var deployOut DeployOutput
	if err := workflow.ExecuteActivity(ctx, ActivityDeploy, deployInput).Get(ctx, &deployOut); err != nil {
		return result, temporal.NewApplicationError("deploy failed", "DeployError", err)
	}

	// ── Step 7: Health check ─────────────────────────────────────────────
	healthInput := HealthCheckInput{
		ServiceID:     input.ServiceID,
		EnvironmentID: input.EnvironmentID,
	}
	var healthOut HealthCheckOutput
	if err := workflow.ExecuteActivity(ctx, ActivityHealthCheck, healthInput).Get(ctx, &healthOut); err != nil {
		// Health check failed — attempt auto-rollback
		logger.Error("health check failed, triggering auto-rollback")
		_, _ = doAutoRollback(ctx, input.ServiceID, input.EnvironmentID, "health check failed")
		result.Status = "rolled_back"
		return result, temporal.NewApplicationError("health check failed, rolled back", "HealthCheckFailed")
	}

	// ── Step 8: Observe window (5 min monitoring) ────────────────────────
	observeInput := ObserveInput{
		ServiceID:     input.ServiceID,
		EnvironmentID: input.EnvironmentID,
		Duration:      "5m",
	}
	var observeOut ObserveOutput
	if err := workflow.ExecuteActivity(ctx, ActivityObserve, observeInput).Get(ctx, &observeOut); err != nil {
		logger.Error("observation failed, triggering auto-rollback")
		_, _ = doAutoRollback(ctx, input.ServiceID, input.EnvironmentID, "observation window failed")
		result.Status = "rolled_back"
		return result, temporal.NewApplicationError("observation failed, rolled back", "ObserveFailed")
	}

	// ── Step 9: Auto-rollback if observation detected anomalies ──────────
	if !observeOut.Stable {
		logger.Warn("observation detected anomalies, triggering auto-rollback",
			"anomalies", observeOut.Anomalies)
		_, _ = doAutoRollback(ctx, input.ServiceID, input.EnvironmentID, observeOut.Msg)
		result.Status = "rolled_back"
		return result, temporal.NewApplicationError("unstable after deploy, rolled back", "UnstableDeploy")
	}

	// ── Step 10: Finalize release ────────────────────────────────────────
	if err := workflow.ExecuteActivity(ctx, ActivityFinalizeRelease, FinalizeReleaseInput{
		ReleaseID: input.ServiceID, // TODO: actual release ID
		Status:    "succeeded",
	}).Get(ctx, nil); err != nil {
		return result, temporal.NewApplicationError("finalize release failed", "FinalizeReleaseError", err)
	}

	result.Status = "succeeded"
	logger.Info("ReleaseWorkflow completed", "riskLevel", result.RiskLevel)
	return result, nil
}

// doAutoRollback executes the auto-rollback activity and returns success/failure.
func doAutoRollback(ctx workflow.Context, serviceID, envID int64, reason string) (AutoRollbackOutput, error) {
	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         StandardRetry,
	}
	rollbackCtx := workflow.WithActivityOptions(ctx, actOpts)
	var out AutoRollbackOutput
	err := workflow.ExecuteActivity(rollbackCtx, ActivityAutoRollback, AutoRollbackInput{
		ServiceID:     serviceID,
		EnvironmentID: envID,
		Reason:        reason,
	}).Get(ctx, &out)
	return out, err
}
