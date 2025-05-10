package database

import (
	"database/sql"
)

type TransactionManager interface {
	Begin() (*sql.Tx, error)
	
	GetDB() *sql.DB
}

type PostgresTransactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) TransactionManager {
	return &PostgresTransactionManager{db: db}
}

func (tm *PostgresTransactionManager) Begin() (*sql.Tx, error) {
	return tm.db.Begin()
}

func (tm *PostgresTransactionManager) GetDB() *sql.DB {
	return tm.db
}