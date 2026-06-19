package repo

import (
	"context"
	"fmt"

	"cnow/backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ServiceRepository defines service persistence operations.
type ServiceRepository interface {
	Create(ctx context.Context, s *domain.Service) error
	GetByID(ctx context.Context, id int64) (*domain.Service, error)
	GetByName(ctx context.Context, name string) (*domain.Service, error)
	List(ctx context.Context, p domain.Pagination) ([]domain.Service, int, error)
	Update(ctx context.Context, s *domain.Service) error
	UpdateStatus(ctx context.Context, id int64, status domain.ServiceStatus) error
}

type serviceRepoPG struct {
	db *pgxpool.Pool
}

func NewServiceRepository(db *pgxpool.Pool) ServiceRepository {
	return &serviceRepoPG{db: db}
}

func (r *serviceRepoPG) Create(ctx context.Context, s *domain.Service) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO service (name, display_name, description, tech_stack, status, owner_id)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		s.Name, s.DisplayName, s.Description, s.TechStack, s.Status, s.OwnerID,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *serviceRepoPG) GetByID(ctx context.Context, id int64) (*domain.Service, error) {
	s := &domain.Service{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, display_name, description, tech_stack, status, owner_id, created_at, updated_at
		 FROM service WHERE id = $1`, id,
	).Scan(&s.ID, &s.Name, &s.DisplayName, &s.Description, &s.TechStack, &s.Status, &s.OwnerID, &s.CreatedAt, &s.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *serviceRepoPG) GetByName(ctx context.Context, name string) (*domain.Service, error) {
	s := &domain.Service{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, display_name, description, tech_stack, status, owner_id, created_at, updated_at
		 FROM service WHERE name = $1`, name,
	).Scan(&s.ID, &s.Name, &s.DisplayName, &s.Description, &s.TechStack, &s.Status, &s.OwnerID, &s.CreatedAt, &s.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *serviceRepoPG) List(ctx context.Context, p domain.Pagination) ([]domain.Service, int, error) {
	p.Normalize()

	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM service`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count services: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, name, display_name, description, tech_stack, status, owner_id, created_at, updated_at
		 FROM service ORDER BY id DESC LIMIT $1 OFFSET $2`, p.Limit, p.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	var items []domain.Service
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.DisplayName, &s.Description, &s.TechStack, &s.Status, &s.OwnerID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, rows.Err()
}

func (r *serviceRepoPG) Update(ctx context.Context, s *domain.Service) error {
	_, err := r.db.Exec(ctx,
		`UPDATE service SET display_name=$1, description=$2, tech_stack=$3, updated_at=NOW() WHERE id=$4`,
		s.DisplayName, s.Description, s.TechStack, s.ID,
	)
	return err
}

func (r *serviceRepoPG) UpdateStatus(ctx context.Context, id int64, status domain.ServiceStatus) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE service SET status=$1, updated_at=NOW() WHERE id=$2`, status, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("service %d not found", id)
	}
	return nil
}
