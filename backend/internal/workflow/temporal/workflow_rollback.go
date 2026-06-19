package temporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// RollbackResult is the workflow return value.
type RollbackResult struct {
	ReleaseID    int64  `json:"releaseId"`
	RollbackID   string `json:"rollbackId"`
	PrevVersion  string `json:"prevVersion"`
	Status       string `json:"status"`
}

// RollbackWorkflow orchestrates a manual or automatic rollback.
func RollbackWorkflow(ctx workflow.Context, input ValidateRollbackInput) (RollbackResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("RollbackWorkflow started", "releaseId", input.ReleaseID)

	result := RollbackResult{ReleaseID: input.ReleaseID}

	// Activity options
	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         StandardRetry,
	}
	ctx = workflow.WithActivityOptions(ctx, actOpts)

	// ── Step 1: Validate rollback ────────────────────────────────────────
	var validateOut ValidateRollbackOutput
	if err := workflow.ExecuteActivity(ctx, ActivityValidateRollback, input).Get(ctx, &validateOut); err != nil {
		return result, temporal.NewApplicationError("validate rollback failed", "ValidateRollbackError", err)
	}
	if !validateOut.Valid {
		return result, temporal.NewApplicationError("rollback not valid", "RollbackInvalid", validateOut.Msg)
	}

	// ── Step 2: Get previous version ─────────────────────────────────────
	var prevOut GetPrevVersionOutput
	if err := workflow.ExecuteActivity(ctx, ActivityGetPrevVersion, GetPrevVersionInput{
		ReleaseID: input.ReleaseID,
	}).Get(ctx, &prevOut); err != nil {
		return result, temporal.NewApplicationError("get previous version failed", "GetPrevVersionError", err)
	}
	result.PrevVersion = prevOut.PrevVersion
	logger.Info("found previous version", "prevVersion", prevOut.PrevVersion)

	// ── Step 3: Execute rollback ─────────────────────────────────────────
	var rbOut ExecuteRollbackOutput
	if err := workflow.ExecuteActivity(ctx, ActivityExecuteRollback, ExecuteRollbackInput{
		ReleaseID:   input.ReleaseID,
		PrevVersion: prevOut.PrevVersion,
		Reason:      "manual rollback",
	}).Get(ctx, &rbOut); err != nil {
		return result, temporal.NewApplicationError("execute rollback failed", "ExecuteRollbackError", err)
	}
	result.RollbackID = rbOut.RollbackID

	// ── Step 4: Verify rollback ──────────────────────────────────────────
	var verifyOut VerifyRollbackOutput
	if err := workflow.ExecuteActivity(ctx, ActivityVerifyRollback, VerifyRollbackInput{
		ReleaseID: input.ReleaseID,
		Version:   prevOut.PrevVersion,
	}).Get(ctx, &verifyOut); err != nil {
		return result, temporal.NewApplicationError("verify rollback failed", "VerifyRollbackError", err)
	}
	if !verifyOut.Verified {
		logger.Warn("rollback verification failed", "msg", verifyOut.Msg)
		result.Status = "verification_failed"
		return result, temporal.NewApplicationError("rollback verification failed", "VerifyRollbackFailed", verifyOut.Msg)
	}

	// ── Step 5: Create rollback record ───────────────────────────────────
	var recOut CreateRollbackRecordOutput
	if err := workflow.ExecuteActivity(ctx, ActivityCreateRollbackRecord, CreateRollbackRecordInput{
		ReleaseID:   input.ReleaseID,
		FromVersion: "current", // TODO: get actual current version from release
		ToVersion:   prevOut.PrevVersion,
		Reason:      "manual rollback",
	}).Get(ctx, &recOut); err != nil {
		logger.Warn("failed to create rollback record (non-fatal)", "error", err)
		// Non-fatal: the rollback itself succeeded even if record-keeping failed
	}

	result.Status = "rolled_back"
	logger.Info("RollbackWorkflow completed",
		"rollbackId", result.RollbackID,
		"prevVersion", result.PrevVersion,
	)
	return result, nil
}
