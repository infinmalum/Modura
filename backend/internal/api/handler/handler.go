// Package handler maps the authoritative HTTP contract to application APIs.
package handler

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modura-dev/modura/backend/internal/api/generated"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
)

const (
	refreshCookie = "modura_refresh"
	csrfCookie    = "modura_csrf"
)

// Identity defines the authentication use cases consumed by HTTP delivery.
type Identity interface {
	Login(context.Context, string, string, string) (identity.Tokens, error)
	Refresh(context.Context, string) (identity.Tokens, error)
	AuthenticateAccess(context.Context, string) (identity.Actor, error)
	Logout(context.Context, identity.Actor) error
	LogoutAll(context.Context, identity.Actor) error
	ChangePassword(context.Context, identity.Actor, string, string, string) (identity.Tokens, error)
	ConsumeOneTimeToken(context.Context, string, identity.OneTimePurpose, string) error
}

// Handler implements the generated HTTP server contract.
type Handler struct {
	identity     Identity
	cookieSecure bool
	newCSRF      func() (string, error)
	ready        func(context.Context) error
}

// New constructs the HTTP contract handler.
func New(identityService Identity, cookieSecure bool, newCSRF func() (string, error), ready func(context.Context) error) *Handler {
	return &Handler{identity: identityService, cookieSecure: cookieSecure, newCSRF: newCSRF, ready: ready}
}

// GetLiveness reports process liveness.
func (h *Handler) GetLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, generated.HealthStatus{Status: generated.Ok})
}

// GetReadiness reports whether configured dependencies are available.
func (h *Handler) GetReadiness(c *gin.Context) {
	if h.ready != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()
		if err := h.ready(ctx); err != nil {
			h.problem(c, http.StatusServiceUnavailable, "service unavailable")
			return
		}
	}
	c.JSON(http.StatusOK, generated.HealthStatus{Status: generated.Ok})
}

// Login establishes a tenant-local session.
func (h *Handler) Login(c *gin.Context) {
	if h.identity == nil {
		h.problem(c, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var request generated.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.Password == nil {
		h.problem(c, http.StatusBadRequest, "invalid request")
		return
	}
	tokens, err := h.identity.Login(c.Request.Context(), request.Tenant, request.Login, *request.Password)
	if err != nil {
		h.problem(c, http.StatusUnauthorized, "authentication failed")
		return
	}
	h.writeTokens(c, tokens)
}

// Refresh rotates the refresh cookie after double-submit CSRF validation.
func (h *Handler) Refresh(c *gin.Context, params generated.RefreshParams) {
	refresh, ok := h.cookieAndCSRF(c, params.XCSRFToken)
	if !ok {
		return
	}
	tokens, err := h.identity.Refresh(c.Request.Context(), refresh)
	if err != nil {
		h.clearCookies(c)
		h.problem(c, http.StatusUnauthorized, "authentication failed")
		return
	}
	h.writeTokens(c, tokens)
}

// Logout revokes the current authenticated session.
func (h *Handler) Logout(c *gin.Context, params generated.LogoutParams) {
	if _, ok := h.cookieAndCSRF(c, params.XCSRFToken); !ok {
		return
	}
	actor, ok := h.bearerActor(c)
	if !ok {
		return
	}
	if err := h.identity.Logout(c.Request.Context(), actor); err != nil {
		h.problem(c, http.StatusUnauthorized, "authentication failed")
		return
	}
	h.clearCookies(c)
	c.Status(http.StatusNoContent)
}

// LogoutAll revokes every session for the authenticated tenant-local user.
func (h *Handler) LogoutAll(c *gin.Context, params generated.LogoutAllParams) {
	if _, ok := h.cookieAndCSRF(c, params.XCSRFToken); !ok {
		return
	}
	actor, ok := h.bearerActor(c)
	if !ok {
		return
	}
	if err := h.identity.LogoutAll(c.Request.Context(), actor); err != nil {
		h.problem(c, http.StatusUnauthorized, "authentication failed")
		return
	}
	h.clearCookies(c)
	c.Status(http.StatusNoContent)
}

// ChangePassword updates credentials and rotates the current authenticated session.
func (h *Handler) ChangePassword(c *gin.Context, params generated.ChangePasswordParams) {
	refresh, ok := h.cookieAndCSRF(c, params.XCSRFToken)
	if !ok {
		return
	}
	actor, ok := h.bearerActor(c)
	if !ok {
		return
	}
	var request generated.ChangePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.CurrentPassword == nil || request.NewPassword == nil {
		h.problem(c, http.StatusBadRequest, "invalid request")
		return
	}
	tokens, err := h.identity.ChangePassword(c.Request.Context(), actor, *request.CurrentPassword, *request.NewPassword, refresh)
	if err != nil {
		h.problem(c, http.StatusUnauthorized, "authentication failed")
		return
	}
	h.writeTokens(c, tokens)
}

// AcceptInvitation consumes a single-use invitation token.
func (h *Handler) AcceptInvitation(c *gin.Context) {
	h.consumeOneTimeCredential(c, identity.PurposeInvitation)
}

// ResetPassword consumes a single-use password-recovery token.
func (h *Handler) ResetPassword(c *gin.Context) {
	h.consumeOneTimeCredential(c, identity.PurposePasswordReset)
}

func (h *Handler) consumeOneTimeCredential(c *gin.Context, purpose identity.OneTimePurpose) {
	if h.identity == nil {
		h.problem(c, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var request generated.OneTimeCredentialRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.Token == nil || request.NewPassword == nil {
		h.problem(c, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.identity.ConsumeOneTimeToken(c.Request.Context(), *request.Token, purpose, *request.NewPassword); err != nil {
		if errors.Is(err, identity.ErrInvalidPassword) {
			h.problem(c, http.StatusBadRequest, "invalid request")
			return
		}
		h.problem(c, http.StatusUnauthorized, "authentication failed")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) bearerActor(c *gin.Context) (identity.Actor, bool) {
	prefix, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
	if !ok || !strings.EqualFold(prefix, "Bearer") || token == "" {
		h.problem(c, http.StatusUnauthorized, "authentication failed")
		return identity.Actor{}, false
	}
	actor, err := h.identity.AuthenticateAccess(c.Request.Context(), token)
	if err != nil {
		h.problem(c, http.StatusUnauthorized, "authentication failed")
		return identity.Actor{}, false
	}
	return actor, true
}

func (h *Handler) cookieAndCSRF(c *gin.Context, header string) (string, bool) {
	refresh, refreshErr := c.Cookie(refreshCookie)
	csrf, csrfErr := c.Cookie(csrfCookie)
	if refreshErr != nil || csrfErr != nil || header == "" || subtle.ConstantTimeCompare([]byte(header), []byte(csrf)) != 1 {
		h.problem(c, http.StatusForbidden, "request verification failed")
		return "", false
	}
	return refresh, true
}

func (h *Handler) writeTokens(c *gin.Context, tokens identity.Tokens) {
	csrf, err := h.newCSRF()
	if err != nil {
		h.problem(c, http.StatusInternalServerError, "internal server error")
		return
	}
	maxAge := int(tokens.RefreshExpiresIn.Seconds())
	h.setCookie(c, refreshCookie, tokens.RefreshToken, maxAge, true)
	h.setCookie(c, csrfCookie, csrf, maxAge, false)
	c.JSON(http.StatusOK, generated.AccessTokenResponse{AccessToken: tokens.AccessToken, TokenType: generated.Bearer, ExpiresIn: int64(tokens.ExpiresIn.Seconds()), CsrfToken: csrf})
}

func (h *Handler) clearCookies(c *gin.Context) {
	h.setCookie(c, refreshCookie, "", -1, true)
	h.setCookie(c, csrfCookie, "", -1, false)
}

func (h *Handler) setCookie(c *gin.Context, name, value string, maxAge int, httpOnly bool) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(name, value, maxAge, "/api/auth", "", h.cookieSecure, httpOnly)
}

func (h *Handler) problem(c *gin.Context, status int, title string) {
	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, generated.Problem{Type: "about:blank", Title: title, Status: status})
}

var _ generated.ServerInterface = (*Handler)(nil)
