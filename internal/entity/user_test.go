package entity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

func TestUserRoles(t *testing.T) {
	owner := &User{Role: RoleOwner}
	if !owner.IsOwner() {
		t.Errorf("expected IsOwner() true for owner, got false")
	}
	if owner.IsCashier() {
		t.Errorf("expected IsCashier() false for owner, got true")
	}

	cashier := &User{Role: RoleCashier}
	if !cashier.IsCashier() {
		t.Errorf("expected IsCashier() true for cashier, got false")
	}
	if cashier.IsOwner() {
		t.Errorf("expected IsOwner() false for cashier, got true")
	}
}

func TestUser_ToResponse(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:        "user-123",
		Name:      "Budi",
		Email:     "budi@example.com",
		Password:  "hashed_secret",
		Role:      RoleCashier,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	res := user.ToResponse()
	if res.ID != user.ID || res.Name != user.Name || res.Email != user.Email || res.Role != user.Role || res.IsActive != user.IsActive {
		t.Errorf("ToResponse mismatch: got %+v", res)
	}

	// Verify json serialization hides password
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}
	if strings.Contains(string(data), "hashed_secret") || strings.Contains(string(data), "password") {
		t.Errorf("user json should not contain password field or value, got: %s", string(data))
	}
}

func TestValidation_LoginRequest(t *testing.T) {
	v := validator.New()

	validReq := LoginRequest{
		Email:    "budi@example.com",
		Password: "password123",
	}
	if errs := v.Validate(validReq); errs != nil {
		t.Errorf("expected no validation errors, got: %v", errs)
	}

	invalidReq := LoginRequest{
		Email:    "invalid-email",
		Password: "123",
	}
	errs := v.Validate(invalidReq)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if _, ok := errs["email"]; !ok {
		t.Errorf("expected error for email field")
	}
	if _, ok := errs["password"]; !ok {
		t.Errorf("expected error for password field")
	}
}

func TestValidation_RegisterRequest(t *testing.T) {
	v := validator.New()

	validReq := RegisterRequest{
		Name:     "Siti",
		Email:    "siti@example.com",
		Password: "password123",
		Role:     RoleOwner,
	}
	if errs := v.Validate(validReq); errs != nil {
		t.Errorf("expected no validation errors, got: %v", errs)
	}

	invalidReq := RegisterRequest{
		Name:     "S",
		Email:    "siti",
		Password: "123",
		Role:     "superuser",
	}
	errs := v.Validate(invalidReq)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	for _, f := range []string{"name", "email", "password", "role"} {
		if _, ok := errs[f]; !ok {
			t.Errorf("expected error for field %s", f)
		}
	}
}
