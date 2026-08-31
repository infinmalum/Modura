package identity

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AccessTokenClaims contains the minimal stable authentication context.
type AccessTokenClaims struct {
	Issuer          string    `json:"iss"`
	Audience        string    `json:"aud"`
	PrincipalType   string    `json:"pty"`
	Subject         UserID    `json:"sub"`
	TenantID        TenantID  `json:"tid,omitempty"`
	SessionID       SessionID `json:"sid"`
	TokenID         string    `json:"jti"`
	SecurityVersion int64     `json:"sv"`
	IssuedAt        int64     `json:"iat"`
	ExpiresAt       int64     `json:"exp"`
}

// AccessTokenSigner creates short-lived HMAC-signed access tokens.
type AccessTokenSigner struct {
	issuer   string
	audience string
	keyID    string
	key      []byte
	lifetime time.Duration
}

// NewAccessTokenSigner validates and constructs an access-token signer.
func NewAccessTokenSigner(issuer, audience, keyID string, key []byte, lifetime time.Duration) (AccessTokenSigner, error) {
	if issuer == "" || audience == "" || keyID == "" || len(key) < 32 || lifetime <= 0 {
		return AccessTokenSigner{}, fmt.Errorf("invalid access token configuration")
	}
	return AccessTokenSigner{issuer: issuer, audience: audience, keyID: keyID, key: append([]byte(nil), key...), lifetime: lifetime}, nil
}

// Sign creates an access token bound to the supplied actor and security version.
func (s AccessTokenSigner) Sign(actor Actor, tokenID string, securityVersion int64, now time.Time) (string, error) {
	header, err := json.Marshal(struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
		KeyID     string `json:"kid"`
	}{Algorithm: "HS256", Type: "JWT", KeyID: s.keyID})
	if err != nil {
		return "", fmt.Errorf("marshal token header: %w", err)
	}
	claims := AccessTokenClaims{Issuer: s.issuer, Audience: s.audience, PrincipalType: "tenant", Subject: actor.UserID, TenantID: actor.TenantID, SessionID: actor.SessionID, TokenID: tokenID, SecurityVersion: securityVersion, IssuedAt: now.Unix(), ExpiresAt: now.Add(s.lifetime).Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal token claims: %w", err)
	}
	unsigned := encodeTokenPart(header) + "." + encodeTokenPart(payload)
	return unsigned + "." + encodeTokenPart(sign(s.key, unsigned)), nil
}

// SignPlatform creates a token for a global platform principal without tenant scope.
func (s AccessTokenSigner) SignPlatform(subject UserID, sessionID SessionID, tokenID string, securityVersion int64, now time.Time) (string, error) {
	header, err := json.Marshal(struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
		KeyID     string `json:"kid"`
	}{Algorithm: "HS256", Type: "JWT", KeyID: s.keyID})
	if err != nil {
		return "", fmt.Errorf("marshal token header: %w", err)
	}
	claims := AccessTokenClaims{Issuer: s.issuer, Audience: s.audience, PrincipalType: "platform", Subject: subject, SessionID: sessionID, TokenID: tokenID, SecurityVersion: securityVersion, IssuedAt: now.Unix(), ExpiresAt: now.Add(s.lifetime).Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal token claims: %w", err)
	}
	unsigned := encodeTokenPart(header) + "." + encodeTokenPart(payload)
	return unsigned + "." + encodeTokenPart(sign(s.key, unsigned)), nil
}

// Lifetime returns the configured access-token lifetime.
func (s AccessTokenSigner) Lifetime() time.Duration { return s.lifetime }

// AccessTokenVerifier validates access tokens against a rotating key set.
type AccessTokenVerifier struct {
	issuer    string
	audience  string
	keys      map[string][]byte
	clockSkew time.Duration
}

// NewAccessTokenVerifier constructs a verifier for an issuer and audience.
func NewAccessTokenVerifier(issuer, audience string, keys map[string][]byte, clockSkew time.Duration) AccessTokenVerifier {
	return AccessTokenVerifier{issuer: issuer, audience: audience, keys: keys, clockSkew: clockSkew}
}

// Verify validates a token signature, identity claims, and time bounds.
func (v AccessTokenVerifier) Verify(token string, now time.Time) (AccessTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return AccessTokenClaims{}, ErrInvalidToken
	}
	var header struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
		KeyID     string `json:"kid"`
	}
	if err := decodeTokenJSON(parts[0], &header); err != nil || header.Algorithm != "HS256" || header.Type != "JWT" {
		return AccessTokenClaims{}, ErrInvalidToken
	}
	key, ok := v.keys[header.KeyID]
	if !ok || len(key) < 32 || !hmac.Equal(sign(key, parts[0]+"."+parts[1]), mustDecodeTokenPart(parts[2])) {
		return AccessTokenClaims{}, ErrInvalidToken
	}
	var claims AccessTokenClaims
	if err := decodeTokenJSON(parts[1], &claims); err != nil {
		return AccessTokenClaims{}, ErrInvalidToken
	}
	validPrincipal := (claims.PrincipalType == "tenant" && claims.TenantID != "") || (claims.PrincipalType == "platform" && claims.TenantID == "")
	if claims.Issuer != v.issuer || claims.Audience != v.audience || !validPrincipal || claims.Subject == "" || claims.SessionID == "" || claims.TokenID == "" || claims.SecurityVersion <= 0 {
		return AccessTokenClaims{}, ErrInvalidToken
	}
	if now.After(time.Unix(claims.ExpiresAt, 0).Add(v.clockSkew)) {
		return AccessTokenClaims{}, ErrExpiredToken
	}
	if time.Unix(claims.IssuedAt, 0).After(now.Add(v.clockSkew)) || claims.ExpiresAt <= claims.IssuedAt {
		return AccessTokenClaims{}, ErrInvalidToken
	}
	return claims, nil
}

// NewOpaqueToken returns a URL-safe token backed by cryptographic randomness.
func NewOpaqueToken(bytes int) (string, error) {
	if bytes < 32 {
		return "", fmt.Errorf("opaque token requires at least 32 random bytes")
	}
	value := make([]byte, bytes)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate opaque token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

// HashOpaqueToken returns the only representation persisted for an opaque token.
func HashOpaqueToken(token string) [32]byte { return sha256.Sum256([]byte(token)) }

func encodeTokenPart(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }
func sign(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
func mustDecodeTokenPart(value string) []byte {
	decoded, _ := base64.RawURLEncoding.Strict().DecodeString(value)
	return decoded
}
func decodeTokenJSON(value string, target any) error {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(string(decoded)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
