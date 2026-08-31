// Package postgres persists authorization-owned data in PostgreSQL.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modura-dev/modura/backend/internal/modules/authorization"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
)

// Store persists authorization data inside caller-owned workflow transactions.
type Store struct{ pool *pgxpool.Pool }

// New constructs an authorization store.
func New(pool *pgxpool.Pool) Store { return Store{pool: pool} }

// CreateReservedRole inserts a reserved role using the supplied workflow transaction.
func (Store) CreateReservedRole(ctx context.Context, tx pgx.Tx, role authorization.Role) error {
	_, err := tx.Exec(ctx, `
INSERT INTO modura.roles (id, tenant_id, code, name, reserved, created_at, updated_at)
VALUES ($1, $2, $3, $4, true, $5, $5)`, role.ID, role.TenantID, role.Code, role.Name, role.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert reserved role: %w", err)
	}
	return nil
}

// AssignRole grants a tenant role to a tenant-local user in the workflow transaction.
func (Store) AssignRole(ctx context.Context, tx pgx.Tx, tenantID identity.TenantID, userID identity.UserID, roleID authorization.RoleID, now time.Time) error {
	_, err := tx.Exec(ctx, `INSERT INTO modura.user_roles (tenant_id, user_id, role_id, created_at) VALUES ($1, $2, $3, $4)`, tenantID, userID, roleID, now)
	if err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
	return nil
}

// Allowed checks tenant-local grants for a stable permission. The reserved
// tenant-admin role receives only the explicit Stage 2 management registry.
func (s Store) Allowed(ctx context.Context, actor identity.Actor, permission authorization.Permission) (bool, error) {
	if s.pool == nil {
		return false, nil
	}
	const query = `
SELECT EXISTS (
    SELECT 1
    FROM modura.user_roles ur
    JOIN modura.roles r ON r.tenant_id = ur.tenant_id AND r.id = ur.role_id
    WHERE ur.tenant_id = $1 AND ur.user_id = $2
      AND r.code = 'tenant-admin' AND r.reserved = true
)`
	var tenantAdministrator bool
	if err := s.pool.QueryRow(ctx, query, actor.TenantID, actor.UserID).Scan(&tenantAdministrator); err != nil {
		return false, fmt.Errorf("query role grants: %w", err)
	}
	if !tenantAdministrator {
		return false, nil
	}
	return tenantAdminPermission(permission), nil
}

func tenantAdminPermission(permission authorization.Permission) bool {
	switch permission.Resource {
	case authorization.ResourceDepartments:
		return permission.Action == authorization.ActionRead || permission.Action == authorization.ActionCreate || permission.Action == authorization.ActionUpdate || permission.Action == authorization.ActionDelete
	case authorization.ResourcePositions, authorization.ResourceUserOrganization:
		return permission.Action == authorization.ActionRead || permission.Action == authorization.ActionCreate || permission.Action == authorization.ActionUpdate
	default:
		return false
	}
}
