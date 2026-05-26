package runtimeevents

import (
	"crypto/rand"
	"encoding/hex"
)

// ID prefixes. Producers should use [NewEventID], [NewSessionID], and
// [NewTurnID] rather than constructing IDs by hand so prefix conventions
// stay consistent across the portfolio.
const (
	prefixEvent   = "evt_"
	prefixSession = "ses_"
	prefixTurn    = "turn_"
)

// NewEventID returns a fresh event ID with the "evt_" prefix.
func NewEventID() string { return prefixEvent + randHex(16) }

// NewSessionID returns a fresh wrapper-session ID with the "ses_" prefix.
// Wrapper sessions are distinct from provider-internal session IDs (which
// live on [Process.ProviderSessionID]).
func NewSessionID() string { return prefixSession + randHex(16) }

// NewTurnID returns a fresh turn ID with the "turn_" prefix.
func NewTurnID() string { return prefixTurn + randHex(16) }

func randHex(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read on modern Go cannot fail; if it ever does the
		// process state is unrecoverable. Panic so callers don't get
		// silently-colliding IDs.
		panic("runtimeevents: crypto/rand.Read failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}
