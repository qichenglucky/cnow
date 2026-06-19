package app

import (
	"context"
	"fmt"

	"cnow/backend/internal/domain"
	"cnow/backend/internal/pkg/errors"
	"cnow/backend/internal/repo"

	"go.uber.org/zap"
)

type ApprovalApp struct {
	approvals repo.ApprovalRepository
	releases  repo.ReleaseRepository
	events    repo.EventRepository
	log       *zap.Logger
}

func NewApprovalApp(approvals repo.ApprovalRepository, releases repo.ReleaseRepository, events repo.EventRepository, log *zap.Logger) *ApprovalApp {
	return &ApprovalApp{approvals: approvals, releases: releases, events: events, log: log.Named("app.approval")}
}

func (a *ApprovalApp) CreateApproval(ctx context.Context, releaseID, approverID int64) (*domain.Approval, error) {
	rel, err := a.releases.GetByID(ctx, releaseID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if rel == nil {
		return nil, errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("release %d not found", releaseID))
	}
	approval := &domain.Approval{
		ReleaseID:  releaseID,
		ApproverID: fmt.Sprintf("%d", approverID),
		Status:     domain.ApprovalPending,
	}
	if err := a.approvals.Create(ctx, approval); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	a.log.Info("approval created", zap.Int64("id", approval.ID), zap.Int64("release", releaseID))
	return approval, nil
}

func (a *ApprovalApp) Approve(ctx context.Context, approvalID, approverID int64, comment string) error {
	approval, err := a.approvals.GetByID(ctx, approvalID)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	if approval == nil {
		return errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("approval %d not found", approvalID))
	}
	if approval.Status != domain.ApprovalPending {
		return errors.WithDetails(errors.ErrConflict, "approval is not in pending status")
	}
	if err := a.approvals.UpdateStatus(ctx, approvalID, domain.ApprovalApproved, comment); err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	// Update release status to approved
	_ = a.releases.UpdateStatus(ctx, approval.ReleaseID, domain.ReleaseApproved)
	a.log.Info("release approved", zap.Int64("approval", approvalID), zap.Int64("release", approval.ReleaseID))
	return nil
}

func (a *ApprovalApp) Reject(ctx context.Context, approvalID, approverID int64, comment string) error {
	approval, err := a.approvals.GetByID(ctx, approvalID)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	if approval == nil {
		return errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("approval %d not found", approvalID))
	}
	if approval.Status != domain.ApprovalPending {
		return errors.WithDetails(errors.ErrConflict, "approval is not in pending status")
	}
	if err := a.approvals.UpdateStatus(ctx, approvalID, domain.ApprovalRejected, comment); err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	a.log.Info("release rejected", zap.Int64("approval", approvalID), zap.Int64("release", approval.ReleaseID))
	return nil
}

func (a *ApprovalApp) ListByRelease(ctx context.Context, releaseID int64) ([]domain.Approval, error) {
	return a.approvals.ListByRelease(ctx, releaseID)
}
