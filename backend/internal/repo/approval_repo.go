package repo

import (
	"context"
	"fmt"

	"cnow/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApprovalRepository interface {
	Create(ctx context.Context, a *domain.Approval) error
	GetByID(ctx context.Context, id int64) (*domain.Approval, error)
	ListByRelease(ctx context.Context, releaseID int64) ([]domain.Approval, error)
	UpdateStatus(ctx context.Context, id int64, status domain.ApprovalStatus, comment string) error
}

type approvalRepoPG struct {
	db *pgxpool.Pool
}

func NewApprovalRepository(db *pgxpool.Pool) ApprovalRepository {
	return &approvalRepoPG{db: db}
}

func (r *approvalRepoPG) Create(ctx context.Context, a *domain.Approval) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO approval (release_id, approver_id, status, reason)
		 VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		a.ReleaseID, a.ApproverID, a.Status, a.Comment,
	).Scan(&a.ID, &a.CreatedAt)
}

func (r *approvalRepoPG) GetByID(ctx context.Context, id int64) (*domain.Approval, error) {
	a := &domain.Approval{}
	err := r.db.QueryRow(ctx,
		`SELECT id, release_id, approver_id, status, reason, created_at FROM approval WHERE id = $1`, id,
	).Scan(&a.ID, &a.ReleaseID, &a.ApproverID, &a.Status, &a.Comment, &a.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return a, err
}

func (r *approvalRepoPG) ListByRelease(ctx context.Context, releaseID int64) ([]domain.Approval, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, release_id, approver_id, status, reason, created_at
		 FROM approval WHERE release_id = $1 ORDER BY created_at DESC`, releaseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.Approval
	for rows.Next() {
		var a domain.Approval
		rows.Scan(&a.ID, &a.ReleaseID, &a.ApproverID, &a.Status, &a.Comment, &a.CreatedAt)
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *approvalRepoPG) UpdateStatus(ctx context.Context, id int64, status domain.ApprovalStatus, comment string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE approval SET status=$1, reason=$2 WHERE id=$3`, status, comment, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("approval %d not found", id)
	}
	return nil
}
