package entity

import "time"

// Order status constants
const (
	OrderStatusDraft          = "draft"
	OrderStatusPendingPayment = "pending_payment"
	OrderStatusPaid           = "paid"
	OrderStatusCancelled      = "cancelled"
)

// OrderItem represents a single item line within a customer order.
type OrderItem struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	ProductID string    `json:"product_id"`
	Product   *Product  `json:"product,omitempty"`
	Quantity  int       `json:"quantity"`
	Price     int64     `json:"price"`
	Subtotal  int64     `json:"subtotal"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Order represents a customer transaction order in the database.
type Order struct {
	ID             string      `json:"id"`
	ShiftID        string      `json:"shift_id"`
	CashierID      string      `json:"cashier_id"`
	Cashier        *User       `json:"cashier,omitempty"`
	TableID        *string     `json:"table_id,omitempty"`
	Table          *Table      `json:"table,omitempty"`
	PromoID        *string     `json:"promo_id,omitempty"`
	Promo          *Promo      `json:"promo,omitempty"`
	Items          []OrderItem `json:"items,omitempty"`
	Status         string      `json:"status"`
	Subtotal       int64       `json:"subtotal"`
	DiscountAmount int64       `json:"discount_amount"`
	Total          int64       `json:"total"`
	Notes          string      `json:"notes,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// CreateOrderRequest defines request body for initializing a new order.
type CreateOrderRequest struct {
	TableID string `json:"table_id" validate:"omitempty,uuid4"`
	Notes   string `json:"notes" validate:"max=500"`
}

// AddOrderItemRequest defines request body for adding an item to an order.
type AddOrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid4"`
	Quantity  int    `json:"quantity" validate:"required,min=1,max=100"`
	Notes     string `json:"notes" validate:"max=200"`
}

// UpdateOrderItemRequest defines request body for updating item quantity or notes in an order.
type UpdateOrderItemRequest struct {
	Quantity int    `json:"quantity" validate:"required,min=1,max=100"`
	Notes    string `json:"notes" validate:"max=200"`
}

// ApplyPromoRequest defines request body for applying a discount promo to an order.
type ApplyPromoRequest struct {
	PromoCode string `json:"promo_code" validate:"required,min=3,max=50"`
}
