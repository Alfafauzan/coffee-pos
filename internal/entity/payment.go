package entity

import "time"

// Payment method constants
const (
	PaymentMethodCash     = "cash"
	PaymentMethodMidtrans = "midtrans"
)

// Payment status constants
const (
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
	PaymentStatusFailed  = "failed"
	PaymentStatusExpired = "expired"
)

// Payment represents a payment record for an order in the database.
type Payment struct {
	ID              string     `json:"id"`
	OrderID         string     `json:"order_id"`
	Method          string     `json:"method"`
	Status          string     `json:"status"`
	Amount          int64      `json:"amount"`
	MidtransOrderID *string    `json:"midtrans_order_id,omitempty"`
	MidtransToken   *string    `json:"midtrans_token,omitempty"`
	MidtransURL     *string    `json:"midtrans_url,omitempty"`
	RawNotification *string    `json:"-"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CheckoutRequest defines request body for checking out an order payment.
type CheckoutRequest struct {
	Method string `json:"method" validate:"required,oneof=cash midtrans"`
}

// CheckoutResponse represents the checkout output sent to client.
type CheckoutResponse struct {
	Payment       Payment `json:"payment"`
	MidtransToken string  `json:"midtrans_token,omitempty"`
	MidtransURL   string  `json:"midtrans_url,omitempty"`
}
