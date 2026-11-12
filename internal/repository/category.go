package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Alfafauzan/coffee-pos/internal/entity"
)

// CategoryRepository specifies data access operations for Category entities.
type CategoryRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Category, error)
	FindByName(ctx context.Context, name string) (*entity.Category, error)
	FindAll(ctx context.Context) ([]entity.Category, error)
	Create(ctx context.Context, category *entity.Category) error
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *sql.Tx) CategoryRepository
}

type categoryRepository struct {
	db sqlDB
}

// NewCategoryRepository creates a new CategoryRepository implementation.
func NewCategoryRepository(db sqlDB) CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

// WithTx returns a new CategoryRepository instance operating within the provided transaction.
func (r *categoryRepository) WithTx(tx *sql.Tx) CategoryRepository {
	return &categoryRepository{db: tx}
}

// FindByID retrieves an active category by its UUID. Returns nil, nil if not found.
func (r *categoryRepository) FindByID(ctx context.Context, id string) (*entity.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at, deleted_at
		FROM categories
		WHERE id = ? AND deleted_at IS NULL
	`

	var category entity.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

// FindByName retrieves an active category by its exact name. Returns nil, nil if not found.
func (r *categoryRepository) FindByName(ctx context.Context, name string) (*entity.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at, deleted_at
		FROM categories
		WHERE name = ? AND deleted_at IS NULL
	`

	var category entity.Category
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

// FindAll retrieves all active categories ordered by name ASC. Returns empty slice if none found.
func (r *categoryRepository) FindAll(ctx context.Context) ([]entity.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at, deleted_at
		FROM categories
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]entity.Category, 0)
	for rows.Next() {
		var category entity.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
			&category.DeletedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// Create inserts a new category record with generated UUID and current timestamps.
func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	category.ID = uuid.New().String()
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	query := `
		INSERT INTO categories (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query, category.ID, category.Name, category.CreatedAt, category.UpdatedAt)
	return err
}

// Update updates an existing active category's name and updated_at timestamp.
func (r *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
	category.UpdatedAt = time.Now()

	query := `
		UPDATE categories
		SET name = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, category.Name, category.UpdatedAt, category.ID)
	return err
}

// Delete marks a category as soft-deleted by setting deleted_at to NOW().
func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE categories
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
