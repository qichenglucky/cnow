package repo

import (
	"context"
	"fmt"

	"cnow/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReleaseRepository interface {
	Create(ctx context.Context, r *domain.Release) error
	GetByID(ctx context.Context, id int64) (*domain.Release, error)
	GetByIDWithEvents(ctx context.Context, id int64) (*domain.Release, []domain.ReleaseEvent, error)
	GetByEnvAndVersion(ctx context.Context, envID int64, version string) (*domain.Release, error)
	ListByService(ctx context.Context, serviceID int64, p domain.Pagination) ([]domain.Release, int, error)
	List(ctx context.Context, p domain.Pagination) ([]domain.Release, int, error)
	ListByStatus(ctx context.Context, status domain.ReleaseStatus, p domain.Pagination) ([]domain.Release, int, error)
	UpdateStatus(ctx context.Context, id int64, status domain.ReleaseStatus) error
}

type releaseRepoPG struct {
	db *pgxpool.Pool
}

func NewReleaseRepository(db *pgxpool.Pool) ReleaseRepository {
	return &releaseRepoPG{db: db}
}

const releaseCols = `id, service_id, environment_id, version, commit_sha, image_tag, strategy, status, triggered_by, approved_by, started_at, finished_at, summary, risk_level`

func (r *releaseRepoPG) scanRelease(row interface{ Scan(...interface{}) error }) (*domain.Release, error) {
	rel := &domain.Release{}
	err := row.Scan(&rel.ID, &rel.ServiceID, &rel.EnvironmentID, &rel.Version, &rel.CommitSHA, &rel.ImageTag, &rel.Strategy, &rel.Status, &rel.TriggeredBy, &rel.ApprovedBy, &rel.StartedAt, &rel.FinishedAt, &rel.Summary, &rel.RiskLevel)
	return rel, err
}

func (r *releaseRepoPG) Create(ctx context.Context, rel *domain.Release) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO "release" (service_id, environment_id, version, commit_sha, image_tag, strategy, status, triggered_by, summary, risk_level)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		rel.ServiceID, rel.EnvironmentID, rel.Version, rel.CommitSHA, rel.ImageTag, rel.Strategy, rel.Status, rel.TriggeredBy, rel.Summary, rel.RiskLevel,
	).Scan(&rel.ID)
}

func (r *releaseRepoPG) GetByID(ctx context.Context, id int64) (*domain.Release, error) {
	row := r.db.QueryRow(ctx, `SELECT `+releaseCols+` FROM "release" WHERE id = $1`, id)
	rel, err := r.scanRelease(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rel, err
}

func (r *releaseRepoPG) GetByEnvAndVersion(ctx context.Context, envID int64, version string) (*domain.Release, error) {
	row := r.db.QueryRow(ctx, `SELECT `+releaseCols+` FROM "release" WHERE environment_id = $1 AND version = $2`, envID, version)
	rel, err := r.scanRelease(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rel, err
}

func (r *releaseRepoPG) GetByIDWithEvents(ctx context.Context, id int64) (*domain.Release, []domain.ReleaseEvent, error) {
	rel, err := r.GetByID(ctx, id)
	if err != nil || rel == nil {
		return rel, nil, err
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, release_id, event_type, status_before, status_after, payload, created_by, created_at
		 FROM release_event WHERE release_id = $1 ORDER BY created_at ASC`, id,
	)
	if err != nil {
		return rel, nil, err
	}
	defer rows.Close()
	var events []domain.ReleaseEvent
	for rows.Next() {
		var e domain.ReleaseEvent
		if err := rows.Scan(&e.ID, &e.ReleaseID, &e.EventType, &e.StatusBefore, &e.StatusAfter, &e.Payload, &e.CreatedBy, &e.CreatedAt); err != nil {
			return rel, nil, err
		}
		events = append(events, e)
	}
	return rel, events, rows.Err()
}

func (r *releaseRepoPG) ListByService(ctx context.Context, serviceID int64, p domain.Pagination) ([]domain.Release, int, error) {
	p.Normalize()
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM "release" WHERE service_id = $1`, serviceID).Scan(&total)
	rows, err := r.db.Query(ctx, `SELECT `+releaseCols+` FROM "release" WHERE service_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3`, serviceID, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return r.scanRows(rows), total, nil
}

func (r *releaseRepoPG) List(ctx context.Context, p domain.Pagination) ([]domain.Release, int, error) {
	p.Normalize()
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM "release"`).Scan(&total)
	rows, err := r.db.Query(ctx, `SELECT `+releaseCols+` FROM "release" ORDER BY id DESC LIMIT $1 OFFSET $2`, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return r.scanRows(rows), total, nil
}

func (r *releaseRepoPG) scanRows(rows pgx.Rows) []domain.Release {
	var items []domain.Release
	for rows.Next() {
		rel := &domain.Release{}
		rows.Scan(&rel.ID, &rel.ServiceID, &rel.EnvironmentID, &rel.Version, &rel.CommitSHA, &rel.ImageTag, &rel.Strategy, &rel.Status, &rel.TriggeredBy, &rel.ApprovedBy, &rel.StartedAt, &rel.FinishedAt, &rel.Summary, &rel.RiskLevel)
		items = append(items, *rel)
	}
	return items
}

func (r *releaseRepoPG) ListByStatus(ctx context.Context, status domain.ReleaseStatus, p domain.Pagination) ([]domain.Release, int, error) {
	p.Normalize()
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM "release" WHERE status = $1`, status).Scan(&total)
	rows, err := r.db.Query(ctx, `SELECT `+releaseCols+` FROM "release" WHERE status = $1 ORDER BY id DESC LIMIT $2 OFFSET $3`, status, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return r.scanRows(rows), total, nil
}

func (r *releaseRepoPG) UpdateStatus(ctx context.Context, id int64, status domain.ReleaseStatus) error {
	tag, err := r.db.Exec(ctx, `UPDATE "release" SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("release %d not found", id)
	}
	return nil
}
