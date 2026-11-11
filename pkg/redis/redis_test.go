package redis

import (
	"strings"
	"testing"
)

func TestNewRedis_PingFailure(t *testing.T) {
	client, err := NewRedis("127.0.0.1:54321", "")
	if err == nil {
		if client != nil {
			_ = client.Close()
		}
		t.Fatal("expected error when pinging unreachable redis server, got nil")
	}

	if !strings.Contains(err.Error(), "failed to ping redis server") {
		t.Errorf("expected descriptive ping error, got: %v", err)
	}
}
