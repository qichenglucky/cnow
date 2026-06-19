package repo

import (
	"context"
	"time"

	"cnow/backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditRepository defines audit log persistence operations.
type AuditRepository interface {
	Create(ctx context.Context, a *domain.AuditLog) error
	ListByResource(ctx context.Context, resourceType string, resourceID int64, p domain.Pagination) ([]domain.AuditLog, int, error)
	ListByActor(ctx context.Context, actorID int64, p domain.Pagination) ([]domain.AuditLog, int, error)
	ListByTimeRange(ctx context.Context, from, to time.Time, p domain.Pagination) ([]domain.AuditLog, int, error)
}

type auditRepoPG struct {
	db *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) AuditRepository {
	return &auditRepoPG{db: db}
}

func (r *auditRepoPG) Create(ctx context.Context, a *domain.AuditLog) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO audit_log (actor_id, actor_role, action, resource_type, resource_id, request_id, detail, result, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		a.ActorID, a.ActorRole, a.Action, a.ResourceType, a.ResourceID,
		a.RequestID, a.Detail, a.Result, time.Now().UTC(),
	).Scan(&a.ID)
}

func (r *auditRepoPG) ListByResource(ctx context.Context, resourceType string, resourceID int64, p domain.Pagination) ([]domain.AuditLog, int, error) {
	p.Normalize()
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_log WHERE resource_type=$1 AND resource_id=$2`, resourceType, resourceID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, actor_id, actor_role, action, resource_type, resource_id, request_id, detail, result, created_at
		 FROM audit_log WHERE resource_type=$1 AND resource_id=$2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		resourceType, resourceID, p.Limit, p.Offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return r.scanRows(rows), total, nil
}

func (r *auditRepoPG) ListByActor(ctx context.Context, actorID int64, p domain.Pagination) ([]domain.AuditLog, int, error) {
	p.Normalize()
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_log WHERE actor_id=$1`, actorID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, actor_id, actor_role, action, resource_type, resource_id, request_id, detail, result, created_at
		 FROM audit_log WHERE actor_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		actorID, p.Limit, p.Offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return r.scanRows(rows), total, nil
}

func (r *auditRepoPG) ListByTimeRange(ctx context.Context, from, to time.Time, p domain.Pagination) ([]domain.AuditLog, int, error) {
	p.Normalize()
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_log WHERE created_at BETWEEN $1 AND $2`, from, to).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, actor_id, actor_role, action, resource_type, resource_id, request_id, detail, result, created_at
		 FROM audit_log WHERE created_at BETWEEN $1 AND $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		from, to, p.Limit, p.Offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return r.scanRows(rows), total, nil
}

func (r *auditRepoPG) scanRows(rows interface{ Next() bool; Scan(...interface{}) error }) []domain.AuditLog {
	var items []domain.AuditLog
	for rows.Next() {
		var a domain.AuditLog
		rows.Scan(&a.ID, &a.ActorID, &a.ActorRole, &a.Action, &a.ResourceType, &a.ResourceID, &a.RequestID, &a.Detail, &a.Result, &a.CreatedAt)
		items = append(items, a)
	}
	return items
}
