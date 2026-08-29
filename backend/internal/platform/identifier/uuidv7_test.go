package identifier

import (
	"bytes"
	"testing"
	"time"
)

func TestNewUUIDv7(t *testing.T) {
	id, err := NewUUIDv7(time.UnixMilli(1_700_000_000_000), bytes.NewReader(make([]byte, 16)))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(id), "018bcfe5-6800-7000-8000-000000000000"; got != want {
		t.Fatalf("id = %q, want %q", got, want)
	}
}
