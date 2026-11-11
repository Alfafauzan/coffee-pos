package config

import (
	"os"
	"strings"
	"testing"
)

func setValidEnv() {
	os.Setenv("DB_HOST", "127.0.0.1")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_NAME", "coffee_pos")
	os.Setenv("DB_USER", "root")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("JWT_SECRET", "super-secret-jwt-key-minimum-32-chars")
}

func clearTestEnv() {
	os.Unsetenv("APP_PORT")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_EXPIRY_HOURS")
	os.Unsetenv("MIDTRANS_SERVER_KEY")
	os.Unsetenv("MIDTRANS_CLIENT_KEY")
	os.Unsetenv("MIDTRANS_ENV")
}

func TestLoadSuccess(t *testing.T) {
	clearTestEnv()
	setValidEnv()
	defer clearTestEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.DBHost != "127.0.0.1" {
		t.Errorf("expected DBHost 127.0.0.1, got: %s", cfg.DBHost)
	}
	if cfg.JWTExpiryHours != 24 {
		t.Errorf("expected default JWTExpiryHours 24, got: %d", cfg.JWTExpiryHours)
	}

	// Test MysqlDSN
	expectedDSN := "root:secret@tcp(127.0.0.1:3306)/coffee_pos?parseTime=true&loc=Local"
	if dsn := cfg.MysqlDSN(); dsn != expectedDSN {
		t.Errorf("expected DSN %s, got %s", expectedDSN, dsn)
	}

	// Test RedisAddr
	expectedRedis := "localhost:6379"
	if addr := cfg.RedisAddr(); addr != expectedRedis {
		t.Errorf("expected RedisAddr %s, got %s", expectedRedis, addr)
	}
}

func TestLoadMissingRequiredFields(t *testing.T) {
	clearTestEnv()
	defer clearTestEnv()

	// Missing DB_HOST
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when required fields are missing, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "wajib diisi") {
		t.Errorf("expected error message to contain 'wajib diisi', got: %s", errMsg)
	}
}

func TestCustomJWTExpiry(t *testing.T) {
	clearTestEnv()
	setValidEnv()
	os.Setenv("JWT_EXPIRY_HOURS", "48")
	defer clearTestEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.JWTExpiryHours != 48 {
		t.Errorf("expected JWTExpiryHours 48, got: %d", cfg.JWTExpiryHours)
	}
}
