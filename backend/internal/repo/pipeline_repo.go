package repo

import (
	"context"
	"fmt"

	"cnow/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PipelineRepository interface {
	Create(ctx context.Context, p *domain.Pipeline) error
	GetByID(ctx context.Context, id int64) (*domain.Pipeline, error)
	ListByService(ctx context.Context, serviceID int64) ([]domain.Pipeline, error)
}

type pipelineRepoPG struct {
	db *pgxpool.Pool
}

func NewPipelineRepository(db *pgxpool.Pool) PipelineRepository {
	return &pipelineRepoPG{db: db}
}

func (r *pipelineRepoPG) Create(ctx context.Context, p *domain.Pipeline) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO pipeline (repo_id, service_id, config_ref, status)
		 VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		p.RepoID, p.ServiceID, p.ConfigRef, p.Status,
	).Scan(&p.ID, &p.CreatedAt)
}

func (r *pipelineRepoPG) GetByID(ctx context.Context, id int64) (*domain.Pipeline, error) {
	p := &domain.Pipeline{}
	err := r.db.QueryRow(ctx,
		`SELECT id, repo_id, service_id, config_ref, status, created_at FROM pipeline WHERE id = $1`, id,
	).Scan(&p.ID, &p.RepoID, &p.ServiceID, &p.ConfigRef, &p.Status, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *pipelineRepoPG) ListByService(ctx context.Context, serviceID int64) ([]domain.Pipeline, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, repo_id, service_id, config_ref, status, created_at
		 FROM pipeline WHERE service_id = $1 ORDER BY created_at DESC`, serviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("list pipelines by service: %w", err)
	}
	defer rows.Close()

	var items []domain.Pipeline
	for rows.Next() {
		var p domain.Pipeline
		if err := rows.Scan(&p.ID, &p.RepoID, &p.ServiceID, &p.ConfigRef, &p.Status, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan pipeline: %w", err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}
