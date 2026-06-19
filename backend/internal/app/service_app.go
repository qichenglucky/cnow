package app

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"cnow/backend/internal/domain"
	"cnow/backend/internal/pkg/audit"
	"cnow/backend/internal/pkg/errors"
	"cnow/backend/internal/repo"

	"go.uber.org/zap"
)

var serviceNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*$`)

// CreateServiceInput is the input for creating a service.
type CreateServiceInput struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	TechStack   string `json:"techStack"`
	OwnerID     int64  `json:"ownerId"`
}

// ServiceFilter holds filters for listing services.
type ServiceFilter struct {
	domain.Pagination
}

// ServiceApp implements the service application layer.
type ServiceApp struct {
	services repo.ServiceRepository
	audit    *audit.Writer
	log      *zap.Logger
}

func NewServiceApp(services repo.ServiceRepository, auditWriter *audit.Writer, log *zap.Logger) *ServiceApp {
	return &ServiceApp{services: services, audit: auditWriter, log: log.Named("app.service")}
}

// CreateService creates a new service with validation and audit.
func (a *ServiceApp) CreateService(ctx context.Context, input CreateServiceInput, requestID string) (*domain.Service, error) {
	if input.Name == "" {
		return nil, errors.WithDetails(errors.ErrInvalidParam, "name is required")
	}
	if len(input.Name) > 128 {
		return nil, errors.WithDetails(errors.ErrInvalidParam, "name must be 1-128 characters")
	}
	if !serviceNameRe.MatchString(input.Name) {
		return nil, errors.WithDetails(errors.ErrInvalidParam, "name must start with a lowercase letter or digit and contain only lowercase letters, digits, hyphens, and dots")
	}
	if input.TechStack == "" {
		input.TechStack = "unknown"
	}

	// Check uniqueness
	existing, err := a.services.GetByName(ctx, input.Name)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if existing != nil {
		return nil, errors.WithDetails(errors.ErrAlreadyExists, fmt.Sprintf("service '%s' already exists", input.Name))
	}

	svc := &domain.Service{
		Name:        input.Name,
		DisplayName: input.DisplayName,
		Description: input.Description,
		TechStack:   input.TechStack,
		Status:      domain.ServiceDraft,
		OwnerID:     input.OwnerID,
	}

	if err := a.services.Create(ctx, svc); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}

	a.audit.Log(audit.Entry{
		ActorID:      input.OwnerID,
		ActorRole:    "user",
		Action:       "create",
		ResourceType: "service",
		ResourceID:   svc.ID,
		RequestID:    requestID,
		Detail:       map[string]interface{}{"after": svc},
		Result:       "success",
	})

	a.log.Info("service created", zap.Int64("id", svc.ID), zap.String("name", svc.Name))
	return svc, nil
}

// GetService returns a service by ID.
func (a *ServiceApp) GetService(ctx context.Context, id int64) (*domain.Service, error) {
	svc, err := a.services.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	if svc == nil {
		return nil, errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("service %d not found", id))
	}
	return svc, nil
}

// ListServices returns a paginated list of services.
func (a *ServiceApp) ListServices(ctx context.Context, filter ServiceFilter) (*domain.PagedResult[domain.Service], error) {
	items, total, err := a.services.List(ctx, filter.Pagination)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, err)
	}
	return &domain.PagedResult[domain.Service]{
		Items:  items,
		Total:  total,
		Offset: filter.Offset,
		Limit:  filter.Limit,
	}, nil
}

// UpdateServiceStatus validates and performs a status transition.
func (a *ServiceApp) UpdateServiceStatus(ctx context.Context, id int64, newStatus domain.ServiceStatus, actorID, requestID string) error {
	svc, err := a.services.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	if svc == nil {
		return errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("service %d not found", id))
	}

	if !svc.Status.CanTransitionTo(newStatus) {
		return errors.WithDetails(errors.ErrConflict,
			fmt.Sprintf("cannot transition from '%s' to '%s'", svc.Status, newStatus))
	}

	beforeStatus := svc.Status
	if err := a.services.UpdateStatus(ctx, id, newStatus); err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}

	actorIDInt, _ := strconv.ParseInt(actorID, 10, 64)
	a.audit.Log(audit.Entry{
		ActorID:      actorIDInt,
		ActorRole:    "user",
		Action:       "update_status",
		ResourceType: "service",
		ResourceID:   id,
		RequestID:    requestID,
		Detail:       map[string]interface{}{"before": map[string]string{"status": string(beforeStatus)}, "after": map[string]string{"status": string(newStatus)}},
		Result:       "success",
	})

	return nil
}


// UpdateService updates a service's editable fields.
func (a *ServiceApp) UpdateService(ctx context.Context, svc *domain.Service) error {
	if svc.ID == 0 {
		return errors.WithDetails(errors.ErrInvalidParam, "service id is required")
	}
	existing, err := a.services.GetByID(ctx, svc.ID)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	if existing == nil {
		return errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("service %d not found", svc.ID))
	}
	if err := a.services.Update(ctx, svc); err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	return nil
}

// DeleteService archives a service (soft delete).
func (a *ServiceApp) DeleteService(ctx context.Context, id int64) error {
	existing, err := a.services.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	if existing == nil {
		return errors.WithDetails(errors.ErrNotFound, fmt.Sprintf("service %d not found", id))
	}
	if err := a.services.UpdateStatus(ctx, id, domain.ServiceArchived); err != nil {
		return errors.Wrap(errors.ErrInternal, err)
	}
	a.log.Info("service archived", zap.Int64("id", id))
	return nil
}
