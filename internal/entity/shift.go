package entity

import "time"

// Shift status constants
const (
	ShiftStatusOpen   = "open"
	ShiftStatusClosed = "closed"
)

// Shift represents a cashier work shift session in the database.
type Shift struct {
	ID                string     `json:"id"`
	CashierID         string     `json:"cashier_id"`
	Cashier           *User      `json:"cashier,omitempty"`
	OpenedAt          time.Time  `json:"opened_at"`
	ClosedAt          *time.Time `json:"closed_at,omitempty"`
	OpeningCash       int64      `json:"opening_cash"`
	ClosingCash       *int64     `json:"closing_cash,omitempty"`
	TotalTransactions int64      `json:"total_transactions"`
	Status            string     `json:"status"`
	Notes             string     `json:"notes,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// OpenShiftRequest defines request body for opening a cashier work shift.
type OpenShiftRequest struct {
	OpeningCash int64  `json:"opening_cash" validate:"min=0"`
	Notes       string `json:"notes" validate:"max=500"`
}

// CloseShiftRequest defines request body for closing a cashier work shift.
type CloseShiftRequest struct {
	ClosingCash int64  `json:"closing_cash" validate:"min=0"`
	Notes       string `json:"notes" validate:"max=500"`
}
