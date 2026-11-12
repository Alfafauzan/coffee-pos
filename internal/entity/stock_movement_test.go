package entity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

func TestStockMovement_JSONSerialization(t *testing.T) {
	now := time.Now()
	movement := StockMovement{
		ID:          "sm-001",
		ProductID:   "prod-001",
		Product:     &Product{ID: "prod-001", Name: "Espresso Beans"},
		UserID:      "user-001",
		User:        &User{ID: "user-001", Name: "Admin"},
		Type:        MovementTypeIn,
		Quantity:    10,
		StockBefore: 5,
		StockAfter:  15,
		Notes:       "Restock from supplier",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(movement)
	if err != nil {
		t.Fatalf("failed to marshal stock movement: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"type":"in"`) {
		t.Errorf("expected type in json, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"product":{"id":"prod-001","category_id":""`) {
		t.Errorf("expected nested product, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"user":{"id":"user-001","name":"Admin"`) {
		t.Errorf("expected nested user, got: %s", jsonStr)
	}
}

func TestStockAdjustmentRequest_Validation(t *testing.T) {
	v := validator.New()

	// Valid cases for in, out, adjustment
	validTypes := []string{MovementTypeIn, MovementTypeOut, MovementTypeAdjustment}
	for _, mType := range validTypes {
		req := StockAdjustmentRequest{
			Quantity: 5,
			Type:     mType,
			Notes:    "Valid adjustment notes",
		}
		if errs := v.Validate(req); errs != nil {
			t.Errorf("expected type '%s' to be valid, got errors: %v", mType, errs)
		}
	}

	// Invalid quantity (0 triggers required)
	zeroQty := StockAdjustmentRequest{
		Quantity: 0,
		Type:     MovementTypeIn,
	}
	errs := v.Validate(zeroQty)
	if errs == nil || errs["quantity"] != "wajib diisi" {
		t.Errorf("expected 'wajib diisi' for quantity 0, got: %v", errs)
	}

	// Invalid quantity (-1 triggers min=1)
	negQty := StockAdjustmentRequest{
		Quantity: -1,
		Type:     MovementTypeIn,
	}
	errs = v.Validate(negQty)
	if errs == nil || errs["quantity"] != "minimal 1" {
		t.Errorf("expected 'minimal 1' for quantity -1, got: %v", errs)
	}

	// Invalid type
	invalidType := StockAdjustmentRequest{
		Quantity: 10,
		Type:     "unknown_type",
	}
	errs = v.Validate(invalidType)
	if errs == nil || errs["type"] != "harus salah satu dari: in out adjustment" {
		t.Errorf("expected oneof error for invalid type, got: %v", errs)
	}

	// Invalid notes (>500 chars)
	longNotes := StockAdjustmentRequest{
		Quantity: 1,
		Type:     MovementTypeAdjustment,
		Notes:    strings.Repeat("a", 501),
	}
	errs = v.Validate(longNotes)
	if errs == nil || errs["notes"] != "maksimal 500 karakter" {
		t.Errorf("expected 'maksimal 500 karakter' for long notes, got: %v", errs)
	}
}
