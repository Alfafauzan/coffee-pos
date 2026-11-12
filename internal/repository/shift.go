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

// ShiftRepository specifies data access operations for cashier shifts.
type ShiftRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Shift, error)
	FindOpenByCashierID(ctx context.Context, cashierID string) (*entity.Shift, error)
	FindAll(ctx context.Context, cashierID string, page, limit int) ([]entity.Shift, int, error)
	Create(ctx context.Context, shift *entity.Shift) error
	Close(ctx context.Context, id string, closingCash int64, notes string) error
	WithTx(tx *sql.Tx) ShiftRepository
}

type shiftRepository struct {
	db sqlDB
}

// NewShiftRepository creates a new ShiftRepository implementation.
func NewShiftRepository(db sqlDB) ShiftRepository {
	return &shiftRepository{
		db: db,
	}
}

// WithTx returns a new ShiftRepository instance operating within the provided transaction.
func (r *shiftRepository) WithTx(tx *sql.Tx) ShiftRepository {
	return &shiftRepository{db: tx}
}

// FindByID retrieves a shift by its ID with joined cashier user details.
// Returns nil, nil if not found.
func (r *shiftRepository) FindByID(ctx context.Context, id string) (*entity.Shift, error) {
	query := `
		SELECT s.id, s.cashier_id, s.opened_at, s.closed_at, s.opening_cash,
		       s.closing_cash, s.total_transactions, s.status, s.notes,
		       s.created_at, s.updated_at,
		       u.id, u.name
		FROM shifts s
		JOIN users u ON u.id = s.cashier_id
		WHERE s.id = ?
	`

	var (
		shift       entity.Shift
		closedAt    sql.NullTime
		closingCash sql.NullInt64
		notes       sql.NullString
		cashierID   string
		cashierName string
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&shift.ID,
		&shift.CashierID,
		&shift.OpenedAt,
		&closedAt,
		&shift.OpeningCash,
		&closingCash,
		&shift.TotalTransactions,
		&shift.Status,
		&notes,
		&shift.CreatedAt,
		&shift.UpdatedAt,
		&cashierID,
		&cashierName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if closedAt.Valid {
		shift.ClosedAt = &closedAt.Time
	}
	if closingCash.Valid {
		shift.ClosingCash = &closingCash.Int64
	}
	if notes.Valid {
		shift.Notes = notes.String
	}

	shift.Cashier = &entity.User{
		ID:   cashierID,
		Name: cashierName,
	}

	return &shift, nil
}

// FindOpenByCashierID checks if a cashier currently has an open shift.
// Returns nil, nil if there is no active open shift.
func (r *shiftRepository) FindOpenByCashierID(ctx context.Context, cashierID string) (*entity.Shift, error) {
	query := `
		SELECT s.id, s.cashier_id, s.opened_at, s.closed_at, s.opening_cash,
		       s.closing_cash, s.total_transactions, s.status, s.notes,
		       s.created_at, s.updated_at,
		       u.id, u.name
		FROM shifts s
		JOIN users u ON u.id = s.cashier_id
		WHERE s.cashier_id = ? AND s.status = 'open' AND u.deleted_at IS NULL
		ORDER BY s.opened_at DESC
		LIMIT 1
	`

	var (
		shift       entity.Shift
		closedAt    sql.NullTime
		closingCash sql.NullInt64
		notes       sql.NullString
		uID         string
		uName       string
	)

	err := r.db.QueryRowContext(ctx, query, cashierID).Scan(
		&shift.ID,
		&shift.CashierID,
		&shift.OpenedAt,
		&closedAt,
		&shift.OpeningCash,
		&closingCash,
		&shift.TotalTransactions,
		&shift.Status,
		&notes,
		&shift.CreatedAt,
		&shift.UpdatedAt,
		&uID,
		&uName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if closedAt.Valid {
		shift.ClosedAt = &closedAt.Time
	}
	if closingCash.Valid {
		shift.ClosingCash = &closingCash.Int64
	}
	if notes.Valid {
		shift.Notes = notes.String
	}

	shift.Cashier = &entity.User{
		ID:   uID,
		Name: uName,
	}

	return &shift, nil
}

// FindAll retrieves paginated shifts optionally filtered by cashierID,
// returning shifts with cashier details, total count, and error.
func (r *shiftRepository) FindAll(ctx context.Context, cashierID string, page, limit int) ([]entity.Shift, int, error) {
	var conditions []string
	var args []interface{}

	if cashierID != "" {
		conditions = append(conditions, "s.cashier_id = ?")
		args = append(args, cashierID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. COUNT query for total matching shifts
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM shifts s
		%s
	`, whereClause)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return make([]entity.Shift, 0), 0, nil
	}

	// 2. Pagination calculation
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	// 3. Data query with cashier JOIN
	dataQuery := fmt.Sprintf(`
		SELECT s.id, s.cashier_id, s.opened_at, s.closed_at, s.opening_cash,
		       s.closing_cash, s.total_transactions, s.status, s.notes,
		       s.created_at, s.updated_at,
		       u.id, u.name
		FROM shifts s
		JOIN users u ON u.id = s.cashier_id
		%s
		ORDER BY s.opened_at DESC
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

	shifts := make([]entity.Shift, 0)
	for rows.Next() {
		var (
			shift       entity.Shift
			closedAt    sql.NullTime
			closingCash sql.NullInt64
			notes       sql.NullString
			uID         string
			uName       string
		)

		err := rows.Scan(
			&shift.ID,
			&shift.CashierID,
			&shift.OpenedAt,
			&closedAt,
			&shift.OpeningCash,
			&closingCash,
			&shift.TotalTransactions,
			&shift.Status,
			&notes,
			&shift.CreatedAt,
			&shift.UpdatedAt,
			&uID,
			&uName,
		)
		if err != nil {
			return nil, 0, err
		}

		if closedAt.Valid {
			shift.ClosedAt = &closedAt.Time
		}
		if closingCash.Valid {
			shift.ClosingCash = &closingCash.Int64
		}
		if notes.Valid {
			shift.Notes = notes.String
		}

		shift.Cashier = &entity.User{
			ID:   uID,
			Name: uName,
		}

		shifts = append(shifts, shift)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return shifts, total, nil
}

// Create inserts a new shift record with status 'open', generated UUID, and timestamps.
func (r *shiftRepository) Create(ctx context.Context, shift *entity.Shift) error {
	shift.ID = uuid.New().String()
	now := time.Now()
	shift.CreatedAt = now
	shift.UpdatedAt = now
	shift.OpenedAt = now

	if shift.Status == "" {
		shift.Status = entity.ShiftStatusOpen
	}

	query := `
		INSERT INTO shifts (
			id, cashier_id, opened_at, opening_cash,
			total_transactions, status, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		shift.ID,
		shift.CashierID,
		shift.OpenedAt,
		shift.OpeningCash,
		shift.TotalTransactions,
		shift.Status,
		shift.Notes,
		shift.CreatedAt,
		shift.UpdatedAt,
	)
	return err
}

// Close marks an open shift as 'closed', setting closing_cash, closed_at, notes, and updated_at.
func (r *shiftRepository) Close(ctx context.Context, id string, closingCash int64, notes string) error {
	query := `
		UPDATE shifts
		SET status = 'closed',
		    closed_at = NOW(),
		    closing_cash = ?,
		    notes = ?,
		    updated_at = NOW()
		WHERE id = ? AND status = 'open'
	`

	res, err := r.db.ExecContext(ctx, query, closingCash, notes, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("shift tidak ditemukan atau sudah ditutup")
	}

	return nil
}
