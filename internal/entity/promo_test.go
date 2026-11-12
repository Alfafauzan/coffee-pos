package entity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

func TestPromo_IsValid(t *testing.T) {
	now := time.Now()
	started := now.Add(-1 * time.Hour)
	ended := now.Add(1 * time.Hour)
	limit5 := 5

	// 1. Valid promo without limit
	promo := Promo{
		IsActive:  true,
		StartedAt: started,
		EndedAt:   ended,
	}
	if !promo.IsValid() {
		t.Errorf("expected IsValid true, got false")
	}

	// 2. Inactive promo
	inactive := promo
	inactive.IsActive = false
	if inactive.IsValid() {
		t.Errorf("expected IsValid false for inactive promo, got true")
	}

	// 3. Not yet started
	future := promo
	future.StartedAt = now.Add(1 * time.Hour)
	future.EndedAt = now.Add(2 * time.Hour)
	if future.IsValid() {
		t.Errorf("expected IsValid false for future promo, got true")
	}

	// 4. Expired promo
	past := promo
	past.StartedAt = now.Add(-2 * time.Hour)
	past.EndedAt = now.Add(-1 * time.Hour)
	if past.IsValid() {
		t.Errorf("expected IsValid false for past promo, got true")
	}

	// 5. Usage limit not reached
	withLimit := promo
	withLimit.UsageLimit = &limit5
	withLimit.UsedCount = 4
	if !withLimit.IsValid() {
		t.Errorf("expected IsValid true when usedCount < limit, got false")
	}

	// 6. Usage limit reached
	withLimit.UsedCount = 5
	if withLimit.IsValid() {
		t.Errorf("expected IsValid false when usedCount >= limit, got true")
	}
}

func TestPromo_JSONSerialization(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)
	promo := Promo{
		ID:        "prm-001",
		Name:      "Diskon 20%",
		Code:      "DISKON20",
		Type:      PromoTypePercentage,
		Value:     20,
		MinOrder:  5000000,
		StartedAt: now,
		EndedAt:   now.Add(24 * time.Hour),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: &deletedAt,
	}

	data, err := json.Marshal(promo)
	if err != nil {
		t.Fatalf("failed to marshal promo: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, "deleted_at") {
		t.Errorf("expected deleted_at omitted, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"code":"DISKON20"`) {
		t.Errorf("expected code in json, got: %s", jsonStr)
	}
}

func TestPromo_ValidationRequests(t *testing.T) {
	v := validator.New()
	now := time.Now()
	maxDisc := int64(1500000)
	limit := 100

	// Valid CreatePromoRequest
	validCreate := CreatePromoRequest{
		Name:        "Promo Gajian",
		Code:        "PAYDAY20",
		Type:        PromoTypePercentage,
		Value:       20,
		MinOrder:    5000000,
		MaxDiscount: &maxDisc,
		UsageLimit:  &limit,
		StartedAt:   now,
		EndedAt:     now.Add(7 * 24 * time.Hour),
	}
	if errs := v.Validate(validCreate); errs != nil {
		t.Errorf("expected valid CreatePromoRequest, got errors: %v", errs)
	}

	// Invalid CreatePromoRequest
	invalidCreate := CreatePromoRequest{
		Name:  "A",       // min 2
		Code:  "NO",      // min 3
		Type:  "invalid", // oneof percentage fixed
		Value: 0,         // min 1
	}
	errs := v.Validate(invalidCreate)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	if errs["name"] != "minimal 2 karakter" {
		t.Errorf("expected name error, got: %s", errs["name"])
	}
	if errs["code"] != "minimal 3 karakter" {
		t.Errorf("expected code error, got: %s", errs["code"])
	}
	if errs["type"] != "harus salah satu dari: percentage fixed" {
		t.Errorf("expected type error, got: %s", errs["type"])
	}
	if errs["value"] != "wajib diisi" {
		t.Errorf("expected value error, got: %s", errs["value"])
	}

	// Valid UpdatePromoRequest (all empty is valid)
	emptyUpdate := UpdatePromoRequest{}
	if errs := v.Validate(emptyUpdate); errs != nil {
		t.Errorf("expected empty UpdatePromoRequest to be valid, got: %v", errs)
	}
}
