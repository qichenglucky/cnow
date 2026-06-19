package app

import (
	"context"
	"fmt"

	"cnow/backend/internal/domain"
	"cnow/backend/internal/pkg/audit"
	"cnow/backend/internal/pkg/errors"
	"cnow/backend/internal/repo"

	"go.uber.org/zap"
)

type CreateReleaseInput struct {
	ServiceID     int64  `json:"serviceId"`
	EnvironmentID int64  `json:"environmentId"`
	Version       string `json:"version"`
	CommitSHA     string `json:"commitSha"`
	Strategy      string `json:"strategy"`
	TriggeredBy   int64  `json:"triggeredBy"`
}

type ReleaseFilter struct {
	ServiceID *int64
	Status    *string
	domain.Pagination
}

type ReleaseApp struct {
	releases repo.ReleaseRepository
	events   repo.EventRepository
	services repo.ServiceRepository
	envs     repo.EnvironmentRepository
	audit    *audit.Writer
	log      *zap.Logger
}

func NewReleaseApp(releases repo.ReleaseRepository, events repo.EventRepository, services repo.ServiceRepository, envs repo.EnvironmentRepository, auditWriter *audit.Writer, log *zap.Logger) *ReleaseApp {
	return &ReleaseApp{releases: releases, events: events, services: services, envs: envs, audit: auditWriter, log: log.Named("app.release")}
}

func (a *ReleaseApp) CreateRelease(ctx context.Context, input CreateReleaseInput, requestID string) (*domain.Release, error) {
	svc, err := a.services.GetByID(ctx, input.ServiceID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if svc == nil {
		return nil, errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("service %d not found", input.ServiceID))
	}
	if svc.Status != domain.ServiceReady {
		return nil, errors.WithDetails(errors.ErrConflict, "service must be in 'ready' status")
	}

	env, err := a.envs.GetByID(ctx, input.EnvironmentID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if env == nil {
		return nil, errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("environment %d not found", input.EnvironmentID))
	}

	if input.Strategy == "" {
		input.Strategy = "direct"
	}

	// Check for duplicate version in the same environment
	existing, err := a.releases.GetByEnvAndVersion(ctx, input.EnvironmentID, input.Version)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if existing != nil {
		return nil, errors.WithDetails(errors.ErrAlreadyExists, fmt.Sprintf("release version '%s' already exists in environment %d", input.Version, input.EnvironmentID))
	}

	rel := &domain.Release{
		ServiceID:     input.ServiceID,
		EnvironmentID: input.EnvironmentID,
		Version:       input.Version,
		CommitSHA:     input.CommitSHA,
		Strategy:      input.Strategy,
		Status:        domain.ReleaseCreated,
		RiskLevel:     "unknown",
		TriggeredBy:   input.TriggeredBy,
	}

	if err := a.releases.Create(ctx, rel); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}

	_ = a.events.Create(ctx, &domain.ReleaseEvent{
		ReleaseID:    rel.ID,
		EventType:    "release_created",
		StatusBefore: "",
		StatusAfter:  string(domain.ReleaseCreated),
		Payload:      map[string]interface{}{"version": input.Version, "strategy": input.Strategy},
		CreatedBy:    input.TriggeredBy,
	})

	a.log.Info("release created", zap.Int64("id", rel.ID), zap.String("version", rel.Version))
	return rel, nil
}

func (a *ReleaseApp) GetRelease(ctx context.Context, id int64) (*domain.Release, []domain.ReleaseEvent, error) {
	rel, events, err := a.releases.GetByIDWithEvents(ctx, id)
	if err != nil {
		return nil, nil, errors.Wrap(errors.ErrInternal, err)
	}
	if rel == nil {
		return nil, nil, errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("release %d not found", id))
	}
	return rel, events, nil
}

func (a *ReleaseApp) ListReleases(ctx context.Context, filter ReleaseFilter) (*domain.PagedResult[domain.Release], error) {
	filter.Normalize()
	var items []domain.Release
	var total int
	var err error

	if filter.ServiceID != nil {
		items, total, err = a.releases.ListByService(ctx, *filter.ServiceID, filter.Pagination)
	} else if filter.Status != nil {
		items, total, err = a.releases.ListByStatus(ctx, domain.ReleaseStatus(*filter.Status), filter.Pagination)
	} else {
		items, total, err = a.releases.List(ctx, filter.Pagination)
	}
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	return &domain.PagedResult[domain.Release]{Items: items, Total: total, Offset: filter.Offset, Limit: filter.Limit}, nil
}

func (a *ReleaseApp) UpdateReleaseStatus(ctx context.Context, id int64, newStatus domain.ReleaseStatus, actorID int64, requestID string) error {
	rel, err := a.releases.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	if rel == nil {
		return errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("release %d not found", id))
	}
	if !rel.Status.CanTransitionTo(newStatus) {
		return errors.WithDetails(errors.ErrConflict, fmt.Sprintf("cannot transition from '%s' to '%s'", rel.Status, newStatus))
	}

	beforeStatus := rel.Status
	if err := a.releases.UpdateStatus(ctx, id, newStatus); err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}

	_ = a.events.Create(ctx, &domain.ReleaseEvent{
		ReleaseID:    id,
		EventType:    "status_change",
		StatusBefore: string(beforeStatus),
		StatusAfter:  string(newStatus),
		Payload:      map[string]interface{}{"actor": actorID},
		CreatedBy:    actorID,
	})

	a.audit.Log(audit.Entry{
		ActorID:      actorID,
		ActorRole:    "user",
		Action:       "update_status",
		ResourceType: "release",
		ResourceID:   id,
		RequestID:    requestID,
		Detail:       map[string]interface{}{"before": map[string]string{"status": string(beforeStatus)}, "after": map[string]string{"status": string(newStatus)}},
		Result:       "success",
	})

	return nil
}
