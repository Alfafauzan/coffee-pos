package entity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

func TestOrder_JSONSerialization(t *testing.T) {
	now := time.Now()
	tableID := "tbl-123"
	promoID := "prm-456"

	order := Order{
		ID:        "ord-001",
		ShiftID:   "shift-001",
		CashierID: "user-001",
		Cashier:   &User{ID: "user-001", Name: "Budi"},
		TableID:   &tableID,
		Table:     &Table{ID: "tbl-123", Name: "Meja 1"},
		PromoID:   &promoID,
		Promo:     &Promo{ID: "prm-456", Code: "HEMAT10"},
		Items: []OrderItem{
			{
				ID:        "item-001",
				OrderID:   "ord-001",
				ProductID: "prod-001",
				Product:   &Product{ID: "prod-001", Name: "Latte"},
				Quantity:  2,
				Price:     2500000,
				Subtotal:  5000000,
				Notes:     "Less ice",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Status:         OrderStatusPaid,
		Subtotal:       5000000,
		DiscountAmount: 1000000,
		Total:          4000000,
		Notes:          "Customer VIP",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("failed to marshal order: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"status":"paid"`) {
		t.Errorf("expected status 'paid' in json, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"table_id":"tbl-123"`) {
		t.Errorf("expected table_id in json, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"items":[{"id":"item-001"`) {
		t.Errorf("expected items in json, got: %s", jsonStr)
	}
}

func TestOrder_ValidationRequests(t *testing.T) {
	v := validator.New()

	// 1. CreateOrderRequest
	validCreate := CreateOrderRequest{
		TableID: "c1f73b64-845f-4a37-b6bb-293774659f10",
		Notes:   "Dekat jendela",
	}
	if errs := v.Validate(validCreate); errs != nil {
		t.Errorf("expected valid CreateOrderRequest, got: %v", errs)
	}

	invalidCreate := CreateOrderRequest{
		TableID: "not-a-uuid",
		Notes:   strings.Repeat("a", 501),
	}
	errs := v.Validate(invalidCreate)
	if errs == nil || errs["table_id"] != "format UUID tidak valid" || errs["notes"] != "maksimal 500 karakter" {
		t.Errorf("expected errors on invalid CreateOrderRequest, got: %v", errs)
	}

	// 2. AddOrderItemRequest
	validAdd := AddOrderItemRequest{
		ProductID: "c1f73b64-845f-4a37-b6bb-293774659f10",
		Quantity:  2,
		Notes:     "Less sugar",
	}
	if errs := v.Validate(validAdd); errs != nil {
		t.Errorf("expected valid AddOrderItemRequest, got: %v", errs)
	}

	invalidAdd := AddOrderItemRequest{
		ProductID: "bad-uuid",
		Quantity:  101, // max 100
		Notes:     strings.Repeat("b", 201),
	}
	errs = v.Validate(invalidAdd)
	if errs == nil || errs["product_id"] != "format UUID tidak valid" || errs["quantity"] != "maksimal 100" || errs["notes"] != "maksimal 200 karakter" {
		t.Errorf("expected errors on invalid AddOrderItemRequest, got: %v", errs)
	}

	// 3. UpdateOrderItemRequest
	validUpdate := UpdateOrderItemRequest{
		Quantity: 5,
		Notes:    "Updated note",
	}
	if errs := v.Validate(validUpdate); errs != nil {
		t.Errorf("expected valid UpdateOrderItemRequest, got: %v", errs)
	}

	invalidUpdate := UpdateOrderItemRequest{
		Quantity: 0, // min 1
	}
	errs = v.Validate(invalidUpdate)
	if errs == nil || errs["quantity"] != "wajib diisi" { // 0 triggers required
		t.Errorf("expected error on invalid UpdateOrderItemRequest quantity 0, got: %v", errs)
	}

	// 4. ApplyPromoRequest
	validPromo := ApplyPromoRequest{
		PromoCode: "HEMAT20",
	}
	if errs := v.Validate(validPromo); errs != nil {
		t.Errorf("expected valid ApplyPromoRequest, got: %v", errs)
	}

	invalidPromo := ApplyPromoRequest{
		PromoCode: "AB", // min 3
	}
	errs = v.Validate(invalidPromo)
	if errs == nil || errs["promo_code"] != "minimal 3 karakter" {
		t.Errorf("expected error on short PromoCode, got: %v", errs)
	}
}
