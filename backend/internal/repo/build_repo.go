package repo

import (
	"context"
	"fmt"

	"cnow/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BuildRepository interface {
	Create(ctx context.Context, b *domain.Build) error
	GetByID(ctx context.Context, id int64) (*domain.Build, error)
	ListByPipeline(ctx context.Context, pipelineID int64, p domain.Pagination) ([]domain.Build, int, error)
	UpdateStatus(ctx context.Context, id int64, status domain.BuildStatus) error
}

type buildRepoPG struct {
	db *pgxpool.Pool
}

func NewBuildRepository(db *pgxpool.Pool) BuildRepository {
	return &buildRepoPG{db: db}
}

func (r *buildRepoPG) Create(ctx context.Context, b *domain.Build) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO build (pipeline_id, commit_sha, branch, status)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		b.PipelineID, b.CommitSHA, b.Branch, b.Status,
	).Scan(&b.ID)
}

func (r *buildRepoPG) GetByID(ctx context.Context, id int64) (*domain.Build, error) {
	b := &domain.Build{}
	err := r.db.QueryRow(ctx,
		`SELECT id, pipeline_id, commit_sha, branch, status, started_at, finished_at, artifact_url
		 FROM build WHERE id = $1`, id,
	).Scan(&b.ID, &b.PipelineID, &b.CommitSHA, &b.Branch, &b.Status, &b.StartedAt, &b.FinishedAt, &b.ArtifactURL)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return b, err
}

func (r *buildRepoPG) ListByPipeline(ctx context.Context, pipelineID int64, p domain.Pagination) ([]domain.Build, int, error) {
	p.Normalize()
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM build WHERE pipeline_id = $1`, pipelineID).Scan(&total)

	rows, err := r.db.Query(ctx,
		`SELECT id, pipeline_id, commit_sha, branch, status, started_at, finished_at, artifact_url
		 FROM build WHERE pipeline_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3`,
		pipelineID, p.Limit, p.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list builds by pipeline: %w", err)
	}
	defer rows.Close()

	var items []domain.Build
	for rows.Next() {
		var b domain.Build
		if err := rows.Scan(&b.ID, &b.PipelineID, &b.CommitSHA, &b.Branch, &b.Status, &b.StartedAt, &b.FinishedAt, &b.ArtifactURL); err != nil {
			return nil, 0, fmt.Errorf("scan build: %w", err)
		}
		items = append(items, b)
	}
	return items, total, rows.Err()
}

func (r *buildRepoPG) UpdateStatus(ctx context.Context, id int64, status domain.BuildStatus) error {
	tag, err := r.db.Exec(ctx, `UPDATE build SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("build %d not found", id)
	}
	return nil
}
