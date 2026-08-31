// Package audit owns durable business audit events.
package audit

import (
	"encoding/json"
	"time"

	"github.com/modura-dev/modura/backend/internal/modules/identity"
)

// Event is immutable evidence for a completed tenant business write.
type Event struct {
	ID            string
	ActorID       identity.UserID
	TenantID      identity.TenantID
	Action        string
	Resource      string
	ResourceID    string
	Reason        string
	CorrelationID string
	OccurredAt    time.Time
	BeforeState   json.RawMessage
	AfterState    json.RawMessage
}

// PlatformEvent is immutable evidence for a completed global platform write.
// It is deliberately separate from Event so a missing tenant cannot be
// mistaken for an incompletely scoped tenant operation.
type PlatformEvent struct {
	ID            string
	ActorID       string
	Action        string
	Resource      string
	ResourceID    string
	Reason        string
	CorrelationID string
	OccurredAt    time.Time
	BeforeState   json.RawMessage
	AfterState    json.RawMessage
}

// Query is a bounded tenant audit search.
type Query struct {
	TenantID identity.TenantID
	Action   string
	Resource string
	Limit    int
	Offset   int
}

// Record is the redacted audit projection returned to authorized readers.
type Record struct {
	ID            string
	ActorType     string
	ActorID       string
	TenantID      identity.TenantID
	Action        string
	Resource      string
	ResourceID    string
	Reason        string
	Result        string
	CorrelationID string
	OccurredAt    time.Time
	BeforeState   json.RawMessage
	AfterState    json.RawMessage
}
