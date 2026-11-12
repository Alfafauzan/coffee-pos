package entity

import "time"

// Product represents a coffee shop product entity in the database.
type Product struct {
	ID          string     `json:"id"`
	CategoryID  string     `json:"category_id"`
	Category    *Category  `json:"category,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Price       int64      `json:"price"`
	Stock       int        `json:"stock"`
	ImageURL    string     `json:"image_url,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"-"`
}

// PriceInRupiah converts price in sen to rupiah for display purposes.
func (p *Product) PriceInRupiah() float64 {
	return float64(p.Price) / 100.0
}

// CreateProductRequest defines request body for creating a new product.
type CreateProductRequest struct {
	CategoryID  string `json:"category_id" validate:"required,uuid4"`
	Name        string `json:"name" validate:"required,min=2,max=200"`
	Description string `json:"description" validate:"max=1000"`
	Price       int64  `json:"price" validate:"required,min=100"`
	ImageURL    string `json:"image_url" validate:"omitempty,url"`
}

// UpdateProductRequest defines request body for updating an existing product.
type UpdateProductRequest struct {
	CategoryID  string `json:"category_id" validate:"omitempty,uuid4"`
	Name        string `json:"name" validate:"omitempty,min=2,max=200"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Price       int64  `json:"price" validate:"omitempty,min=100"`
	ImageURL    string `json:"image_url" validate:"omitempty,url"`
	IsActive    *bool  `json:"is_active"`
}
