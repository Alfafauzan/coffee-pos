package entity

import "time"

// Category represents a product category entity in the database.
type Category struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

// CreateCategoryRequest defines request body for creating a new product category.
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// UpdateCategoryRequest defines request body for updating an existing product category.
type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}
