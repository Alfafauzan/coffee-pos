package database

import (
	"strings"
	"testing"
)

func TestNewMySQL_PingFailure(t *testing.T) {
	// DSN pointing to an unreachable port/host
	invalidDSN := "invalid_user:invalid_pass@tcp(127.0.0.1:54321)/non_existent_db?timeout=1s"

	db, err := NewMySQL(invalidDSN)
	if err == nil {
		if db != nil {
			_ = db.Close()
		}
		t.Fatal("expected error when pinging unreachable mysql server, got nil")
	}

	if !strings.Contains(err.Error(), "failed to ping mysql database") {
		t.Errorf("expected descriptive ping error message, got: %v", err)
	}
}
