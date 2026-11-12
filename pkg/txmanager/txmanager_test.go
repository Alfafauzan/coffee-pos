package txmanager

import (
	"database/sql"
	"testing"
)

func TestNew(t *testing.T) {
	dummyDB := &sql.DB{}
	tm := New(dummyDB)

	if tm == nil {
		t.Fatal("expected non-nil TxManager instance, got nil")
	}

	if tm.db != dummyDB {
		t.Errorf("expected db pointer to match, got %v, want %v", tm.db, dummyDB)
	}
}
