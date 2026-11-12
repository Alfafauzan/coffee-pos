package validator

import (
	"testing"
)

type TestDTO struct {
	Name      string `json:"name" validate:"required,min=3,max=10"`
	Email     string `json:"email_address" validate:"required,email"`
	Age       int    `json:"user_age" validate:"min=18,max=65"`
	Role      string `json:"role" validate:"oneof=owner cashier"`
	UUID      string `json:"id_uuid" validate:"uuid4"`
	Code      string `json:"security_code" validate:"len=6"`
	Phone     string `json:"phone_number" validate:"numeric"`
	CustomURL string `json:"website_url" validate:"url"`
}

func TestValidator_Valid(t *testing.T) {
	v := New()
	dto := TestDTO{
		Name:      "Budi",
		Email:     "budi@example.com",
		Age:       25,
		Role:      "cashier",
		UUID:      "c1f73b64-845f-4a37-b6bb-293774659f10",
		Code:      "123456",
		Phone:     "08123456789",
		CustomURL: "https://example.com",
	}

	errs := v.Validate(dto)
	if errs != nil {
		t.Fatalf("expected nil for valid struct, got: %v", errs)
	}

	// Test pointer
	errsPtr := v.Validate(&dto)
	if errsPtr != nil {
		t.Fatalf("expected nil for valid pointer, got: %v", errsPtr)
	}
}

func TestValidator_Errors(t *testing.T) {
	v := New()
	dto := TestDTO{
		Name:      "Al",       // min=3 error
		Email:     "not-an-email", // email error
		Age:       16,         // min=18 error
		Role:      "admin",    // oneof error
		UUID:      "invalid-uuid", // uuid4 error
		Code:      "123",      // len=6 error
		Phone:     "abc123",   // numeric error
		CustomURL: "not-a-url", // default ("tidak valid") error
	}

	errs := v.Validate(dto)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	expected := map[string]string{
		"name":          "minimal 3 karakter",
		"email_address": "format email tidak valid",
		"user_age":      "minimal 18",
		"role":          "harus salah satu dari: owner cashier",
		"id_uuid":       "format UUID tidak valid",
		"security_code": "harus tepat 6 karakter",
		"phone_number":  "harus berupa angka",
		"website_url":   "tidak valid",
	}

	for field, expectedMsg := range expected {
		actualMsg, exists := errs[field]
		if !exists {
			t.Errorf("expected field '%s' in errors, but was missing", field)
			continue
		}
		if actualMsg != expectedMsg {
			t.Errorf("field '%s': expected message '%s', got '%s'", field, expectedMsg, actualMsg)
		}
	}
}

func TestValidator_MaxRulesAndRequired(t *testing.T) {
	v := New()
	type MaxDTO struct {
		Title  string `json:"book_title" validate:"required,max=5"`
		Score  int    `json:"student_score" validate:"max=100"`
	}

	dto := MaxDTO{
		Title: "TooLongTitle",
		Score: 105,
	}

	errs := v.Validate(dto)
	if errs == nil {
		t.Fatal("expected errors, got nil")
	}

	if errs["book_title"] != "maksimal 5 karakter" {
		t.Errorf("expected 'maksimal 5 karakter', got: %s", errs["book_title"])
	}
	if errs["student_score"] != "maksimal 100" {
		t.Errorf("expected 'maksimal 100', got: %s", errs["student_score"])
	}

	emptyDTO := MaxDTO{}
	emptyErrs := v.Validate(emptyDTO)
	if emptyErrs["book_title"] != "wajib diisi" {
		t.Errorf("expected 'wajib diisi', got: %s", emptyErrs["book_title"])
	}
}
