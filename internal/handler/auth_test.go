package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/Alfafauzan/coffee-pos/internal/entity"
	"github.com/Alfafauzan/coffee-pos/internal/repository"
	"github.com/Alfafauzan/coffee-pos/internal/service"
	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

type mockUserRepoForHandler struct {
	users map[string]*entity.User
}

func (m *mockUserRepoForHandler) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return nil, nil
}

func (m *mockUserRepoForHandler) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepoForHandler) Create(ctx context.Context, user *entity.User) error {
	user.ID = "generated-uuid-user"
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepoForHandler) Update(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *mockUserRepoForHandler) FindAll(ctx context.Context) ([]entity.User, error) {
	return nil, nil
}

func (m *mockUserRepoForHandler) WithTx(tx *sql.Tx) repository.UserRepository {
	return m
}

func setupAuthHandlerTest() (*AuthHandler, *mockUserRepoForHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	repo := &mockUserRepoForHandler{users: make(map[string]*entity.User)}
	authService := service.NewAuthService(repo, "test-secret-key-for-handlers-1234", 24)
	v := validator.New()
	h := NewAuthHandler(authService, v)

	r := gin.New()
	r.POST("/api/v1/auth/register", h.Register)
	r.POST("/api/v1/auth/login", h.Login)

	return h, repo, r
}

func TestAuthHandler_RegisterSuccess(t *testing.T) {
	_, _, r := setupAuthHandlerTest()

	body := map[string]interface{}{
		"name":     "Barista Lisa",
		"email":    "lisa@coffee.com",
		"password": "password123",
		"role":     "cashier",
	}
	jsonBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["success"] != true {
		t.Errorf("expected success true, got %v", res["success"])
	}
	if res["message"] != "Registrasi berhasil" {
		t.Errorf("expected message 'Registrasi berhasil', got %v", res["message"])
	}
}

func TestAuthHandler_RegisterInvalidFormat(t *testing.T) {
	_, _, r := setupAuthHandlerTest()

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString("invalid json format"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_RegisterValidationError(t *testing.T) {
	_, _, r := setupAuthHandlerTest()

	body := map[string]interface{}{
		"name":     "L", // min 2
		"email":    "not-an-email",
		"password": "123", // min 6
		"role":     "admin", // must be owner or cashier
	}
	jsonBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d (UnprocessableEntity), got %d. Body: %s", http.StatusUnprocessableEntity, w.Code, w.Body.String())
	}
}

func TestAuthHandler_RegisterDuplicateEmail(t *testing.T) {
	_, repo, r := setupAuthHandlerTest()
	repo.users["existing@coffee.com"] = &entity.User{
		ID:    "u-1",
		Email: "existing@coffee.com",
	}

	body := map[string]interface{}{
		"name":     "Duplicate User",
		"email":    "existing@coffee.com",
		"password": "password123",
		"role":     "cashier",
	}
	jsonBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_LoginSuccess(t *testing.T) {
	_, repo, r := setupAuthHandlerTest()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.users["user@coffee.com"] = &entity.User{
		ID:        "u-123",
		Name:      "Cashier User",
		Email:     "user@coffee.com",
		Password:  string(hash),
		Role:      "cashier",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	body := map[string]interface{}{
		"email":    "user@coffee.com",
		"password": "password123",
	}
	jsonBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["message"] != "Login berhasil" {
		t.Errorf("expected message 'Login berhasil', got %v", res["message"])
	}
}

func TestAuthHandler_LoginInvalidCredentials(t *testing.T) {
	_, _, r := setupAuthHandlerTest()

	body := map[string]interface{}{
		"email":    "nonexistent@coffee.com",
		"password": "password123",
	}
	jsonBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
