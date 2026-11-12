package entity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

func TestCategory_JSONSerialization(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)
	cat := Category{
		ID:        "cat-123",
		Name:      "Espresso Based",
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: &deletedAt,
	}

	data, err := json.Marshal(cat)
	if err != nil {
		t.Fatalf("failed to marshal category: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, "deleted_at") {
		t.Errorf("expected deleted_at to be ignored in json serialization, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"name":"Espresso Based"`) {
		t.Errorf("expected name in json, got: %s", jsonStr)
	}
}

func TestCategory_Validation(t *testing.T) {
	v := validator.New()

	// Valid CreateCategoryRequest
	createReq := CreateCategoryRequest{Name: "Manual Brew"}
	if errs := v.Validate(createReq); errs != nil {
		t.Errorf("expected no errors for valid CreateCategoryRequest, got: %v", errs)
	}

	// Invalid CreateCategoryRequest (empty)
	emptyCreate := CreateCategoryRequest{Name: ""}
	errs := v.Validate(emptyCreate)
	if errs == nil || errs["name"] != "wajib diisi" {
		t.Errorf("expected 'wajib diisi' for empty name, got: %v", errs)
	}

	// Invalid CreateCategoryRequest (too short)
	shortCreate := CreateCategoryRequest{Name: "A"}
	errs = v.Validate(shortCreate)
	if errs == nil || errs["name"] != "minimal 2 karakter" {
		t.Errorf("expected 'minimal 2 karakter' for 1-char name, got: %v", errs)
	}

	// Valid UpdateCategoryRequest
	updateReq := UpdateCategoryRequest{Name: "Signature Drinks"}
	if errs := v.Validate(updateReq); errs != nil {
		t.Errorf("expected no errors for valid UpdateCategoryRequest, got: %v", errs)
	}

	// Invalid UpdateCategoryRequest (too long)
	longUpdate := UpdateCategoryRequest{Name: strings.Repeat("A", 101)}
	errs = v.Validate(longUpdate)
	if errs == nil || errs["name"] != "maksimal 100 karakter" {
		t.Errorf("expected 'maksimal 100 karakter' for long name, got: %v", errs)
	}
}
