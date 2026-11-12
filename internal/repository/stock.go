package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Alfafauzan/coffee-pos/internal/entity"
)

// StockFilter defines filtering and pagination parameters for stock movements.
type StockFilter struct {
	ProductID string
	UserID    string
	Type      string
	Page      int
	Limit     int
}

// StockRepository specifies data access operations for stock movements.
type StockRepository interface {
	Create(ctx context.Context, movement *entity.StockMovement) error
	FindByProductID(ctx context.Context, productID string, filter StockFilter) ([]entity.StockMovement, int, error)
	FindAll(ctx context.Context, filter StockFilter) ([]entity.StockMovement, int, error)
	WithTx(tx *sql.Tx) StockRepository
}

type stockRepository struct {
	db sqlDB
}

// NewStockRepository creates a new StockRepository implementation.
func NewStockRepository(db sqlDB) StockRepository {
	return &stockRepository{
		db: db,
	}
}

// WithTx returns a new StockRepository instance operating within the provided transaction.
func (r *stockRepository) WithTx(tx *sql.Tx) StockRepository {
	return &stockRepository{db: tx}
}

// Create inserts a new stock movement record with generated UUID and current timestamps.
func (r *stockRepository) Create(ctx context.Context, movement *entity.StockMovement) error {
	movement.ID = uuid.New().String()
	now := time.Now()
	movement.CreatedAt = now
	movement.UpdatedAt = now

	query := `
		INSERT INTO stock_movements (
			id, product_id, user_id, type, quantity,
			stock_before, stock_after, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		movement.ID,
		movement.ProductID,
		movement.UserID,
		movement.Type,
		movement.Quantity,
		movement.StockBefore,
		movement.StockAfter,
		movement.Notes,
		movement.CreatedAt,
		movement.UpdatedAt,
	)
	return err
}

// FindByProductID retrieves paginated stock movements for a specific product ID.
func (r *stockRepository) FindByProductID(ctx context.Context, productID string, filter StockFilter) ([]entity.StockMovement, int, error) {
	filter.ProductID = productID
	return r.FindAll(ctx, filter)
}

// FindAll retrieves paginated stock movements matching dynamic filter criteria,
// returning movements with joined product and user details, total count, and error.
func (r *stockRepository) FindAll(ctx context.Context, filter StockFilter) ([]entity.StockMovement, int, error) {
	var conditions []string
	var args []interface{}

	if filter.ProductID != "" {
		conditions = append(conditions, "sm.product_id = ?")
		args = append(args, filter.ProductID)
	}

	if filter.UserID != "" {
		conditions = append(conditions, "sm.user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.Type != "" {
		conditions = append(conditions, "sm.type = ?")
		args = append(args, filter.Type)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. COUNT query for total matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM stock_movements sm
		%s
	`, whereClause)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return make([]entity.StockMovement, 0), 0, nil
	}

	// 2. Pagination calculation
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	// 3. Data query with JOIN and LIMIT OFFSET
	dataQuery := fmt.Sprintf(`
		SELECT sm.id, sm.product_id, sm.user_id, sm.type, sm.quantity,
		       sm.stock_before, sm.stock_after, sm.notes, sm.created_at, sm.updated_at,
		       p.id, p.name,
		       u.id, u.name
		FROM stock_movements sm
		JOIN products p ON p.id = sm.product_id
		JOIN users u ON u.id = sm.user_id
		%s
		ORDER BY sm.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	dataArgs := make([]interface{}, len(args), len(args)+2)
	copy(dataArgs, args)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	movements := make([]entity.StockMovement, 0)
	for rows.Next() {
		var (
			movement entity.StockMovement
			notes    sql.NullString
			product  entity.Product
			user     entity.User
		)

		err := rows.Scan(
			&movement.ID,
			&movement.ProductID,
			&movement.UserID,
			&movement.Type,
			&movement.Quantity,
			&movement.StockBefore,
			&movement.StockAfter,
			&notes,
			&movement.CreatedAt,
			&movement.UpdatedAt,
			&product.ID,
			&product.Name,
			&user.ID,
			&user.Name,
		)
		if err != nil {
			return nil, 0, err
		}

		if notes.Valid {
			movement.Notes = notes.String
		}

		movement.Product = &product
		movement.User = &user

		movements = append(movements, movement)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return movements, total, nil
}
