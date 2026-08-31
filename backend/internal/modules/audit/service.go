package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
)

// Store persists audit-owned records inside caller-owned transactions.
type Store interface {
	Record(context.Context, pgx.Tx, Event) error
	RecordPlatform(context.Context, pgx.Tx, PlatformEvent) error
}

// RecordPlatformWrite records successful global-administrator activity in the supplied transaction.
func (s *Service) RecordPlatformWrite(ctx context.Context, tx pgx.Tx, event PlatformEvent) error {
	event.ActorID = strings.TrimSpace(event.ActorID)
	event.Action = strings.TrimSpace(event.Action)
	event.Resource = strings.TrimSpace(event.Resource)
	event.ResourceID = strings.TrimSpace(event.ResourceID)
	event.Reason = strings.TrimSpace(event.Reason)
	event.CorrelationID = strings.TrimSpace(event.CorrelationID)
	if tx == nil || event.ActorID == "" || event.Action == "" || event.Resource == "" || event.ResourceID == "" || event.Reason == "" || event.CorrelationID == "" || event.OccurredAt.IsZero() {
		return fmt.Errorf("invalid platform audit event")
	}
	if (len(event.BeforeState) > 0 && !json.Valid(event.BeforeState)) || (len(event.AfterState) > 0 && !json.Valid(event.AfterState)) {
		return fmt.Errorf("invalid platform audit state")
	}
	id, err := s.newID(event.OccurredAt)
	if err != nil {
		return fmt.Errorf("generate audit event ID: %w", err)
	}
	event.ID = id
	if err := s.store.RecordPlatform(ctx, tx, event); err != nil {
		return fmt.Errorf("record platform audit event: %w", err)
	}
	return nil
}

// Service creates validated durable audit evidence.
type Service struct {
	store   Store
	newID   func(time.Time) (string, error)
	queries QueryStore
}

// QueryStore loads immutable tenant audit projections.
type QueryStore interface {
	List(context.Context, Query) ([]Record, error)
}

// NewService constructs the audit application service.
func NewService(store Store, newID func(time.Time) (string, error)) (*Service, error) {
	if store == nil || newID == nil {
		return nil, fmt.Errorf("invalid audit service configuration")
	}
	return &Service{store: store, newID: newID}, nil
}

// EnableQueries configures the read side without changing transactional writes.
func (s *Service) EnableQueries(store QueryStore) error {
	if store == nil {
		return fmt.Errorf("invalid audit query configuration")
	}
	s.queries = store
	return nil
}

// List returns a bounded, redacted audit page for the verified tenant actor.
func (s *Service) List(ctx context.Context, tenantID identity.TenantID, action, resource string, limit, offset int) ([]Record, error) {
	action = strings.TrimSpace(action)
	resource = strings.TrimSpace(resource)
	if tenantID == "" || s.queries == nil || limit < 1 || limit > 100 || offset < 0 || len(action) > 128 || len(resource) > 128 {
		return nil, fmt.Errorf("invalid audit query")
	}
	records, err := s.queries.List(ctx, Query{TenantID: tenantID, Action: action, Resource: resource, Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	for i := range records {
		records[i].BeforeState = redactState(records[i].BeforeState)
		records[i].AfterState = redactState(records[i].AfterState)
	}
	return records, nil
}

func redactState(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return json.RawMessage(`"[redacted-invalid-state]"`)
	}
	redactValue(value)
	redacted, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`"[redacted]"`)
	}
	return redacted
}

func redactValue(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if sensitiveKey(key) {
				typed[key] = "[redacted]"
				continue
			}
			redactValue(child)
		}
	case []any:
		for _, child := range typed {
			redactValue(child)
		}
	}
}

func sensitiveKey(key string) bool {
	normalized := strings.ToLower(key)
	for _, fragment := range []string{"password", "secret", "token", "credential", "authorization", "cookie", "session"} {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}

// RecordTenantWrite records successful tenant-user activity in the supplied transaction.
func (s *Service) RecordTenantWrite(ctx context.Context, tx pgx.Tx, event Event) error {
	event.Action = strings.TrimSpace(event.Action)
	event.Resource = strings.TrimSpace(event.Resource)
	event.ResourceID = strings.TrimSpace(event.ResourceID)
	event.Reason = strings.TrimSpace(event.Reason)
	event.CorrelationID = strings.TrimSpace(event.CorrelationID)
	if tx == nil || event.ActorID == "" || event.TenantID == "" || event.Action == "" || event.Resource == "" || event.ResourceID == "" || event.Reason == "" || event.CorrelationID == "" || event.OccurredAt.IsZero() {
		return fmt.Errorf("invalid tenant audit event")
	}
	if (len(event.BeforeState) > 0 && !json.Valid(event.BeforeState)) || (len(event.AfterState) > 0 && !json.Valid(event.AfterState)) {
		return fmt.Errorf("invalid tenant audit state")
	}
	id, err := s.newID(event.OccurredAt)
	if err != nil {
		return fmt.Errorf("generate audit event ID: %w", err)
	}
	event.ID = id
	if err := s.store.Record(ctx, tx, event); err != nil {
		return fmt.Errorf("record tenant audit event: %w", err)
	}
	return nil
}
