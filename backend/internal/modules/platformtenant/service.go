// Package platformtenant owns platform-level tenant lifecycle use cases.
package platformtenant

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modura-dev/modura/backend/internal/modules/identity"
	"github.com/modura-dev/modura/backend/internal/modules/platformadmin"
)

var (
	// ErrNotFound means the target tenant does not exist.
	ErrNotFound = errors.New("tenant not found")
	// ErrInvalidTransition means the requested lifecycle transition is not allowed.
	ErrInvalidTransition = errors.New("invalid tenant lifecycle transition")
)

// Tenant is the platform-visible tenant summary.
type Tenant struct {
	ID          identity.TenantID
	Slug        string
	DisplayName string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// LifecycleChange carries mandatory cross-tenant authorization evidence.
type LifecycleChange struct {
	Actor         platformadmin.Actor
	TenantID      identity.TenantID
	Reason        string
	CorrelationID string
	AuditID       string
	OccurredAt    time.Time
}

// Store is the persistence boundary consumed by platform tenant use cases.
type Store interface {
	List(context.Context) ([]Tenant, error)
	ChangeStatus(context.Context, LifecycleChange, string, string) error
}

// Service implements platform tenant queries and lifecycle changes.
type Service struct {
	store Store
	now   func() time.Time
	newID func(time.Time) (string, error)
}

// NewService constructs a platform tenant service.
func NewService(store Store, now func() time.Time, newID func(time.Time) (string, error)) (*Service, error) {
	if store == nil || now == nil || newID == nil {
		return nil, fmt.Errorf("invalid platform tenant service configuration")
	}
	return &Service{store: store, now: now, newID: newID}, nil
}

// List returns all tenants to an already verified platform actor.
func (s *Service) List(ctx context.Context, actor platformadmin.Actor) ([]Tenant, error) {
	if actor.AdministratorID == "" || actor.SessionID == "" {
		return nil, platformadmin.ErrInvalidToken
	}
	return s.store.List(ctx)
}

// Suspend prevents tenant-local authentication and session validation.
func (s *Service) Suspend(ctx context.Context, actor platformadmin.Actor, tenantID identity.TenantID, reason, correlationID string) error {
	return s.changeStatus(ctx, actor, tenantID, reason, correlationID, "active", "suspended")
}

// Reactivate restores a suspended tenant.
func (s *Service) Reactivate(ctx context.Context, actor platformadmin.Actor, tenantID identity.TenantID, reason, correlationID string) error {
	return s.changeStatus(ctx, actor, tenantID, reason, correlationID, "suspended", "active")
}

func (s *Service) changeStatus(ctx context.Context, actor platformadmin.Actor, tenantID identity.TenantID, reason, correlationID, from, to string) error {
	reason = strings.TrimSpace(reason)
	correlationID = strings.TrimSpace(correlationID)
	if actor.AdministratorID == "" || actor.SessionID == "" || tenantID == "" || reason == "" || correlationID == "" {
		return fmt.Errorf("invalid platform tenant lifecycle request")
	}
	now := s.now().UTC()
	auditID, err := s.newID(now)
	if err != nil {
		return fmt.Errorf("generate audit event ID: %w", err)
	}
	change := LifecycleChange{Actor: actor, TenantID: tenantID, Reason: reason, CorrelationID: correlationID, AuditID: auditID, OccurredAt: now}
	if err := s.store.ChangeStatus(ctx, change, from, to); err != nil {
		return fmt.Errorf("change tenant status: %w", err)
	}
	return nil
}
