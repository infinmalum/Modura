package audit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Store persists audit-owned records inside caller-owned transactions.
type Store interface {
	Record(context.Context, pgx.Tx, Event) error
}

// Service creates validated durable audit evidence.
type Service struct {
	store Store
	newID func(time.Time) (string, error)
}

// NewService constructs the audit application service.
func NewService(store Store, newID func(time.Time) (string, error)) (*Service, error) {
	if store == nil || newID == nil {
		return nil, fmt.Errorf("invalid audit service configuration")
	}
	return &Service{store: store, newID: newID}, nil
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
