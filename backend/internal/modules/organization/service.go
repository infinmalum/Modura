package organization

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
)

// Store is the persistence boundary consumed by organization use cases.
type Store interface {
	ListDepartments(context.Context, identity.TenantID) ([]DepartmentView, error)
	ListPositions(context.Context, identity.TenantID) ([]PositionView, error)
	CreateDepartment(context.Context, Department) error
	MoveDepartment(context.Context, identity.TenantID, DepartmentID, DepartmentID, time.Time) error
	DeleteDepartment(context.Context, identity.TenantID, DepartmentID) error
	CreatePosition(context.Context, Position) error
	AssignUser(context.Context, identity.TenantID, identity.UserID, DepartmentID, *PositionID, time.Time) error
	ProvisionInitialOrganization(context.Context, pgx.Tx, Department, identity.UserID) error
}

// ListPositions returns the actor tenant's position catalog.
func (s *Service) ListPositions(ctx context.Context, tenantID identity.TenantID) ([]PositionView, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("invalid tenant scope")
	}
	positions, err := s.store.ListPositions(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list positions: %w", err)
	}
	return positions, nil
}

// ListDepartments returns the actor tenant's department projection.
func (s *Service) ListDepartments(ctx context.Context, tenantID identity.TenantID) ([]DepartmentView, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("invalid tenant scope")
	}
	departments, err := s.store.ListDepartments(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	return departments, nil
}

// Service implements tenant-scoped organization use cases.
type Service struct {
	store Store
	now   func() time.Time
	newID func(time.Time) (string, error)
}

// NewService constructs an organization service.
func NewService(store Store, now func() time.Time, newID func(time.Time) (string, error)) (*Service, error) {
	if store == nil || now == nil || newID == nil {
		return nil, fmt.Errorf("invalid organization service configuration")
	}
	return &Service{store: store, now: now, newID: newID}, nil
}

// CreateDepartment creates a root or child department in an explicit tenant.
func (s *Service) CreateDepartment(ctx context.Context, tenantID identity.TenantID, parentID *DepartmentID, name string, sortOrder int) (DepartmentID, error) {
	normalized := NormalizeName(name)
	if tenantID == "" || normalized == "" {
		return "", fmt.Errorf("invalid department")
	}
	now := s.now().UTC()
	id, err := s.newID(now)
	if err != nil {
		return "", fmt.Errorf("generate department ID: %w", err)
	}
	department := Department{ID: DepartmentID(id), TenantID: tenantID, ParentID: parentID, Name: strings.TrimSpace(name), NormalizedName: normalized, SortOrder: sortOrder, CreatedAt: now}
	if err := s.store.CreateDepartment(ctx, department); err != nil {
		return "", fmt.Errorf("create department: %w", err)
	}
	return department.ID, nil
}

// MoveDepartment moves a non-root department while preventing cycles.
func (s *Service) MoveDepartment(ctx context.Context, tenantID identity.TenantID, departmentID, newParentID DepartmentID) error {
	if tenantID == "" || departmentID == "" || newParentID == "" {
		return fmt.Errorf("invalid department move")
	}
	return s.store.MoveDepartment(ctx, tenantID, departmentID, newParentID, s.now().UTC())
}

// DeleteDepartment deletes an unused non-root department.
func (s *Service) DeleteDepartment(ctx context.Context, tenantID identity.TenantID, departmentID DepartmentID) error {
	if tenantID == "" || departmentID == "" {
		return fmt.Errorf("invalid department delete")
	}
	return s.store.DeleteDepartment(ctx, tenantID, departmentID)
}

// CreatePosition creates an active tenant position.
func (s *Service) CreatePosition(ctx context.Context, tenantID identity.TenantID, name string) (PositionID, error) {
	normalized := NormalizeName(name)
	if tenantID == "" || normalized == "" {
		return "", fmt.Errorf("invalid position")
	}
	now := s.now().UTC()
	id, err := s.newID(now)
	if err != nil {
		return "", fmt.Errorf("generate position ID: %w", err)
	}
	position := Position{ID: PositionID(id), TenantID: tenantID, Name: strings.TrimSpace(name), NormalizedName: normalized, CreatedAt: now}
	if err := s.store.CreatePosition(ctx, position); err != nil {
		return "", fmt.Errorf("create position: %w", err)
	}
	return position.ID, nil
}

// AssignUser sets the user's single primary department and optional position.
func (s *Service) AssignUser(ctx context.Context, tenantID identity.TenantID, userID identity.UserID, departmentID DepartmentID, positionID *PositionID) error {
	if tenantID == "" || userID == "" || departmentID == "" {
		return fmt.Errorf("invalid user organization assignment")
	}
	return s.store.AssignUser(ctx, tenantID, userID, departmentID, positionID, s.now().UTC())
}

// ProvisionInitialOrganization creates the sole root department and assigns
// the first administrator inside the caller-owned provisioning transaction.
func (s *Service) ProvisionInitialOrganization(ctx context.Context, tx pgx.Tx, root Department, administratorID identity.UserID) error {
	if tx == nil || root.TenantID == "" || root.ID == "" || root.ParentID != nil || administratorID == "" {
		return fmt.Errorf("invalid initial organization")
	}
	if err := s.store.ProvisionInitialOrganization(ctx, tx, root, administratorID); err != nil {
		return fmt.Errorf("provision initial organization: %w", err)
	}
	return nil
}
