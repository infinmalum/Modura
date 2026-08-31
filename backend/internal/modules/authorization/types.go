// Package authorization owns tenant roles, grants, and policy evaluation.
package authorization

import (
	"errors"
	"time"

	"github.com/modura-dev/modura/backend/internal/modules/identity"
)

var (
	// ErrDenied means the actor lacks the requested stable permission.
	ErrDenied = errors.New("authorization denied")
)

// Resource is a stable authorization resource identifier.
type Resource string

// Action is a stable authorization action identifier.
type Action string

const (
	// ResourceDepartments protects department management.
	ResourceDepartments Resource = "organization.departments"
	// ResourcePositions protects position management.
	ResourcePositions Resource = "organization.positions"
	// ResourceUserOrganization protects user organization assignments.
	ResourceUserOrganization Resource = "organization.user-organization"
	// ActionRead permits viewing a resource.
	ActionRead Action = "read"
	// ActionCreate permits creating a resource.
	ActionCreate Action = "create"
	// ActionUpdate permits changing a resource.
	ActionUpdate Action = "update"
	// ActionDelete permits deleting a resource.
	ActionDelete Action = "delete"
)

// Permission is the stable resource/action pair checked by the server.
type Permission struct {
	Resource Resource
	Action   Action
}

// RoleID identifies a tenant-owned role.
type RoleID string

// Role is the state required to provision a role.
type Role struct {
	ID        RoleID
	TenantID  identity.TenantID
	Code      string
	Name      string
	Reserved  bool
	CreatedAt time.Time
}
