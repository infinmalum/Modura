package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modura-dev/modura/backend/internal/api/generated"
	"github.com/modura-dev/modura/backend/internal/modules/authorization"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
	"github.com/modura-dev/modura/backend/internal/modules/organization"
	"github.com/modura-dev/modura/backend/internal/modules/platformadmin"
	"github.com/modura-dev/modura/backend/internal/modules/platformtenant"
)

type identityStub struct {
	actor identity.Actor
}

type authorizerStub struct{ denied bool }

func (s authorizerStub) Authorize(context.Context, identity.Actor, authorization.Permission) error {
	if s.denied {
		return authorization.ErrDenied
	}
	return nil
}

type organizationStub struct{}

type platformAdminStub struct{}

type platformTenantStub struct{}

func (platformAdminStub) Login(_ context.Context, username, password string) (platformadmin.Tokens, error) {
	if username != "operator" || password != "secret password" {
		return platformadmin.Tokens{}, platformadmin.ErrInvalidCredentials
	}
	return platformadmin.Tokens{AccessToken: "platform-access", RefreshToken: "platform-refresh", ExpiresIn: 5 * time.Minute, RefreshExpiresIn: 24 * time.Hour}, nil
}

func (platformAdminStub) Refresh(_ context.Context, refresh string) (platformadmin.Tokens, error) {
	if refresh != "platform-refresh" {
		return platformadmin.Tokens{}, platformadmin.ErrInvalidToken
	}
	return platformadmin.Tokens{AccessToken: "rotated-platform-access", RefreshToken: "rotated-platform-refresh", ExpiresIn: 5 * time.Minute, RefreshExpiresIn: 24 * time.Hour}, nil
}

func (platformAdminStub) AuthenticateAccess(_ context.Context, token string) (platformadmin.Actor, error) {
	if token != "platform-access" {
		return platformadmin.Actor{}, platformadmin.ErrInvalidToken
	}
	return platformadmin.Actor{AdministratorID: "018bcfe5-6800-7000-8000-000000000801", SessionID: "018bcfe5-6800-7000-8000-000000000802"}, nil
}

func (platformTenantStub) List(context.Context, platformadmin.Actor) ([]platformtenant.Tenant, error) {
	return nil, nil
}
func (platformTenantStub) Suspend(context.Context, platformadmin.Actor, identity.TenantID, string, string) error {
	return nil
}
func (platformTenantStub) Reactivate(context.Context, platformadmin.Actor, identity.TenantID, string, string) error {
	return nil
}

func (organizationStub) ListDepartments(context.Context, identity.TenantID) ([]organization.DepartmentView, error) {
	return nil, nil
}
func (organizationStub) ListPositions(context.Context, identity.TenantID) ([]organization.PositionView, error) {
	return nil, nil
}
func (organizationStub) CreateDepartment(context.Context, identity.TenantID, *organization.DepartmentID, string, int) (organization.DepartmentID, error) {
	return "018bcfe5-6800-7000-8000-000000000701", nil
}
func (organizationStub) MoveDepartment(context.Context, identity.TenantID, organization.DepartmentID, organization.DepartmentID) error {
	return nil
}
func (organizationStub) DeleteDepartment(context.Context, identity.TenantID, organization.DepartmentID) error {
	return nil
}
func (organizationStub) CreatePosition(context.Context, identity.TenantID, string) (organization.PositionID, error) {
	return "018bcfe5-6800-7000-8000-000000000702", nil
}
func (organizationStub) AssignUser(context.Context, identity.TenantID, identity.UserID, organization.DepartmentID, *organization.PositionID) error {
	return nil
}

func (s *identityStub) Login(_ context.Context, tenant, login, password string) (identity.Tokens, error) {
	if tenant != "acme" || login != "alice" || password != "secret password" {
		return identity.Tokens{}, identity.ErrInvalidCredentials
	}
	return identity.Tokens{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 5 * time.Minute, RefreshExpiresIn: 24 * time.Hour}, nil
}
func (s *identityStub) Refresh(_ context.Context, refresh string) (identity.Tokens, error) {
	if refresh != "refresh" {
		return identity.Tokens{}, identity.ErrInvalidToken
	}
	return identity.Tokens{AccessToken: "rotated-access", RefreshToken: "rotated-refresh", ExpiresIn: 5 * time.Minute, RefreshExpiresIn: 24 * time.Hour}, nil
}
func (s *identityStub) AuthenticateAccess(_ context.Context, token string) (identity.Actor, error) {
	if token != "access" {
		return identity.Actor{}, identity.ErrInvalidToken
	}
	return s.actor, nil
}
func (*identityStub) Logout(context.Context, identity.Actor) error    { return nil }
func (*identityStub) LogoutAll(context.Context, identity.Actor) error { return nil }
func (*identityStub) ChangePassword(context.Context, identity.Actor, string, string, string) (identity.Tokens, error) {
	return identity.Tokens{AccessToken: "changed-access", RefreshToken: "changed-refresh", ExpiresIn: 5 * time.Minute, RefreshExpiresIn: 24 * time.Hour}, nil
}
func (*identityStub) ConsumeOneTimeToken(_ context.Context, token string, _ identity.OneTimePurpose, password string) error {
	if token != strings.Repeat("t", 32) {
		return identity.ErrInvalidToken
	}
	if len(password) < 12 {
		return identity.ErrInvalidPassword
	}
	return nil
}

func TestLoginSetsProtectedSessionCookies(t *testing.T) {
	router := testRouter(&identityStub{})
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/auth/login", strings.NewReader(`{"tenant":"acme","login":"alice","password":"secret password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 2 || !cookies[0].HttpOnly || cookies[1].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatalf("unexpected cookies: %+v", cookies)
	}
	var body generated.AccessTokenResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.AccessToken != "access" || body.CsrfToken != strings.Repeat("c", 32) {
		t.Fatalf("body = %+v", body)
	}
}

func TestRefreshRequiresMatchingCSRF(t *testing.T) {
	router := testRouter(&identityStub{})
	for _, header := range []string{"", strings.Repeat("x", 32)} {
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/auth/refresh", nil)
		request.AddCookie(&http.Cookie{Name: "modura_refresh", Value: "refresh"})
		request.AddCookie(&http.Cookie{Name: "modura_csrf", Value: strings.Repeat("c", 32)})
		if header != "" {
			request.Header.Set("X-CSRF-Token", header)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		want := http.StatusBadRequest
		if header != "" {
			want = http.StatusForbidden
		}
		if response.Code != want {
			t.Fatalf("header %q status = %d, want %d", header, response.Code, want)
		}
	}
}

func TestPlatformLoginUsesDistinctCookies(t *testing.T) {
	router := testRouter(&identityStub{})
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/platform/auth/login", strings.NewReader(`{"username":"operator","password":"secret password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 2 || cookies[0].Name != "modura_platform_refresh" || cookies[1].Name != "modura_platform_csrf" {
		t.Fatalf("unexpected platform cookies: %+v", cookies)
	}
	for _, cookie := range cookies {
		if cookie.Path != "/api/platform/auth" {
			t.Fatalf("cookie %s path = %q", cookie.Name, cookie.Path)
		}
	}
}

func TestPlatformRefreshRejectsTenantCookies(t *testing.T) {
	router := testRouter(&identityStub{})
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/platform/auth/refresh", nil)
	request.AddCookie(&http.Cookie{Name: "modura_refresh", Value: "refresh"})
	request.AddCookie(&http.Cookie{Name: "modura_csrf", Value: strings.Repeat("c", 32)})
	request.Header.Set("X-CSRF-Token", strings.Repeat("c", 32))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestOneTimeCredentialResponseDoesNotDiscloseTokenState(t *testing.T) {
	router := testRouter(&identityStub{})
	for _, token := range []string{strings.Repeat("x", 32), strings.Repeat("t", 32)} {
		body := `{"token":"` + token + `","newPassword":"a secure new password"}`
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/auth/password-resets", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		want := http.StatusUnauthorized
		if token == strings.Repeat("t", 32) {
			want = http.StatusNoContent
		}
		if response.Code != want {
			t.Fatalf("status = %d, want %d, body=%s", response.Code, want, response.Body.String())
		}
	}
}

func TestDepartmentManagementRequiresServerAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	service := &identityStub{actor: identity.Actor{TenantID: "tenant", UserID: "user", SessionID: "session"}}
	h := New(Dependencies{Identity: service, Authorizer: authorizerStub{denied: true}, Organization: organizationStub{}, PlatformAdmin: platformAdminStub{}, PlatformTenant: platformTenantStub{}}, true, func() (string, error) { return strings.Repeat("c", 32), nil })
	generated.RegisterHandlersWithOptions(router, h, generated.GinServerOptions{BaseURL: "/api"})
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/organization/departments", nil)
	request.Header.Set("Authorization", "Bearer access")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body=%s", response.Code, http.StatusForbidden, response.Body.String())
	}
}

func testRouter(service Identity) http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := New(Dependencies{Identity: service, Authorizer: authorizerStub{}, Organization: organizationStub{}, PlatformAdmin: platformAdminStub{}, PlatformTenant: platformTenantStub{}}, true, func() (string, error) { return strings.Repeat("c", 32), nil })
	generated.RegisterHandlersWithOptions(router, h, generated.GinServerOptions{BaseURL: "/api"})
	return router
}
