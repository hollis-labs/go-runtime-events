package runtimeevents

import (
	"encoding/json"
	"time"
)

// SchemaVersion is the on-the-wire version of the [Event] envelope. Bump
// when the envelope shape changes in a non-additive way. Adding new
// [EventKind] values, new [SourceChannel] values, or new optional fields
// does not require a bump.
const SchemaVersion = "1"

// Event is the runtime activity event envelope.
//
// Producers must populate SchemaVersion, ID, Kind, Time, SessionID,
// Sequence, and Source.Channel. All other fields are optional and may be
// omitted when not meaningful for the event.
//
// Payload is opaque to this package; per-kind payload shapes are defined
// by the wrapper, by adapters, or by consuming apps. Apps that need to
// inspect a payload should switch on Kind and unmarshal Payload into the
// app-local type for that kind.
type Event struct {
	SchemaVersion string    `json:"schema_version"`
	ID            string    `json:"id"`
	Kind          EventKind `json:"kind"`
	Time          time.Time `json:"time"`

	App       string `json:"app,omitempty"`
	SessionID string `json:"session_id"`
	TurnID    string `json:"turn_id,omitempty"`
	Sequence  uint64 `json:"sequence"`

	// ParentID correlates request/response pairs (a response carries the
	// request event's ID) and chains derived events back to their source.
	ParentID string `json:"parent_id,omitempty"`

	// RawOffset, when set, is the byte offset into the per-session raw log
	// where this event's content begins. Used for byte-level replay of
	// stdin/stdout/stderr streams. Pointer so a zero offset is
	// distinguishable from "unset".
	RawOffset *int64 `json:"raw_offset,omitempty"`

	Process Process         `json:"process"`
	Source  Source          `json:"source"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Process describes the subprocess the event pertains to. Zero values mean
// "unknown" or "not yet assigned" — e.g., a plant.started event emitted
// before the process spawns will have PID == 0.
type Process struct {
	PID int `json:"pid,omitempty"`

	// Provider is the upstream agent identity ("claude", "codex",
	// "opencode", ...). Open string — apps may use provider names this
	// package does not enumerate.
	Provider string `json:"provider,omitempty"`

	// Runtime is the transport the wrapper selected for this provider
	// ("pty", "streaming-stdio", "jsonrpc-stdio", "http-sse", ...).
	Runtime string `json:"runtime,omitempty"`

	// ProviderSessionID is the provider's own session identity (e.g., a
	// Claude session UUID), when the provider exposes one. SessionID at
	// the [Event] level always refers to the wrapper's session identity.
	ProviderSessionID string `json:"provider_session_id,omitempty"`
}

// Source describes how the event was observed. Channel identifies the
// observation transport; Confidence describes how directly the underlying
// signal maps to the event semantics.
type Source struct {
	Channel    SourceChannel `json:"channel"`
	Confidence Confidence    `json:"confidence,omitempty"`
}
