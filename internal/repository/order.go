package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Alfafauzan/coffee-pos/internal/entity"
)

// OrderFilter defines filtering and pagination parameters for listing orders.
type OrderFilter struct {
	ShiftID   string
	CashierID string
	Status    string
	Page      int
	Limit     int
}

// OrderRepository specifies data access operations for customer orders and items.
type OrderRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Order, error)
	FindAll(ctx context.Context, filter OrderFilter) ([]entity.Order, int, error)
	Create(ctx context.Context, order *entity.Order) error
	Update(ctx context.Context, order *entity.Order) error
	AddItem(ctx context.Context, item *entity.OrderItem) error
	UpdateItem(ctx context.Context, item *entity.OrderItem) error
	DeleteItem(ctx context.Context, itemID string) error
	FindItemByID(ctx context.Context, itemID string) (*entity.OrderItem, error)
	WithTx(tx *sql.Tx) OrderRepository
}

type orderRepository struct {
	db sqlDB
}

// NewOrderRepository creates a new OrderRepository implementation.
func NewOrderRepository(db sqlDB) OrderRepository {
	return &orderRepository{
		db: db,
	}
}

// WithTx returns a new OrderRepository instance operating within the provided transaction.
func (r *orderRepository) WithTx(tx *sql.Tx) OrderRepository {
	return &orderRepository{db: tx}
}

// FindByID retrieves an order by its ID with joined cashier and table,
// followed by a separate query to fetch all order items with product details.
func (r *orderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
	orderQuery := `
		SELECT o.id, o.shift_id, o.cashier_id, o.table_id, o.promo_id,
		       o.status, o.subtotal, o.discount_amount, o.total, o.notes,
		       o.created_at, o.updated_at,
		       u.id, u.name,
		       t.id, t.name,
		       pr.id, pr.name, pr.code
		FROM orders o
		JOIN users u ON u.id = o.cashier_id
		LEFT JOIN tables t ON t.id = o.table_id
		LEFT JOIN promos pr ON pr.id = o.promo_id
		WHERE o.id = ?
	`

	var (
		order       entity.Order
		tableID     sql.NullString
		promoID     sql.NullString
		notes       sql.NullString
		cashierID   string
		cashierName string
		tID         sql.NullString
		tName       sql.NullString
		prID        sql.NullString
		prName      sql.NullString
		prCode      sql.NullString
	)

	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&order.ID,
		&order.ShiftID,
		&order.CashierID,
		&tableID,
		&promoID,
		&order.Status,
		&order.Subtotal,
		&order.DiscountAmount,
		&order.Total,
		&notes,
		&order.CreatedAt,
		&order.UpdatedAt,
		&cashierID,
		&cashierName,
		&tID,
		&tName,
		&prID,
		&prName,
		&prCode,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if tableID.Valid {
		order.TableID = &tableID.String
		if tID.Valid {
			order.Table = &entity.Table{
				ID:   tID.String,
				Name: tName.String,
			}
		}
	}

	if promoID.Valid {
		order.PromoID = &promoID.String
		if prID.Valid {
			order.Promo = &entity.Promo{
				ID:   prID.String,
				Name: prName.String,
				Code: prCode.String,
			}
		}
	}

	if notes.Valid {
		order.Notes = notes.String
	}

	order.Cashier = &entity.User{
		ID:   cashierID,
		Name: cashierName,
	}

	// Fetch items associated with this order
	itemsQuery := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.subtotal,
		       oi.notes, oi.created_at, oi.updated_at,
		       p.id, p.name, p.price as product_price
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = ?
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, order.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.OrderItem, 0)
	for rows.Next() {
		var (
			item         entity.OrderItem
			itemNotes    sql.NullString
			prodID       string
			prodName     string
			productPrice int64
		)

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
			&item.Subtotal,
			&itemNotes,
			&item.CreatedAt,
			&item.UpdatedAt,
			&prodID,
			&prodName,
			&productPrice,
		)
		if err != nil {
			return nil, err
		}

		if itemNotes.Valid {
			item.Notes = itemNotes.String
		}

		item.Product = &entity.Product{
			ID:    prodID,
			Name:  prodName,
			Price: productPrice,
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	order.Items = items

	return &order, nil
}

// FindAll retrieves paginated orders without items for list view,
// filtering dynamically by ShiftID, CashierID, and Status.
func (r *orderRepository) FindAll(ctx context.Context, filter OrderFilter) ([]entity.Order, int, error) {
	var conditions []string
	var args []interface{}

	if filter.ShiftID != "" {
		conditions = append(conditions, "o.shift_id = ?")
		args = append(args, filter.ShiftID)
	}

	if filter.CashierID != "" {
		conditions = append(conditions, "o.cashier_id = ?")
		args = append(args, filter.CashierID)
	}

	if filter.Status != "" {
		conditions = append(conditions, "o.status = ?")
		args = append(args, filter.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. COUNT query for total matching orders
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM orders o
		%s
	`, whereClause)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return make([]entity.Order, 0), 0, nil
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

	// 3. Data query with cashier JOIN
	dataQuery := fmt.Sprintf(`
		SELECT o.id, o.shift_id, o.cashier_id, o.table_id, o.promo_id,
		       o.status, o.subtotal, o.discount_amount, o.total, o.notes,
		       o.created_at, o.updated_at,
		       u.id, u.name
		FROM orders o
		JOIN users u ON u.id = o.cashier_id
		%s
		ORDER BY o.created_at DESC
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

	orders := make([]entity.Order, 0)
	for rows.Next() {
		var (
			order       entity.Order
			tableID     sql.NullString
			promoID     sql.NullString
			notes       sql.NullString
			cashierID   string
			cashierName string
		)

		err := rows.Scan(
			&order.ID,
			&order.ShiftID,
			&order.CashierID,
			&tableID,
			&promoID,
			&order.Status,
			&order.Subtotal,
			&order.DiscountAmount,
			&order.Total,
			&notes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&cashierID,
			&cashierName,
		)
		if err != nil {
			return nil, 0, err
		}

		if tableID.Valid {
			order.TableID = &tableID.String
		}

		if promoID.Valid {
			order.PromoID = &promoID.String
		}

		if notes.Valid {
			order.Notes = notes.String
		}

		order.Cashier = &entity.User{
			ID:   cashierID,
			Name: cashierName,
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// Create inserts a new order record with generated UUID and current timestamps.
// Nullable pointer fields (table_id, promo_id) are automatically mapped to NULL when nil.
func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	order.ID = uuid.New().String()
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	if order.Status == "" {
		order.Status = entity.OrderStatusDraft
	}

	query := `
		INSERT INTO orders (
			id, shift_id, cashier_id, table_id, promo_id,
			status, subtotal, discount_amount, total, notes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.ID,
		order.ShiftID,
		order.CashierID,
		order.TableID,
		order.PromoID,
		order.Status,
		order.Subtotal,
		order.DiscountAmount,
		order.Total,
		order.Notes,
		order.CreatedAt,
		order.UpdatedAt,
	)
	return err
}

// Update updates order state, table, promo, pricing totals, and notes.
func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	order.UpdatedAt = time.Now()

	query := `
		UPDATE orders
		SET table_id = ?, promo_id = ?, status = ?, subtotal = ?,
		    discount_amount = ?, total = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.TableID,
		order.PromoID,
		order.Status,
		order.Subtotal,
		order.DiscountAmount,
		order.Total,
		order.Notes,
		order.UpdatedAt,
		order.ID,
	)
	return err
}

// AddItem inserts a new order line item, calculating its subtotal in the repository.
func (r *orderRepository) AddItem(ctx context.Context, item *entity.OrderItem) error {
	item.ID = uuid.New().String()
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now
	item.Subtotal = item.Price * int64(item.Quantity)

	query := `
		INSERT INTO order_items (
			id, order_id, product_id, quantity, price, subtotal,
			notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.OrderID,
		item.ProductID,
		item.Quantity,
		item.Price,
		item.Subtotal,
		item.Notes,
		item.CreatedAt,
		item.UpdatedAt,
	)
	return err
}

// UpdateItem updates quantity, notes, subtotal, and updated_at for an existing line item.
func (r *orderRepository) UpdateItem(ctx context.Context, item *entity.OrderItem) error {
	item.UpdatedAt = time.Now()

	query := `
		UPDATE order_items
		SET quantity = ?, notes = ?, subtotal = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.Quantity,
		item.Notes,
		item.Subtotal,
		item.UpdatedAt,
		item.ID,
	)
	return err
}

// DeleteItem performs a hard delete on an order item by its ID.
func (r *orderRepository) DeleteItem(ctx context.Context, itemID string) error {
	query := `DELETE FROM order_items WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, itemID)
	return err
}

// FindItemByID retrieves a single order item by its ID without joining product details.
// Returns nil, nil if the item is not found.
func (r *orderRepository) FindItemByID(ctx context.Context, itemID string) (*entity.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, price, subtotal,
		       notes, created_at, updated_at
		FROM order_items
		WHERE id = ?
	`

	var (
		item  entity.OrderItem
		notes sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID,
		&item.OrderID,
		&item.ProductID,
		&item.Quantity,
		&item.Price,
		&item.Subtotal,
		&notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if notes.Valid {
		item.Notes = notes.String
	}

	return &item, nil
}
