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
	"github.com/modura-dev/modura/backend/internal/modules/identity"
)

type identityStub struct {
	actor identity.Actor
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
		request.AddCookie(&http.Cookie{Name: refreshCookie, Value: "refresh"})
		request.AddCookie(&http.Cookie{Name: csrfCookie, Value: strings.Repeat("c", 32)})
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

func testRouter(service Identity) http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := New(service, true, func() (string, error) { return strings.Repeat("c", 32), nil }, nil)
	generated.RegisterHandlersWithOptions(router, h, generated.GinServerOptions{BaseURL: "/api"})
	return router
}
