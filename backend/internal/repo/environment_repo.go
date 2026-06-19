package repo

import (
	"context"
	"fmt"

	"cnow/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EnvironmentRepository interface {
	Create(ctx context.Context, e *domain.Environment) error
	GetByID(ctx context.Context, id int64) (*domain.Environment, error)
	GetByName(ctx context.Context, serviceID int64, name string) (*domain.Environment, error)
	ListByService(ctx context.Context, serviceID int64) ([]domain.Environment, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}

type envRepoPG struct {
	db *pgxpool.Pool
}

func NewEnvironmentRepository(db *pgxpool.Pool) EnvironmentRepository {
	return &envRepoPG{db: db}
}

func (r *envRepoPG) Create(ctx context.Context, e *domain.Environment) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO environment (service_id, name, type, version, status)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		e.ServiceID, e.Name, e.Type, e.Version, e.Status,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func (r *envRepoPG) GetByID(ctx context.Context, id int64) (*domain.Environment, error) {
	e := &domain.Environment{}
	err := r.db.QueryRow(ctx,
		`SELECT id, service_id, name, type, version, status, domain_id, log_source_id, metric_panel_id, created_at, updated_at
		 FROM environment WHERE id = $1`, id,
	).Scan(&e.ID, &e.ServiceID, &e.Name, &e.Type, &e.Version, &e.Status, &e.DomainID, &e.LogSourceID, &e.MetricPanelID, &e.CreatedAt, &e.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return e, err
}

func (r *envRepoPG) GetByName(ctx context.Context, serviceID int64, name string) (*domain.Environment, error) {
	e := &domain.Environment{}
	err := r.db.QueryRow(ctx,
		`SELECT id, service_id, name, type, version, status, domain_id, log_source_id, metric_panel_id, created_at, updated_at
		 FROM environment WHERE service_id = $1 AND name = $2`, serviceID, name,
	).Scan(&e.ID, &e.ServiceID, &e.Name, &e.Type, &e.Version, &e.Status, &e.DomainID, &e.LogSourceID, &e.MetricPanelID, &e.CreatedAt, &e.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return e, err
}

func (r *envRepoPG) ListByService(ctx context.Context, serviceID int64) ([]domain.Environment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, service_id, name, type, version, status, domain_id, log_source_id, metric_panel_id, created_at, updated_at
		 FROM environment WHERE service_id = $1 ORDER BY id`, serviceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Environment
	for rows.Next() {
		var e domain.Environment
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.Name, &e.Type, &e.Version, &e.Status, &e.DomainID, &e.LogSourceID, &e.MetricPanelID, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

func (r *envRepoPG) UpdateStatus(ctx context.Context, id int64, status string) error {
	tag, err := r.db.Exec(ctx, `UPDATE environment SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("environment %d not found", id)
	}
	return nil
}
