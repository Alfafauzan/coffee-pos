package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Alfafauzan/coffee-pos/internal/entity"
	"github.com/Alfafauzan/coffee-pos/internal/repository"
)

type mockUserRepo struct {
	usersByEmail map[string]*entity.User
	createFn     func(ctx context.Context, user *entity.User) error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByEmail: make(map[string]*entity.User),
	}
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	if u, ok := m.usersByEmail[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	user.ID = "test-uuid-1234"
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *mockUserRepo) FindAll(ctx context.Context) ([]entity.User, error) {
	return nil, nil
}

func (m *mockUserRepo) WithTx(tx *sql.Tx) repository.UserRepository {
	return m
}

func TestAuthService_RegisterSuccess(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "secret-key", 24)

	req := entity.RegisterRequest{
		Name:     "Barista John",
		Email:    "john@coffee.com",
		Password: "password123",
		Role:     entity.RoleCashier,
	}

	resp, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp.Email != req.Email {
		t.Errorf("expected email %s, got: %s", req.Email, resp.Email)
	}
	if !resp.IsActive {
		t.Errorf("expected IsActive to be true")
	}

	// Verify password was hashed
	storedUser := repo.usersByEmail[req.Email]
	if storedUser == nil {
		t.Fatal("expected user to be saved in repository")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(req.Password)); err != nil {
		t.Errorf("stored password does not match original password hash: %v", err)
	}
}

func TestAuthService_RegisterEmailAlreadyExists(t *testing.T) {
	repo := newMockUserRepo()
	repo.usersByEmail["existing@coffee.com"] = &entity.User{
		ID:    "user-1",
		Email: "existing@coffee.com",
	}

	svc := NewAuthService(repo, "secret-key", 24)

	req := entity.RegisterRequest{
		Name:     "Duplicate User",
		Email:    "existing@coffee.com",
		Password: "password123",
		Role:     entity.RoleCashier,
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	if err.Error() != "email sudah terdaftar" {
		t.Errorf("expected 'email sudah terdaftar', got: %s", err.Error())
	}
}

func TestAuthService_LoginSuccess(t *testing.T) {
	repo := newMockUserRepo()
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.usersByEmail["cashier@coffee.com"] = &entity.User{
		ID:        "user-id-555",
		Name:      "Cashier Bob",
		Email:     "cashier@coffee.com",
		Password:  string(hashedPass),
		Role:      entity.RoleCashier,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	secret := "test-jwt-secret-key-32-chars-long"
	svc := NewAuthService(repo, secret, 24)

	resp, err := svc.Login(context.Background(), entity.LoginRequest{
		Email:    "cashier@coffee.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected login success, got: %v", err)
	}

	if resp.Token == "" {
		t.Fatal("expected token string, got empty")
	}
	if resp.User.Email != "cashier@coffee.com" {
		t.Errorf("expected user email cashier@coffee.com, got: %s", resp.User.Email)
	}

	// Parse and verify token claims
	parsedToken, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !parsedToken.Valid {
		t.Fatalf("expected valid JWT token, got error: %v", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}
	if claims["user_id"] != "user-id-555" {
		t.Errorf("expected user_id user-id-555, got: %v", claims["user_id"])
	}
	if claims["role"] != entity.RoleCashier {
		t.Errorf("expected role %s, got: %v", entity.RoleCashier, claims["role"])
	}
}

func TestAuthService_LoginEmailNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "secret", 24)

	_, err := svc.Login(context.Background(), entity.LoginRequest{
		Email:    "nonexistent@coffee.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "email atau password salah" {
		t.Errorf("expected 'email atau password salah', got: %s", err.Error())
	}
}

func TestAuthService_LoginWrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	repo.usersByEmail["user@coffee.com"] = &entity.User{
		ID:       "u-1",
		Email:    "user@coffee.com",
		Password: string(hashedPass),
		IsActive: true,
	}

	svc := NewAuthService(repo, "secret", 24)

	_, err := svc.Login(context.Background(), entity.LoginRequest{
		Email:    "user@coffee.com",
		Password: "wrong-password",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "email atau password salah" {
		t.Errorf("expected 'email atau password salah', got: %s", err.Error())
	}
}

func TestAuthService_LoginInactiveUser(t *testing.T) {
	repo := newMockUserRepo()
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.usersByEmail["inactive@coffee.com"] = &entity.User{
		ID:       "u-inactive",
		Email:    "inactive@coffee.com",
		Password: string(hashedPass),
		IsActive: false,
	}

	svc := NewAuthService(repo, "secret", 24)

	_, err := svc.Login(context.Background(), entity.LoginRequest{
		Email:    "inactive@coffee.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "akun dinonaktifkan" {
		t.Errorf("expected 'akun dinonaktifkan', got: %s", err.Error())
	}
}
