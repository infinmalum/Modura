package authorization

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
)

// ProvisioningStore is the transaction-aware boundary used by tenant provisioning.
type ProvisioningStore interface {
	CreateReservedRole(context.Context, pgx.Tx, Role) error
	AssignRole(context.Context, pgx.Tx, identity.TenantID, identity.UserID, RoleID, time.Time) error
	Allowed(context.Context, identity.Actor, Permission) (bool, error)
}

// Authorize verifies a stable permission for a current tenant actor.
func (s *Service) Authorize(ctx context.Context, actor identity.Actor, permission Permission) error {
	if actor.TenantID == "" || actor.UserID == "" || actor.SessionID == "" || !knownPermission(permission) {
		return ErrDenied
	}
	allowed, err := s.store.Allowed(ctx, actor, permission)
	if err != nil {
		return fmt.Errorf("check authorization: %w", err)
	}
	if !allowed {
		return ErrDenied
	}
	return nil
}

func knownPermission(permission Permission) bool {
	actions := map[Action]bool{ActionRead: true, ActionCreate: true, ActionUpdate: true, ActionDelete: true}
	resources := map[Resource]bool{ResourceDepartments: true, ResourcePositions: true, ResourceUserOrganization: true}
	return actions[permission.Action] && resources[permission.Resource]
}

// Service exposes authorization-owned application operations.
type Service struct{ store ProvisioningStore }

// NewService constructs an authorization service.
func NewService(store ProvisioningStore) (*Service, error) {
	if store == nil {
		return nil, fmt.Errorf("invalid authorization service configuration")
	}
	return &Service{store: store}, nil
}

// ProvisionTenantAdministrator creates and assigns the reserved tenant-admin role.
func (s *Service) ProvisionTenantAdministrator(ctx context.Context, tx pgx.Tx, role Role, userID identity.UserID) error {
	if role.TenantID == "" || role.ID == "" || role.Code != "tenant-admin" || userID == "" || !role.Reserved {
		return fmt.Errorf("invalid tenant administrator role")
	}
	if err := s.store.CreateReservedRole(ctx, tx, role); err != nil {
		return err
	}
	if err := s.store.AssignRole(ctx, tx, role.TenantID, userID, role.ID, role.CreatedAt); err != nil {
		return err
	}
	return nil
}
