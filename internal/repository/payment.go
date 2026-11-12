package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Alfafauzan/coffee-pos/internal/entity"
)

// PaymentRepository specifies data access operations for order payments.
type PaymentRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Payment, error)
	FindByOrderID(ctx context.Context, orderID string) (*entity.Payment, error)
	FindByMidtransOrderID(ctx context.Context, midtransOrderID string) (*entity.Payment, error)
	Create(ctx context.Context, payment *entity.Payment) error
	UpdateStatus(ctx context.Context, id string, status string, paidAt *time.Time) error
	UpdateMidtransData(ctx context.Context, id string, token string, url string, midtransOrderID string) error
	SaveRawNotification(ctx context.Context, id string, raw string) error
	WithTx(tx *sql.Tx) PaymentRepository
}

type paymentRepository struct {
	db sqlDB
}

// NewPaymentRepository creates a new PaymentRepository implementation.
func NewPaymentRepository(db sqlDB) PaymentRepository {
	return &paymentRepository{
		db: db,
	}
}

// WithTx returns a new PaymentRepository instance operating within the provided transaction.
func (r *paymentRepository) WithTx(tx *sql.Tx) PaymentRepository {
	return &paymentRepository{db: tx}
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanPaymentRow(row rowScanner) (*entity.Payment, error) {
	var (
		p               entity.Payment
		midtransOrderID sql.NullString
		midtransToken   sql.NullString
		midtransURL     sql.NullString
		rawNotification sql.NullString
		paidAt          sql.NullTime
	)

	err := row.Scan(
		&p.ID,
		&p.OrderID,
		&p.Method,
		&p.Status,
		&p.Amount,
		&midtransOrderID,
		&midtransToken,
		&midtransURL,
		&rawNotification,
		&paidAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if midtransOrderID.Valid {
		p.MidtransOrderID = &midtransOrderID.String
	}
	if midtransToken.Valid {
		p.MidtransToken = &midtransToken.String
	}
	if midtransURL.Valid {
		p.MidtransURL = &midtransURL.String
	}
	if rawNotification.Valid {
		p.RawNotification = &rawNotification.String
	}
	if paidAt.Valid {
		p.PaidAt = &paidAt.Time
	}

	return &p, nil
}

// FindByID retrieves a payment by its UUID. Returns nil, nil if not found.
func (r *paymentRepository) FindByID(ctx context.Context, id string) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, method, status, amount,
		       midtrans_order_id, midtrans_token, midtrans_url,
		       raw_notification, paid_at, created_at, updated_at
		FROM payments
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanPaymentRow(row)
}

// FindByOrderID retrieves a payment by its associated order ID. Returns nil, nil if not found.
func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID string) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, method, status, amount,
		       midtrans_order_id, midtrans_token, midtrans_url,
		       raw_notification, paid_at, created_at, updated_at
		FROM payments
		WHERE order_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, orderID)
	return scanPaymentRow(row)
}

// FindByMidtransOrderID retrieves a payment by its Midtrans order ID. Returns nil, nil if not found.
func (r *paymentRepository) FindByMidtransOrderID(ctx context.Context, midtransOrderID string) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, method, status, amount,
		       midtrans_order_id, midtrans_token, midtrans_url,
		       raw_notification, paid_at, created_at, updated_at
		FROM payments
		WHERE midtrans_order_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, midtransOrderID)
	return scanPaymentRow(row)
}

// Create inserts a new payment record with status 'pending', generated UUID, and timestamps.
// CATATAN: kolom order_id di tabel payments adalah UNIQUE — satu order
// hanya boleh punya satu payment record. Kalau Create dipanggil dua kali
// untuk order yang sama, MySQL akan throw duplicate key error. Service layer
// harus cek FindByOrderID dulu sebelum Create.
func (r *paymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	payment.ID = uuid.New().String()
	now := time.Now()
	payment.CreatedAt = now
	payment.UpdatedAt = now

	if payment.Status == "" {
		payment.Status = entity.PaymentStatusPending
	}

	query := `
		INSERT INTO payments (
			id, order_id, method, status, amount,
			midtrans_order_id, midtrans_token, midtrans_url,
			raw_notification, paid_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.Method,
		payment.Status,
		payment.Amount,
		payment.MidtransOrderID,
		payment.MidtransToken,
		payment.MidtransURL,
		payment.RawNotification,
		payment.PaidAt,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	return err
}

// UpdateStatus updates payment status, paid_at timestamp, and updated_at.
func (r *paymentRepository) UpdateStatus(ctx context.Context, id string, status string, paidAt *time.Time) error {
	query := `
		UPDATE payments
		SET status = ?, paid_at = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, status, paidAt, id)
	return err
}

// UpdateMidtransData updates Midtrans token, redirect URL, and Midtrans order ID.
func (r *paymentRepository) UpdateMidtransData(ctx context.Context, id string, token string, url string, midtransOrderID string) error {
	query := `
		UPDATE payments
		SET midtrans_token = ?, midtrans_url = ?, midtrans_order_id = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, token, url, midtransOrderID, id)
	return err
}

// SaveRawNotification stores the raw webhook payload JSON received from payment gateway.
func (r *paymentRepository) SaveRawNotification(ctx context.Context, id string, raw string) error {
	query := `
		UPDATE payments
		SET raw_notification = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, raw, id)
	return err
}
