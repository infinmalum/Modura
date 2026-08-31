// Package database provides shared PostgreSQL infrastructure.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Transactor runs an application-owned unit of work in one PostgreSQL transaction.
type Transactor struct{ pool *pgxpool.Pool }

// NewTransactor constructs a transaction runner.
func NewTransactor(pool *pgxpool.Pool) Transactor { return Transactor{pool: pool} }

// WithinTransaction commits only when work succeeds and otherwise rolls back.
func (t Transactor) WithinTransaction(ctx context.Context, work func(pgx.Tx) error) error {
	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := work(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
