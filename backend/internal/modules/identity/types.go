// Package identity owns tenants, local users, credentials, and login sessions.
package identity

import (
	"errors"
	"strings"
	"unicode"
)

var (
	// ErrInvalidCredentials deliberately hides which login input was invalid.
	ErrInvalidCredentials = errors.New("invalid tenant or credentials")
	// ErrInactiveTenant means the resolved tenant cannot establish sessions.
	ErrInactiveTenant = errors.New("tenant is not active")
	// ErrInactiveUser means the local account cannot establish sessions.
	ErrInactiveUser = errors.New("user is not active")
	// ErrInvalidToken means a token is unknown, malformed, or revoked.
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken means a valid token is outside its lifetime.
	ErrExpiredToken = errors.New("expired token")
	// ErrRefreshReuse means a previously consumed refresh token was presented.
	ErrRefreshReuse = errors.New("refresh token reuse detected")
	// ErrInvalidPassword means a proposed password violates the password policy.
	ErrInvalidPassword = errors.New("invalid password")
)

// TenantID is a verified tenant identifier.
type TenantID string

// UserID identifies a user inside a tenant.
type UserID string

// SessionID identifies a server-side authentication session.
type SessionID string

// OneTimePurpose classifies a single-use identity token.
type OneTimePurpose string

const (
	// PurposeInvitation establishes the first credential for an invited user.
	PurposeInvitation OneTimePurpose = "invitation"
	// PurposePasswordReset replaces a credential after account recovery.
	PurposePasswordReset OneTimePurpose = "password_reset"
)

// Actor is the verified tenant, user, and session scope of a request.
type Actor struct {
	TenantID  TenantID
	UserID    UserID
	SessionID SessionID
}

// NormalizeLogin canonicalizes a tenant slug, username, or email for lookup.
func NormalizeLogin(value string) string {
	return strings.Map(unicode.ToLower, strings.TrimSpace(value))
}
