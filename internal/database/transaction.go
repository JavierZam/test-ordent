// internal/database/transaction.go
package database

import (
	"database/sql"
)

// TransactionManager manages database transactions
type TransactionManager interface {
	// Begin starts a new transaction
	Begin() (*sql.Tx, error)
	
	// GetDB returns the database connection
	GetDB() *sql.DB
}

// PostgresTransactionManager implements TransactionManager
type PostgresTransactionManager struct {
	db *sql.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sql.DB) TransactionManager {
	return &PostgresTransactionManager{db: db}
}

// Begin starts a new transaction
func (tm *PostgresTransactionManager) Begin() (*sql.Tx, error) {
	return tm.db.Begin()
}

// GetDB returns the database connection
func (tm *PostgresTransactionManager) GetDB() *sql.DB {
	return tm.db
}