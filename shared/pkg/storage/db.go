// Package storage provides database abstraction layer for all microservices.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/quality-gateway/shared/pkg/logger"
)

// DB wraps sql.DB with additional functionality.
type DB struct {
	*sql.DB
	logger *logger.Logger
}

// Config holds database configuration.
type Config struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // seconds
	ConnMaxIdleTime int // seconds
}

// Open opens a database connection.
func Open(cfg Config, log *logger.Logger) (*DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}
	if cfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Database connection established", logger.String("driver", cfg.Driver))

	return &DB{DB: db, logger: log}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.DB != nil {
		db.logger.Info("Closing database connection")
		return db.DB.Close()
	}
	return nil
}

// BeginTx begins a transaction.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, logger: db.logger}, nil
}

// Exec executes a query without returning any rows.
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db.logger.Debug("Executing query", logger.String("query", query))
	return db.DB.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows.
func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	db.logger.Debug("Querying", logger.String("query", query))
	return db.DB.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	db.logger.Debug("Querying row", logger.String("query", query))
	return db.DB.QueryRowContext(ctx, query, args...)
}

// Tx represents a database transaction.
type Tx struct {
	*sql.Tx
	logger *logger.Logger
}

// Commit commits the transaction.
func (tx *Tx) Commit() error {
	tx.logger.Debug("Committing transaction")
	return tx.Tx.Commit()
}

// Rollback aborts the transaction.
func (tx *Tx) Rollback() error {
	tx.logger.Debug("Rolling back transaction")
	return tx.Tx.Rollback()
}

// Exec executes a query within the transaction.
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.Tx.Exec(query, args...)
}

// Query executes a query within the transaction.
func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.Tx.Query(query, args...)
}

// QueryRow executes a query within the transaction.
func (tx *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	return tx.Tx.QueryRow(query, args...)
}

// Scanner is an interface for scanning database rows.
type Scanner interface {
	Scan(dest ...interface{}) error
}

// RowScanner is a helper for scanning rows into structs.
type RowScanner[T any] struct {
	scan func(Scanner) (T, error)
}

// NewRowScanner creates a new RowScanner.
func NewRowScanner[T any](scan func(Scanner) (T, error)) *RowScanner[T] {
	return &RowScanner[T]{scan: scan}
}

// Scan scans a row into a value.
func (rs *RowScanner[T]) Scan(row Scanner) (T, error) {
	return rs.scan(row)
}

// SliceScanner scans multiple rows into a slice.
func SliceScanner[T any](scanner *RowScanner[T], rows *sql.Rows) ([]T, error) {
	defer rows.Close()

	var result []T
	for rows.Next() {
		v, err := scanner.Scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Repository provides common database operations.
type Repository[T any] struct {
	db     *DB
	table  string
	logger *logger.Logger
}

// NewRepository creates a new repository.
func NewRepository[T any](db *DB, table string, log *logger.Logger) *Repository[T] {
	return &Repository[T]{
		db:     db,
		table:  table,
		logger: log,
	}
}

// GetByID retrieves an entity by ID.
func (r *Repository[T]) GetByID(ctx context.Context, id int64, scan func(Scanner) (T, error)) (T, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", r.table)
	row := r.db.QueryRow(ctx, query, id)
	return scan(row)
}

// List retrieves all entities.
func (r *Repository[T]) List(ctx context.Context, scan func(Scanner) (T, error)) ([]T, error) {
	query := fmt.Sprintf("SELECT * FROM %s", r.table)
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	scanner := NewRowScanner(scan)
	return SliceScanner(scanner, rows)
}

// Delete deletes an entity by ID.
func (r *Repository[T]) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", r.table)
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Count returns the count of entities.
func (r *Repository[T]) Count(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.table)
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// Exists checks if an entity exists by ID.
func (r *Repository[T]) Exists(ctx context.Context, id int64) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = ?)", r.table)
	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	return exists, err
}

// Inserter provides batch insert functionality.
type Inserter[T any] struct {
	db     *DB
	table  string
	logger *logger.Logger
}

// NewInserter creates a new inserter.
func NewInserter[T any](db *DB, table string, log *logger.Logger) *Inserter[T] {
	return &Inserter[T]{
		db:     db,
		table:  table,
		logger: log,
	}
}

// InsertMany inserts multiple entities in a single transaction.
func (i *Inserter[T]) InsertMany(ctx context.Context, entities []T, insertFunc func(*Tx, T) error) error {
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, entity := range entities {
		if err := insertFunc(tx, entity); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Helper function for building INSERT queries
func BuildInsertQuery(table string, columns []string) string {
	query := fmt.Sprintf("INSERT INTO %s (", table)
	for i, col := range columns {
		if i > 0 {
			query += ", "
		}
		query += col
	}
	query += ") VALUES ("
	for i := range columns {
		if i > 0 {
			query += ", "
		}
		query += "?"
	}
	query += ")"
	return query
}

// Helper function for building UPDATE queries
func BuildUpdateQuery(table string, columns []string, where string) string {
	query := fmt.Sprintf("UPDATE %s SET ", table)
	for i, col := range columns {
		if i > 0 {
			query += ", "
		}
		query += col + " = ?"
	}
	if where != "" {
		query += " WHERE " + where
	}
	return query
}
