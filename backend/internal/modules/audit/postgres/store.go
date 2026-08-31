// Package postgres persists audit-owned records in PostgreSQL.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modura-dev/modura/backend/internal/modules/audit"
)

// Store persists immutable audit evidence.
type Store struct{ pool *pgxpool.Pool }

// New constructs an audit store.
func New(pools ...*pgxpool.Pool) Store {
	var pool *pgxpool.Pool
	if len(pools) > 0 {
		pool = pools[0]
	}
	return Store{pool: pool}
}

// Record inserts a successful tenant-user audit event in the caller transaction.
func (Store) Record(ctx context.Context, tx pgx.Tx, event audit.Event) error {
	_, err := tx.Exec(ctx, `INSERT INTO modura.audit_events (id, actor_type, actor_id, tenant_id, action, resource, resource_id, reason, result, correlation_id, occurred_at, before_state, after_state) VALUES ($1, 'tenant_user', $2, $3, $4, $5, $6, $7, 'succeeded', $8, $9, $10, $11)`, event.ID, event.ActorID, event.TenantID, event.Action, event.Resource, event.ResourceID, event.Reason, event.CorrelationID, event.OccurredAt, nullableJSON(event.BeforeState), nullableJSON(event.AfterState))
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

// RecordPlatform inserts a successful global-administrator audit event.
func (Store) RecordPlatform(ctx context.Context, tx pgx.Tx, event audit.PlatformEvent) error {
	_, err := tx.Exec(ctx, `INSERT INTO modura.audit_events (id, actor_type, actor_id, tenant_id, action, resource, resource_id, reason, result, correlation_id, occurred_at, before_state, after_state) VALUES ($1, 'platform_administrator', $2, NULL, $3, $4, $5, $6, 'succeeded', $7, $8, $9, $10)`, event.ID, event.ActorID, event.Action, event.Resource, event.ResourceID, event.Reason, event.CorrelationID, event.OccurredAt, nullableJSON(event.BeforeState), nullableJSON(event.AfterState))
	if err != nil {
		return fmt.Errorf("insert platform audit event: %w", err)
	}
	return nil
}

// List returns only immutable events in the explicitly supplied tenant.
func (s Store) List(ctx context.Context, query audit.Query) ([]audit.Record, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("audit query store is unavailable")
	}
	rows, err := s.pool.Query(ctx, `SELECT id, actor_type, actor_id, tenant_id, action, resource, resource_id, reason, result, correlation_id, occurred_at, before_state, after_state FROM modura.audit_events WHERE tenant_id = $1 AND ($2 = '' OR action = $2) AND ($3 = '' OR resource = $3) ORDER BY occurred_at DESC, id DESC LIMIT $4 OFFSET $5`, query.TenantID, query.Action, query.Resource, query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()
	records := make([]audit.Record, 0)
	for rows.Next() {
		var record audit.Record
		if err := rows.Scan(&record.ID, &record.ActorType, &record.ActorID, &record.TenantID, &record.Action, &record.Resource, &record.ResourceID, &record.Reason, &record.Result, &record.CorrelationID, &record.OccurredAt, &record.BeforeState, &record.AfterState); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func nullableJSON(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}
