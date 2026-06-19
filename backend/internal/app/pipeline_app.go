package app

import (
	"context"
	"fmt"

	"cnow/backend/internal/domain"
	"cnow/backend/internal/pkg/errors"
	"cnow/backend/internal/repo"

	"go.uber.org/zap"
)

// PipelineApp manages pipelines and builds.
type PipelineApp struct {
	pipelines repo.PipelineRepository
	builds    repo.BuildRepository
	log       *zap.Logger
}

func NewPipelineApp(pipelines repo.PipelineRepository, builds repo.BuildRepository, log *zap.Logger) *PipelineApp {
	return &PipelineApp{pipelines: pipelines, builds: builds, log: log.Named("app.pipeline")}
}

// --- Pipeline ---

type CreatePipelineInput struct {
	ServiceID int64  `json:"serviceId"`
	ConfigRef string `json:"configRef"`
}

func (a *PipelineApp) CreatePipeline(ctx context.Context, input CreatePipelineInput) (*domain.Pipeline, error) {
	if input.ServiceID == 0 {
		return nil, errors.WithDetails(errors.ErrInvalidParam, "serviceId is required")
	}

	p := &domain.Pipeline{
		ServiceID: input.ServiceID,
		ConfigRef: input.ConfigRef,
		Status:    "active",
	}

	if err := a.pipelines.Create(ctx, p); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}

	a.log.Info("pipeline created", zap.Int64("id", p.ID), zap.Int64("serviceId", p.ServiceID))
	return p, nil
}

func (a *PipelineApp) GetPipeline(ctx context.Context, id int64) (*domain.Pipeline, error) {
	p, err := a.pipelines.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if p == nil {
		return nil, errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("pipeline %d not found", id))
	}
	return p, nil
}

// --- Build ---

type CreateBuildInput struct {
	PipelineID int64  `json:"pipelineId"`
	CommitSHA  string `json:"commitSha"`
	Branch     string `json:"branch"`
}

type BuildFilter struct {
	PipelineID *int64
	domain.Pagination
}

func (a *PipelineApp) CreateBuild(ctx context.Context, input CreateBuildInput) (*domain.Build, error) {
	if input.PipelineID == 0 {
		return nil, errors.WithDetails(errors.ErrInvalidParam, "pipelineId is required")
	}
	if input.CommitSHA == "" {
		return nil, errors.WithDetails(errors.ErrInvalidParam, "commitSha is required")
	}

	// Verify pipeline exists
	p, err := a.pipelines.GetByID(ctx, input.PipelineID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if p == nil {
		return nil, errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("pipeline %d not found", input.PipelineID))
	}

	b := &domain.Build{
		PipelineID: input.PipelineID,
		CommitSHA:  input.CommitSHA,
		Branch:     input.Branch,
		Status:     domain.BuildPending,
	}

	if err := a.builds.Create(ctx, b); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}

	a.log.Info("build created", zap.Int64("id", b.ID), zap.String("commit", input.CommitSHA))
	return b, nil
}

func (a *PipelineApp) UpdateBuildStatus(ctx context.Context, id int64, status domain.BuildStatus) error {
	b, err := a.builds.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	if b == nil {
		return errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("build %d not found", id))
	}

	if err := a.builds.UpdateStatus(ctx, id, status); err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}

	a.log.Info("build status updated", zap.Int64("id", id), zap.String("status", string(status)))
	return nil
}

func (a *PipelineApp) ListBuilds(ctx context.Context, filter BuildFilter) (*domain.PagedResult[domain.Build], error) {
	if filter.PipelineID == nil {
		return nil, errors.WithDetails(errors.ErrInvalidParam, "pipeline_id is required")
	}
	items, total, err := a.builds.ListByPipeline(ctx, *filter.PipelineID, filter.Pagination)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	return &domain.PagedResult[domain.Build]{
		Items:  items,
		Total:  total,
		Offset: filter.Offset,
		Limit:  filter.Limit,
	}, nil
}
