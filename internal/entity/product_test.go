package entity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

func TestProduct_PriceInRupiah(t *testing.T) {
	product := &Product{
		Price: 2500000, // Rp 25.000,00
	}
	expected := 25000.0
	if got := product.PriceInRupiah(); got != expected {
		t.Errorf("expected %.2f rupiah, got %.2f", expected, got)
	}
}

func TestProduct_JSONSerialization(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)
	product := Product{
		ID:          "prod-123",
		CategoryID:  "cat-456",
		Category:    &Category{ID: "cat-456", Name: "Coffee"},
		Name:        "Iced Americano",
		Description: "Cold water and espresso",
		Price:       2500000,
		Stock:       50,
		ImageURL:    "https://example.com/img.png",
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   &deletedAt,
	}

	data, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("failed to marshal product: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, "deleted_at") {
		t.Errorf("expected deleted_at to be omitted from json, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"category":{"id":"cat-456","name":"Coffee"`) {
		t.Errorf("expected nested category in json, got: %s", jsonStr)
	}
}

func TestProduct_Validation(t *testing.T) {
	v := validator.New()

	// Valid CreateProductRequest
	validCreate := CreateProductRequest{
		CategoryID:  "c1f73b64-845f-4a37-b6bb-293774659f10",
		Name:        "Iced Latte",
		Description: "Fresh milk & espresso",
		Price:       2800000,
		ImageURL:    "https://example.com/latte.jpg",
	}
	if errs := v.Validate(validCreate); errs != nil {
		t.Errorf("expected no validation errors, got: %v", errs)
	}

	// Invalid CreateProductRequest
	invalidCreate := CreateProductRequest{
		CategoryID: "invalid-uuid",
		Name:       "A",
		Price:      50, // under 100 sen
		ImageURL:   "not-a-url",
	}
	errs := v.Validate(invalidCreate)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if errs["category_id"] != "format UUID tidak valid" {
		t.Errorf("expected UUID error, got: %s", errs["category_id"])
	}
	if errs["name"] != "minimal 2 karakter" {
		t.Errorf("expected name min error, got: %s", errs["name"])
	}
	if errs["price"] != "minimal 100" {
		t.Errorf("expected price min error, got: %s", errs["price"])
	}
	if errs["image_url"] != "tidak valid" {
		t.Errorf("expected url error, got: %s", errs["image_url"])
	}

	// Valid UpdateProductRequest (empty is valid because all are omitempty)
	emptyUpdate := UpdateProductRequest{}
	if errs := v.Validate(emptyUpdate); errs != nil {
		t.Errorf("expected empty update to be valid, got: %v", errs)
	}

	// Valid UpdateProductRequest with IsActive false
	isActiveFalse := false
	partialUpdate := UpdateProductRequest{
		IsActive: &isActiveFalse,
	}
	if errs := v.Validate(partialUpdate); errs != nil {
		t.Errorf("expected update with isActive=false to be valid, got: %v", errs)
	}
}
