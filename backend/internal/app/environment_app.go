package app

import (
	"context"
	"fmt"

	"cnow/backend/internal/domain"
	"cnow/backend/internal/pkg/errors"
	"cnow/backend/internal/repo"

	"go.uber.org/zap"
)

type CreateEnvironmentInput struct {
	ServiceID int64  `json:"serviceId"`
	Name      string `json:"name"`
	Type      string `json:"type"`    // staging, production, preview
	Version   string `json:"version"` // e.g. v1.0.0
}

type EnvironmentApp struct {
	envs     repo.EnvironmentRepository
	services repo.ServiceRepository
	log      *zap.Logger
}

func NewEnvironmentApp(envs repo.EnvironmentRepository, services repo.ServiceRepository, log *zap.Logger) *EnvironmentApp {
	return &EnvironmentApp{envs: envs, services: services, log: log.Named("app.env")}
}

func (a *EnvironmentApp) CreateEnvironment(ctx context.Context, input CreateEnvironmentInput) (*domain.Environment, error) {
	if input.Name == "" {
		return nil, errors.WithDetails(errors.ErrInvalidParam, "environment name is required")
	}
	if input.Type == "" {
		input.Type = "staging"
	}

	svc, err := a.services.GetByID(ctx, input.ServiceID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if svc == nil {
		return nil, errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("service %d not found", input.ServiceID))
	}

	existing, err := a.envs.GetByName(ctx, input.ServiceID, input.Name)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if existing != nil {
		return nil, errors.WithDetails(errors.ErrAlreadyExists,
			fmt.Sprintf("environment '%s' already exists for service %d", input.Name, input.ServiceID))
	}

	env := &domain.Environment{
		ServiceID: input.ServiceID,
		Name:      input.Name,
		Type:      input.Type,
		Version:   input.Version,
		Status:    "creating",
	}

	if err := a.envs.Create(ctx, env); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}

	a.log.Info("environment created", zap.Int64("id", env.ID), zap.String("name", env.Name))
	return env, nil
}

func (a *EnvironmentApp) GetEnvironment(ctx context.Context, id int64) (*domain.Environment, error) {
	env, err := a.envs.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if env == nil {
		return nil, errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("environment %d not found", id))
	}
	return env, nil
}

func (a *EnvironmentApp) ListByService(ctx context.Context, serviceID int64) ([]domain.Environment, error) {
	envs, err := a.envs.ListByService(ctx, serviceID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	return envs, nil
}
