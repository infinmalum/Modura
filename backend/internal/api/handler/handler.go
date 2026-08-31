// Package handler composes module-owned HTTP adapters into the generated contract.
package handler

import (
	"context"

	"github.com/modura-dev/modura/backend/internal/api/generated"
	apihttp "github.com/modura-dev/modura/backend/internal/api/transport"
	identityhttp "github.com/modura-dev/modura/backend/internal/modules/identity/transport/http"
	organizationhttp "github.com/modura-dev/modura/backend/internal/modules/organization/transport/http"
	platformadminhttp "github.com/modura-dev/modura/backend/internal/modules/platformadmin/transport/http"
	platformtenanthttp "github.com/modura-dev/modura/backend/internal/modules/platformtenant/transport/http"
)

// Dependencies are the application capabilities required by HTTP delivery.
type Dependencies struct {
	Identity       identityhttp.Service
	Authorizer     organizationhttp.Authorizer
	Organization   organizationhttp.Service
	PlatformAdmin  platformadminhttp.Service
	PlatformTenant platformtenanthttp.Service
	Ready          func(context.Context) error
}

// Identity is the tenant identity API consumed by HTTP delivery.
type Identity = identityhttp.Service

// Authorizer is the permission API consumed by organization HTTP delivery.
type Authorizer = organizationhttp.Authorizer

// Organization is the organization API consumed by HTTP delivery.
type Organization = organizationhttp.Service

// PlatformAdmin is the global administrator API consumed by HTTP delivery.
type PlatformAdmin = platformadminhttp.Service

// PlatformTenant is the tenant lifecycle API consumed by HTTP delivery.
type PlatformTenant = platformtenanthttp.Service

// Handler contains no business behavior; embedding composes the operation sets.
type Handler struct {
	*identityhttp.IdentityHandler
	*organizationhttp.OrganizationHandler
	*platformadminhttp.PlatformAdminHandler
	*platformtenanthttp.PlatformTenantHandler
	*SystemHandler
}

// New composes module-owned adapters into the complete generated server interface.
func New(deps Dependencies, cookieSecure bool, newCSRF func() (string, error)) *Handler {
	security := apihttp.NewSecurity(cookieSecure, newCSRF)
	identityHandler := identityhttp.NewHandler(deps.Identity, security)
	platformAdminHandler := platformadminhttp.NewHandler(deps.PlatformAdmin, security)
	return &Handler{
		IdentityHandler:       identityHandler,
		OrganizationHandler:   organizationhttp.NewHandler(deps.Organization, deps.Authorizer, identityHandler, security),
		PlatformAdminHandler:  platformAdminHandler,
		PlatformTenantHandler: platformtenanthttp.NewHandler(deps.PlatformTenant, platformAdminHandler, security),
		SystemHandler:         newSystemHandler(deps.Ready, security),
	}
}

var _ generated.ServerInterface = (*Handler)(nil)
