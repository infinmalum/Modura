package identity

import "testing"

func TestNormalizeLogin(t *testing.T) {
	if got, want := NormalizeLogin("  UsÉr@Example.COM "), "usér@example.com"; got != want {
		t.Fatalf("NormalizeLogin() = %q, want %q", got, want)
	}
}
