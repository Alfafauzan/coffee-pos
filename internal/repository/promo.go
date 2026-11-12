package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Alfafauzan/coffee-pos/internal/entity"
)

// PromoRepository specifies data access operations for promotional vouchers.
type PromoRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Promo, error)
	FindByCode(ctx context.Context, code string) (*entity.Promo, error)
	FindAll(ctx context.Context, page, limit int) ([]entity.Promo, int, error)
	Create(ctx context.Context, promo *entity.Promo) error
	Update(ctx context.Context, promo *entity.Promo) error
	Delete(ctx context.Context, id string) error
	IncrementUsedCount(ctx context.Context, id string) error
	WithTx(tx *sql.Tx) PromoRepository
}

type promoRepository struct {
	db sqlDB
}

// NewPromoRepository creates a new PromoRepository implementation.
func NewPromoRepository(db sqlDB) PromoRepository {
	return &promoRepository{
		db: db,
	}
}

// WithTx returns a new PromoRepository instance operating within the provided transaction.
func (r *promoRepository) WithTx(tx *sql.Tx) PromoRepository {
	return &promoRepository{db: tx}
}

// FindByID retrieves an active promo by its UUID. Returns nil, nil if not found.
func (r *promoRepository) FindByID(ctx context.Context, id string) (*entity.Promo, error) {
	query := `
		SELECT id, name, code, type, value, min_order, max_discount, usage_limit,
		       used_count, started_at, ended_at, is_active, created_at, updated_at, deleted_at
		FROM promos
		WHERE id = ? AND deleted_at IS NULL
	`

	var (
		promo       entity.Promo
		maxDiscount sql.NullInt64
		usageLimit  sql.NullInt64
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&promo.ID,
		&promo.Name,
		&promo.Code,
		&promo.Type,
		&promo.Value,
		&promo.MinOrder,
		&maxDiscount,
		&usageLimit,
		&promo.UsedCount,
		&promo.StartedAt,
		&promo.EndedAt,
		&promo.IsActive,
		&promo.CreatedAt,
		&promo.UpdatedAt,
		&promo.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if maxDiscount.Valid {
		promo.MaxDiscount = &maxDiscount.Int64
	}
	if usageLimit.Valid {
		limit := int(usageLimit.Int64)
		promo.UsageLimit = &limit
	}

	return &promo, nil
}

// FindByCode retrieves an active promo by its unique code string. Returns nil, nil if not found.
func (r *promoRepository) FindByCode(ctx context.Context, code string) (*entity.Promo, error) {
	query := `
		SELECT id, name, code, type, value, min_order, max_discount, usage_limit,
		       used_count, started_at, ended_at, is_active, created_at, updated_at, deleted_at
		FROM promos
		WHERE code = ? AND deleted_at IS NULL
	`

	var (
		promo       entity.Promo
		maxDiscount sql.NullInt64
		usageLimit  sql.NullInt64
	)

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&promo.ID,
		&promo.Name,
		&promo.Code,
		&promo.Type,
		&promo.Value,
		&promo.MinOrder,
		&maxDiscount,
		&usageLimit,
		&promo.UsedCount,
		&promo.StartedAt,
		&promo.EndedAt,
		&promo.IsActive,
		&promo.CreatedAt,
		&promo.UpdatedAt,
		&promo.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if maxDiscount.Valid {
		promo.MaxDiscount = &maxDiscount.Int64
	}
	if usageLimit.Valid {
		limit := int(usageLimit.Int64)
		promo.UsageLimit = &limit
	}

	return &promo, nil
}

// FindAll retrieves paginated active promos ordered by created_at DESC.
func (r *promoRepository) FindAll(ctx context.Context, page, limit int) ([]entity.Promo, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM promos
		WHERE deleted_at IS NULL
	`

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return make([]entity.Promo, 0), 0, nil
	}

	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	dataQuery := `
		SELECT id, name, code, type, value, min_order, max_discount, usage_limit,
		       used_count, started_at, ended_at, is_active, created_at, updated_at, deleted_at
		FROM promos
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, dataQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	promos := make([]entity.Promo, 0)
	for rows.Next() {
		var (
			promo       entity.Promo
			maxDiscount sql.NullInt64
			usageLimit  sql.NullInt64
		)

		err := rows.Scan(
			&promo.ID,
			&promo.Name,
			&promo.Code,
			&promo.Type,
			&promo.Value,
			&promo.MinOrder,
			&maxDiscount,
			&usageLimit,
			&promo.UsedCount,
			&promo.StartedAt,
			&promo.EndedAt,
			&promo.IsActive,
			&promo.CreatedAt,
			&promo.UpdatedAt,
			&promo.DeletedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if maxDiscount.Valid {
			promo.MaxDiscount = &maxDiscount.Int64
		}
		if usageLimit.Valid {
			limit := int(usageLimit.Int64)
			promo.UsageLimit = &limit
		}

		promos = append(promos, promo)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return promos, total, nil
}

// Create inserts a new promo record with generated UUID and timestamps.
func (r *promoRepository) Create(ctx context.Context, promo *entity.Promo) error {
	promo.ID = uuid.New().String()
	now := time.Now()
	promo.CreatedAt = now
	promo.UpdatedAt = now

	query := `
		INSERT INTO promos (
			id, name, code, type, value, min_order, max_discount, usage_limit,
			used_count, started_at, ended_at, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		promo.ID,
		promo.Name,
		promo.Code,
		promo.Type,
		promo.Value,
		promo.MinOrder,
		promo.MaxDiscount,
		promo.UsageLimit,
		promo.UsedCount,
		promo.StartedAt,
		promo.EndedAt,
		promo.IsActive,
		promo.CreatedAt,
		promo.UpdatedAt,
	)
	return err
}

// Update updates promo properties including name, type, value, limits, active dates, and status.
func (r *promoRepository) Update(ctx context.Context, promo *entity.Promo) error {
	promo.UpdatedAt = time.Now()

	query := `
		UPDATE promos
		SET name = ?, type = ?, value = ?, min_order = ?, max_discount = ?,
		    usage_limit = ?, started_at = ?, ended_at = ?, is_active = ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		promo.Name,
		promo.Type,
		promo.Value,
		promo.MinOrder,
		promo.MaxDiscount,
		promo.UsageLimit,
		promo.StartedAt,
		promo.EndedAt,
		promo.IsActive,
		promo.UpdatedAt,
		promo.ID,
	)
	return err
}

// Delete marks a promo as soft-deleted by setting deleted_at to NOW().
func (r *promoRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE promos
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// IncrementUsedCount increments the used_count counter for a promo by 1.
func (r *promoRepository) IncrementUsedCount(ctx context.Context, id string) error {
	query := `
		UPDATE promos
		SET used_count = used_count + 1, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
