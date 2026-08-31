// Package postgres persists organization-owned data in PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
	"github.com/modura-dev/modura/backend/internal/modules/organization"
)

// Store persists tenant-scoped organization data.
type Store struct{ pool *pgxpool.Pool }

// New constructs a PostgreSQL organization store.
func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// ListDepartments returns only rows explicitly scoped to the supplied tenant.
func (s *Store) ListDepartments(ctx context.Context, tenantID identity.TenantID) ([]organization.DepartmentView, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, parent_id, name, sort_order FROM modura.departments WHERE tenant_id = $1 ORDER BY sort_order, normalized_name, id`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query departments: %w", err)
	}
	defer rows.Close()
	departments := make([]organization.DepartmentView, 0)
	for rows.Next() {
		var department organization.DepartmentView
		if err := rows.Scan(&department.ID, &department.ParentID, &department.Name, &department.SortOrder); err != nil {
			return nil, fmt.Errorf("scan department: %w", err)
		}
		departments = append(departments, department)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate departments: %w", err)
	}
	return departments, nil
}

// ListPositions returns only positions explicitly scoped to the supplied tenant.
func (s *Store) ListPositions(ctx context.Context, tenantID identity.TenantID) ([]organization.PositionView, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, status FROM modura.positions WHERE tenant_id = $1 ORDER BY normalized_name, id`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}
	defer rows.Close()
	positions := make([]organization.PositionView, 0)
	for rows.Next() {
		var position organization.PositionView
		if err := rows.Scan(&position.ID, &position.Name, &position.Status); err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		positions = append(positions, position)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate positions: %w", err)
	}
	return positions, nil
}

// CreateDepartment inserts a department and relies on composite foreign keys
// for same-tenant parent ownership.
func (s *Store) CreateDepartment(ctx context.Context, tx pgx.Tx, department organization.Department) error {
	_, err := tx.Exec(ctx, `
INSERT INTO modura.departments
    (id, tenant_id, parent_id, name, normalized_name, sort_order, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`, department.ID, department.TenantID, department.ParentID, department.Name, department.NormalizedName, department.SortOrder, department.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert department: %w", err)
	}
	return nil
}

// MoveDepartment atomically rejects root moves, cross-tenant parents, and cycles.
func (s *Store) MoveDepartment(ctx context.Context, tx pgx.Tx, tenantID identity.TenantID, departmentID, newParentID organization.DepartmentID, now time.Time) error {
	var isRoot bool
	if err := tx.QueryRow(ctx, `SELECT parent_id IS NULL FROM modura.departments WHERE tenant_id = $1 AND id = $2 FOR UPDATE`, tenantID, departmentID).Scan(&isRoot); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return organization.ErrNotFound
		}
		return fmt.Errorf("lock department: %w", err)
	}
	if isRoot {
		return organization.ErrRootDepartment
	}
	var parentExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM modura.departments WHERE tenant_id = $1 AND id = $2)`, tenantID, newParentID).Scan(&parentExists); err != nil {
		return fmt.Errorf("check department parent: %w", err)
	}
	if !parentExists {
		return organization.ErrNotFound
	}
	const cycleQuery = `
WITH RECURSIVE descendants AS (
    SELECT id FROM modura.departments WHERE tenant_id = $1 AND id = $2
    UNION ALL
    SELECT d.id FROM modura.departments d
    JOIN descendants parent ON d.parent_id = parent.id
    WHERE d.tenant_id = $1
)
SELECT EXISTS (SELECT 1 FROM descendants WHERE id = $3)`
	var cycle bool
	if err := tx.QueryRow(ctx, cycleQuery, tenantID, departmentID, newParentID).Scan(&cycle); err != nil {
		return fmt.Errorf("check department cycle: %w", err)
	}
	if cycle {
		return organization.ErrCycle
	}
	if _, err := tx.Exec(ctx, `UPDATE modura.departments SET parent_id = $3, updated_at = $4 WHERE tenant_id = $1 AND id = $2`, tenantID, departmentID, newParentID, now); err != nil {
		return fmt.Errorf("move department: %w", err)
	}
	return nil
}

// DeleteDepartment deletes an unused non-root department.
func (s *Store) DeleteDepartment(ctx context.Context, tx pgx.Tx, tenantID identity.TenantID, departmentID organization.DepartmentID) error {
	command, err := tx.Exec(ctx, `DELETE FROM modura.departments WHERE tenant_id = $1 AND id = $2 AND parent_id IS NOT NULL`, tenantID, departmentID)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23503" {
			return organization.ErrInUse
		}
		return fmt.Errorf("delete department: %w", err)
	}
	if command.RowsAffected() != 1 {
		var isRoot bool
		err := tx.QueryRow(ctx, `SELECT parent_id IS NULL FROM modura.departments WHERE tenant_id = $1 AND id = $2`, tenantID, departmentID).Scan(&isRoot)
		if errors.Is(err, pgx.ErrNoRows) {
			return organization.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("check deleted department: %w", err)
		}
		if isRoot {
			return organization.ErrRootDepartment
		}
	}
	return nil
}

// CreatePosition inserts an active tenant position.
func (s *Store) CreatePosition(ctx context.Context, tx pgx.Tx, position organization.Position) error {
	_, err := tx.Exec(ctx, `
INSERT INTO modura.positions (id, tenant_id, name, normalized_name, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, 'active', $5, $5)`, position.ID, position.TenantID, position.Name, position.NormalizedName, position.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert position: %w", err)
	}
	return nil
}

// AssignUser upserts the single primary department and optional position.
func (s *Store) AssignUser(ctx context.Context, tx pgx.Tx, tenantID identity.TenantID, userID identity.UserID, departmentID organization.DepartmentID, positionID *organization.PositionID, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO modura.user_organization
    (tenant_id, user_id, primary_department_id, position_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $5)
ON CONFLICT (tenant_id, user_id) DO UPDATE
SET primary_department_id = EXCLUDED.primary_department_id,
    position_id = EXCLUDED.position_id,
    updated_at = EXCLUDED.updated_at`, tenantID, userID, departmentID, positionID, now)
	if err != nil {
		return fmt.Errorf("assign user organization: %w", err)
	}
	return nil
}

// ProvisionInitialOrganization creates the root and first assignment in a workflow transaction.
func (s *Store) ProvisionInitialOrganization(ctx context.Context, tx pgx.Tx, root organization.Department, administratorID identity.UserID) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO modura.departments
    (id, tenant_id, parent_id, name, normalized_name, sort_order, created_at, updated_at)
VALUES ($1, $2, NULL, $3, $4, $5, $6, $6)`, root.ID, root.TenantID, root.Name, root.NormalizedName, root.SortOrder, root.CreatedAt); err != nil {
		return fmt.Errorf("insert root department: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO modura.user_organization
    (tenant_id, user_id, primary_department_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $4)`, root.TenantID, administratorID, root.ID, root.CreatedAt); err != nil {
		return fmt.Errorf("assign initial administrator organization: %w", err)
	}
	return nil
}
