package entity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

func TestTable_JSONSerialization(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)
	table := Table{
		ID:        "tbl-001",
		Name:      "Meja 01",
		Capacity:  4,
		Status:    TableStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: &deletedAt,
	}

	data, err := json.Marshal(table)
	if err != nil {
		t.Fatalf("failed to marshal table: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, "deleted_at") {
		t.Errorf("expected deleted_at to be omitted from json, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"status":"available"`) {
		t.Errorf("expected status 'available' in json, got: %s", jsonStr)
	}
}

func TestTable_Validation(t *testing.T) {
	v := validator.New()

	// Valid CreateTableRequest
	validCreate := CreateTableRequest{
		Name:     "T-01",
		Capacity: 4,
	}
	if errs := v.Validate(validCreate); errs != nil {
		t.Errorf("expected no validation errors, got: %v", errs)
	}

	// Invalid CreateTableRequest (empty name, capacity 0)
	invalidCreate := CreateTableRequest{
		Name:     "",
		Capacity: 0,
	}
	errs := v.Validate(invalidCreate)
	if errs == nil {
		t.Fatal("expected errors, got nil")
	}
	if errs["name"] != "wajib diisi" {
		t.Errorf("expected 'wajib diisi' for name, got: %s", errs["name"])
	}
	if errs["capacity"] != "wajib diisi" {
		t.Errorf("expected 'wajib diisi' for capacity 0, got: %s", errs["capacity"])
	}

	// Invalid capacity (> 50)
	overCapacity := CreateTableRequest{
		Name:     "Big Table",
		Capacity: 51,
	}
	errs = v.Validate(overCapacity)
	if errs == nil || errs["capacity"] != "maksimal 50" {
		t.Errorf("expected 'maksimal 50' for capacity 51, got: %v", errs)
	}

	// Valid UpdateTableRequest (empty is valid)
	emptyUpdate := UpdateTableRequest{}
	if errs := v.Validate(emptyUpdate); errs != nil {
		t.Errorf("expected empty update to be valid, got: %v", errs)
	}

	// Valid UpdateTableRequest with status
	validStatuses := []string{TableStatusAvailable, TableStatusOccupied, TableStatusReserved}
	for _, st := range validStatuses {
		req := UpdateTableRequest{Status: st}
		if errs := v.Validate(req); errs != nil {
			t.Errorf("expected status '%s' to be valid, got: %v", st, errs)
		}
	}

	// Invalid status
	invalidStatusReq := UpdateTableRequest{
		Status: "broken",
	}
	errs = v.Validate(invalidStatusReq)
	if errs == nil || errs["status"] != "harus salah satu dari: available occupied reserved" {
		t.Errorf("expected oneof error for invalid status, got: %v", errs)
	}
}
