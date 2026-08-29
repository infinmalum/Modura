// Package identifier creates UUIDv7 identifiers without global mutable state.
package identifier

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"
)

// UUID is the canonical textual representation of a UUID.
type UUID string

// NewUUIDv7 creates a time-ordered UUID using the supplied clock and entropy.
func NewUUIDv7(now time.Time, entropy io.Reader) (UUID, error) {
	if entropy == nil {
		entropy = rand.Reader
	}
	var value [16]byte
	if _, err := io.ReadFull(entropy, value[:]); err != nil {
		return "", fmt.Errorf("read UUID entropy: %w", err)
	}
	milliseconds := uint64(now.UTC().UnixMilli())
	value[0] = byte(milliseconds >> 40)
	value[1] = byte(milliseconds >> 32)
	value[2] = byte(milliseconds >> 24)
	value[3] = byte(milliseconds >> 16)
	value[4] = byte(milliseconds >> 8)
	value[5] = byte(milliseconds)
	value[6] = value[6]&0x0f | 0x70
	value[8] = value[8]&0x3f | 0x80
	return UUID(fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16])), nil
}
