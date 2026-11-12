package entity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

func TestPayment_JSONSerialization(t *testing.T) {
	now := time.Now()
	rawNotif := `{"status":"settlement","gross_amount":"50000"}`
	token := "snap-token-123"
	url := "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-token-123"
	orderID := "midtrans-order-001"

	payment := Payment{
		ID:              "pay-001",
		OrderID:         "ord-001",
		Method:          PaymentMethodMidtrans,
		Status:          PaymentStatusPaid,
		Amount:          5000000,
		MidtransOrderID: &orderID,
		MidtransToken:   &token,
		MidtransURL:     &url,
		RawNotification: &rawNotif,
		PaidAt:          &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	data, err := json.Marshal(payment)
	if err != nil {
		t.Fatalf("failed to marshal payment: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, "raw_notification") || strings.Contains(jsonStr, "gross_amount") {
		t.Errorf("raw_notification must be hidden in json, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"status":"paid"`) {
		t.Errorf("expected status 'paid', got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"midtrans_token":"snap-token-123"`) {
		t.Errorf("expected midtrans_token in json, got: %s", jsonStr)
	}
}

func TestCheckoutResponse_JSON(t *testing.T) {
	// Cash checkout response (no midtrans token/url)
	cashResp := CheckoutResponse{
		Payment: Payment{
			ID:     "pay-cash-1",
			Method: PaymentMethodCash,
			Status: PaymentStatusPaid,
			Amount: 2500000,
		},
	}
	cashData, _ := json.Marshal(cashResp)
	if strings.Contains(string(cashData), "midtrans_token") || strings.Contains(string(cashData), "midtrans_url") {
		t.Errorf("midtrans fields should be omitted for cash response: %s", string(cashData))
	}
}

func TestCheckoutRequest_Validation(t *testing.T) {
	v := validator.New()

	// Valid cash
	validCash := CheckoutRequest{Method: PaymentMethodCash}
	if errs := v.Validate(validCash); errs != nil {
		t.Errorf("expected cash method valid, got: %v", errs)
	}

	// Valid midtrans
	validMidtrans := CheckoutRequest{Method: PaymentMethodMidtrans}
	if errs := v.Validate(validMidtrans); errs != nil {
		t.Errorf("expected midtrans method valid, got: %v", errs)
	}

	// Invalid method
	invalidMethod := CheckoutRequest{Method: "crypto"}
	errs := v.Validate(invalidMethod)
	if errs == nil || errs["method"] != "harus salah satu dari: cash midtrans" {
		t.Errorf("expected oneof error for method, got: %v", errs)
	}
}
