package txmanager

import (
	"context"
	"database/sql"
)

// TxManager manages database transaction execution and lifecycle.
type TxManager struct {
	db *sql.DB
}

// New creates a new TxManager instance with the provided database connection.
func New(db *sql.DB) *TxManager {
	return &TxManager{
		db: db,
	}
}

// WithTx executes the given function fn inside an atomic database transaction.
// If fn returns an error or panics, the transaction will be rolled back.
// If fn succeeds without error, the transaction will be committed.
func (tm *TxManager) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
