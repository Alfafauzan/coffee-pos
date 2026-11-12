package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Alfafauzan/coffee-pos/internal/entity"
)

// TableRepository specifies data access operations for Table entities.
type TableRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Table, error)
	FindAll(ctx context.Context) ([]entity.Table, error)
	Create(ctx context.Context, table *entity.Table) error
	Update(ctx context.Context, table *entity.Table) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status string) error
	WithTx(tx *sql.Tx) TableRepository
}

type tableRepository struct {
	db sqlDB
}

// NewTableRepository creates a new TableRepository implementation.
func NewTableRepository(db sqlDB) TableRepository {
	return &tableRepository{
		db: db,
	}
}

// WithTx returns a new TableRepository instance operating within the provided transaction.
func (r *tableRepository) WithTx(tx *sql.Tx) TableRepository {
	return &tableRepository{db: tx}
}

// FindByID retrieves an active table by its UUID. Returns nil, nil if not found.
func (r *tableRepository) FindByID(ctx context.Context, id string) (*entity.Table, error) {
	query := `
		SELECT id, name, capacity, status, created_at, updated_at, deleted_at
		FROM tables
		WHERE id = ? AND deleted_at IS NULL
	`

	var table entity.Table
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&table.ID,
		&table.Name,
		&table.Capacity,
		&table.Status,
		&table.CreatedAt,
		&table.UpdatedAt,
		&table.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &table, nil
}

// FindAll retrieves all active tables ordered by name ASC. Returns empty slice if none found.
func (r *tableRepository) FindAll(ctx context.Context) ([]entity.Table, error) {
	query := `
		SELECT id, name, capacity, status, created_at, updated_at, deleted_at
		FROM tables
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]entity.Table, 0)
	for rows.Next() {
		var table entity.Table
		if err := rows.Scan(
			&table.ID,
			&table.Name,
			&table.Capacity,
			&table.Status,
			&table.CreatedAt,
			&table.UpdatedAt,
			&table.DeletedAt,
		); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

// Create inserts a new table with generated UUID and current timestamps.
func (r *tableRepository) Create(ctx context.Context, table *entity.Table) error {
	table.ID = uuid.New().String()
	now := time.Now()
	table.CreatedAt = now
	table.UpdatedAt = now

	if table.Status == "" {
		table.Status = entity.TableStatusAvailable
	}

	query := `
		INSERT INTO tables (id, name, capacity, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		table.ID,
		table.Name,
		table.Capacity,
		table.Status,
		table.CreatedAt,
		table.UpdatedAt,
	)
	return err
}

// Update updates name, capacity, and updated_at for an active table.
func (r *tableRepository) Update(ctx context.Context, table *entity.Table) error {
	table.UpdatedAt = time.Now()

	query := `
		UPDATE tables
		SET name = ?, capacity = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		table.Name,
		table.Capacity,
		table.UpdatedAt,
		table.ID,
	)
	return err
}

// Delete marks a table as soft-deleted by setting deleted_at to NOW().
func (r *tableRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE tables
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// UpdateStatus updates the occupancy status and updated_at timestamp of a table.
func (r *tableRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	now := time.Now()
	query := `
		UPDATE tables
		SET status = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, status, now, id)
	return err
}
