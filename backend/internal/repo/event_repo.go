package repo

import (
	"context"
	"time"

	"cnow/backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository interface {
	Create(ctx context.Context, e *domain.ReleaseEvent) error
	ListByRelease(ctx context.Context, releaseID int64) ([]domain.ReleaseEvent, error)
}

type eventRepoPG struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) EventRepository {
	return &eventRepoPG{db: db}
}

func (r *eventRepoPG) Create(ctx context.Context, e *domain.ReleaseEvent) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO release_event (release_id, event_type, status_before, status_after, payload, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		e.ReleaseID, e.EventType, e.StatusBefore, e.StatusAfter, e.Payload, e.CreatedBy, time.Now().UTC(),
	).Scan(&e.ID)
}

func (r *eventRepoPG) ListByRelease(ctx context.Context, releaseID int64) ([]domain.ReleaseEvent, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, release_id, event_type, status_before, status_after, payload, created_by, created_at
		 FROM release_event WHERE release_id = $1 ORDER BY created_at ASC`, releaseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ReleaseEvent
	for rows.Next() {
		var e domain.ReleaseEvent
		if err := rows.Scan(&e.ID, &e.ReleaseID, &e.EventType, &e.StatusBefore, &e.StatusAfter, &e.Payload, &e.CreatedBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}
