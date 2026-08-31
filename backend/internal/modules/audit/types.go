// Package audit owns durable business audit events.
package audit

import (
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
}
