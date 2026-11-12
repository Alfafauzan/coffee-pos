package entity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

func TestShift_JSONSerialization(t *testing.T) {
	now := time.Now()
	closedAt := now.Add(8 * time.Hour)
	closingCash := int64(50000000)

	shift := Shift{
		ID:                "shift-001",
		CashierID:         "user-002",
		Cashier:           &User{ID: "user-002", Name: "Budi"},
		OpenedAt:          now,
		ClosedAt:          &closedAt,
		OpeningCash:       20000000,
		ClosingCash:       &closingCash,
		TotalTransactions: 45,
		Status:            ShiftStatusClosed,
		Notes:             "Normal shift closing",
		CreatedAt:         now,
		UpdatedAt:         closedAt,
	}

	data, err := json.Marshal(shift)
	if err != nil {
		t.Fatalf("failed to marshal shift: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"status":"closed"`) {
		t.Errorf("expected status 'closed', got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"cashier":{"id":"user-002","name":"Budi"`) {
		t.Errorf("expected nested cashier in json, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"closing_cash":50000000`) {
		t.Errorf("expected closing_cash in json, got: %s", jsonStr)
	}
}

func TestShift_Validation(t *testing.T) {
	v := validator.New()

	// Valid OpenShiftRequest
	validOpen := OpenShiftRequest{
		OpeningCash: 20000000, // Rp 200.000 (20000000 sen)
		Notes:       "Pecahan uang kecil lengkap",
	}
	if errs := v.Validate(validOpen); errs != nil {
		t.Errorf("expected no error for valid OpenShiftRequest, got: %v", errs)
	}

	// Valid OpenShiftRequest with 0 cash
	zeroCashOpen := OpenShiftRequest{
		OpeningCash: 0,
	}
	if errs := v.Validate(zeroCashOpen); errs != nil {
		t.Errorf("expected 0 cash to be valid for OpenShiftRequest, got: %v", errs)
	}

	// Invalid OpenShiftRequest (negative cash)
	negOpen := OpenShiftRequest{
		OpeningCash: -100,
	}
	errs := v.Validate(negOpen)
	if errs == nil || errs["opening_cash"] != "minimal 0" {
		t.Errorf("expected 'minimal 0' for negative opening cash, got: %v", errs)
	}

	// Valid CloseShiftRequest
	validClose := CloseShiftRequest{
		ClosingCash: 52000000,
		Notes:       "Rekonsiliasi cocok",
	}
	if errs := v.Validate(validClose); errs != nil {
		t.Errorf("expected no error for valid CloseShiftRequest, got: %v", errs)
	}

	// Invalid CloseShiftRequest (negative cash, notes too long)
	invalidClose := CloseShiftRequest{
		ClosingCash: -500,
		Notes:       strings.Repeat("a", 501),
	}
	errs = v.Validate(invalidClose)
	if errs == nil {
		t.Fatal("expected errors, got nil")
	}
	if errs["closing_cash"] != "minimal 0" {
		t.Errorf("expected 'minimal 0' for closing_cash, got: %s", errs["closing_cash"])
	}
	if errs["notes"] != "maksimal 500 karakter" {
		t.Errorf("expected 'maksimal 500 karakter' for notes, got: %s", errs["notes"])
	}
}
