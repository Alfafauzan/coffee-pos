package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Alfafauzan/coffee-pos/config"
	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

func TestNewRouter_Health(t *testing.T) {
	cfg := &config.Config{
		AppEnv:         "development",
		JWTSecret:      "test-secret",
		JWTExpireHours: 24,
	}
	v := validator.New()
	router := NewRouter(&sql.DB{}, cfg, v)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("expected success to be true, got %v", response["success"])
	}

	if message, ok := response["message"].(string); !ok || message != "server is running" {
		t.Errorf("expected message 'server is running', got '%v'", response["message"])
	}
}

func TestNewRouter_Modes(t *testing.T) {
	v := validator.New()

	prodCfg := &config.Config{AppEnv: "production", JWTSecret: "secret", JWTExpireHours: 24}
	_ = NewRouter(&sql.DB{}, prodCfg, v)
	if gin.Mode() != gin.ReleaseMode {
		t.Errorf("expected ReleaseMode in production, got: %s", gin.Mode())
	}

	devCfg := &config.Config{AppEnv: "development", JWTSecret: "secret", JWTExpireHours: 24}
	_ = NewRouter(&sql.DB{}, devCfg, v)
	if gin.Mode() != gin.DebugMode {
		t.Errorf("expected DebugMode in development, got: %s", gin.Mode())
	}
}
