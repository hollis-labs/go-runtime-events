package runtimeevents

import (
	"strings"
	"testing"
)

func TestNewIDsHavePrefix(t *testing.T) {
	cases := []struct {
		name   string
		gen    func() string
		prefix string
	}{
		{"event", NewEventID, "evt_"},
		{"session", NewSessionID, "ses_"},
		{"turn", NewTurnID, "turn_"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			id := c.gen()
			if !strings.HasPrefix(id, c.prefix) {
				t.Fatalf("%s: id %q missing prefix %q", c.name, id, c.prefix)
			}
			// 16 bytes hex-encoded = 32 chars after prefix.
			if got := len(id) - len(c.prefix); got != 32 {
				t.Fatalf("%s: id %q has %d hex chars after prefix, want 32", c.name, id, got)
			}
		})
	}
}

func TestNewIDsAreUnique(t *testing.T) {
	const n = 1000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id := NewEventID()
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate id at iteration %d: %s", i, id)
		}
		seen[id] = struct{}{}
	}
}
