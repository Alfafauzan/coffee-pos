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

// ProductFilter defines filtering and pagination parameters for listing products.
type ProductFilter struct {
	CategoryID string
	IsActive   *bool
	Search     string
	Page       int
	Limit      int
}

// ProductRepository specifies data access operations for Product entities.
type ProductRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Product, error)
	FindAll(ctx context.Context, filter ProductFilter) ([]entity.Product, int, error)
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, id string, stock int) error
	WithTx(tx *sql.Tx) ProductRepository
}

type productRepository struct {
	db sqlDB
}

// NewProductRepository creates a new ProductRepository implementation.
func NewProductRepository(db sqlDB) ProductRepository {
	return &productRepository{
		db: db,
	}
}

// WithTx returns a new ProductRepository instance operating within the provided transaction.
func (r *productRepository) WithTx(tx *sql.Tx) ProductRepository {
	return &productRepository{db: tx}
}

// FindByID retrieves an active product by ID, left-joining category information.
// Returns nil, nil if the product is not found.
func (r *productRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	query := `
		SELECT p.id, p.category_id, p.name, p.description, p.price, p.stock,
		       p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
		WHERE p.id = ? AND p.deleted_at IS NULL
	`

	var (
		product     entity.Product
		category entity.Category
		description sql.NullString
		imageURL    sql.NullString
		catID       sql.NullString
		catName     sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.CategoryID,
		&product.Name,
		&description,
		&product.Price,
		&product.Stock,
		&imageURL,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
		&catID,
		&catName,
		&category.ID,
		&category.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if description.Valid {
		product.Description = description.String
	}
	if imageURL.Valid {
		product.ImageURL = imageURL.String
	}
	if catID.Valid {
		product.Category = &entity.Category{
			ID:   catID.String,
			Name: catName.String,
		}
	}

	return &product, nil
}

// FindAll retrieves paginated products matching dynamic filter criteria,
// returning products slice with populated categories, total count, and error.
func (r *productRepository) FindAll(ctx context.Context, filter ProductFilter) ([]entity.Product, int, error) {
	conditions := []string{"p.deleted_at IS NULL"}
	var args []interface{}

	if filter.CategoryID != "" {
		conditions = append(conditions, "p.category_id = ?")
		args = append(args, filter.CategoryID)
	}

	if filter.IsActive != nil {
		conditions = append(conditions, "p.is_active = ?")
		args = append(args, *filter.IsActive)
	}

	if filter.Search != "" {
		conditions = append(conditions, "p.name LIKE ?")
		args = append(args, "%"+filter.Search+"%")
	}

	whereClause := strings.Join(conditions, " AND ")

	// 1. COUNT query for total matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM products p
		WHERE %s
	`, whereClause)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return make([]entity.Product, 0), 0, nil
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

	// 3. Data query with LIMIT and OFFSET
	dataQuery := fmt.Sprintf(`
		SELECT p.id, p.category_id, p.name, p.description, p.price, p.stock,
		       p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
		WHERE %s
		ORDER BY p.created_at DESC
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

	products := make([]entity.Product, 0)
	for rows.Next() {
		var (
			product     entity.Product
			description sql.NullString
			imageURL    sql.NullString
			catID       sql.NullString
			catName     sql.NullString
		)

		err := rows.Scan(
			&product.ID,
			&product.CategoryID,
			&product.Name,
			&description,
			&product.Price,
			&product.Stock,
			&imageURL,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
			&catID,
			&catName,
		)
		if err != nil {
			return nil, 0, err
		}

		if description.Valid {
			product.Description = description.String
		}
		if imageURL.Valid {
			product.ImageURL = imageURL.String
		}
		if catID.Valid {
			product.Category = &entity.Category{
				ID:   catID.String,
				Name: catName.String,
			}
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// Create inserts a new product with generated UUID and current timestamps.
func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
	product.ID = uuid.New().String()
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	query := `
		INSERT INTO products (
			id, category_id, name, description, price, stock,
			image_url, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		product.ID,
		product.CategoryID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.ImageURL,
		product.IsActive,
		product.CreatedAt,
		product.UpdatedAt,
	)
	return err
}

// Update updates product information including category_id, name, description, price, image_url, is_active.
func (r *productRepository) Update(ctx context.Context, product *entity.Product) error {
	product.UpdatedAt = time.Now()

	query := `
		UPDATE products
		SET category_id = ?, name = ?, description = ?, price = ?,
		    image_url = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		product.CategoryID,
		product.Name,
		product.Description,
		product.Price,
		product.ImageURL,
		product.IsActive,
		product.UpdatedAt,
		product.ID,
	)
	return err
}

// Delete marks a product as soft-deleted by setting deleted_at to NOW().
func (r *productRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE products
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// UpdateStock updates product stock count and updated_at timestamp.
// CATATAN: method ini tidak boleh dipanggil langsung dari handler.
// Selalu dipanggil bersamaan dengan StockRepository.Create() dalam satu
// database transaction — akan diimplementasikan di lesson Database Transaction di Go.
func (r *productRepository) UpdateStock(ctx context.Context, id string, stock int) error {
	now := time.Now()
	query := `
		UPDATE products
		SET stock = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, stock, now, id)
	return err
}
