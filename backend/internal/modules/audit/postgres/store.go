// Package postgres persists audit-owned records in PostgreSQL.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/modura-dev/modura/backend/internal/modules/audit"
)

// Store persists immutable audit evidence.
type Store struct{}

// New constructs an audit store.
func New() Store { return Store{} }

// Record inserts a successful tenant-user audit event in the caller transaction.
func (Store) Record(ctx context.Context, tx pgx.Tx, event audit.Event) error {
	_, err := tx.Exec(ctx, `INSERT INTO modura.audit_events (id, actor_type, actor_id, tenant_id, action, resource, resource_id, reason, result, correlation_id, occurred_at) VALUES ($1, 'tenant_user', $2, $3, $4, $5, $6, $7, 'succeeded', $8, $9)`, event.ID, event.ActorID, event.TenantID, event.Action, event.Resource, event.ResourceID, event.Reason, event.CorrelationID, event.OccurredAt)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}
