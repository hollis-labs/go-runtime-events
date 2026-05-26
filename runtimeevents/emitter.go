package runtimeevents

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Emitter is a convenience producer that fills in [Event] envelope fields
// the caller would otherwise repeat on every emission (App, SessionID,
// Process, sequence, time, ID). It is safe for concurrent use.
//
// A wrapper typically owns one Emitter per session, configured with the
// session's identity and a shared [Sink] and [Sequencer].
//
// Process: the embedded process metadata is safe for concurrent read
// (Emit takes an internal lock). Use [Emitter.SetProcess] or
// [Emitter.SetProviderSessionID] to mutate it from goroutines other
// than the one that constructed the Emitter. Direct field assignment
// to Emitter.Process is permitted before the Emitter is shared across
// goroutines (typically during one-time setup) but is racy after
// that.
type Emitter struct {
	Sink      Sink
	App       string
	SessionID string
	Process   Process
	Sequencer *Sequencer

	// Now overrides the timestamp source. Leave nil to use [time.Now].
	// Tests inject a fixed clock here.
	Now func() time.Time

	// processMu guards concurrent access to Process from Emit (reader)
	// and the SetProcess/SetProviderSessionID mutators. Direct field
	// assignment is racy once the Emitter is shared across goroutines;
	// the setter methods exist for that case.
	processMu sync.RWMutex
}

// SetProcess replaces the process metadata under the Emitter's lock.
// Safe for concurrent use; the next [Emitter.Emit] call observes the
// new value.
func (e *Emitter) SetProcess(p Process) {
	e.processMu.Lock()
	defer e.processMu.Unlock()
	e.Process = p
}

// SetProviderSessionID updates only the ProviderSessionID field on
// the embedded Process under the Emitter's lock. Safe for concurrent
// use. This is the typical lifecycle: the wrapper-session is bound at
// Emitter setup time, and the provider-side session ID arrives later
// in the stream once the underlying adapter assigns one.
func (e *Emitter) SetProviderSessionID(id string) {
	e.processMu.Lock()
	defer e.processMu.Unlock()
	e.Process.ProviderSessionID = id
}

// snapshotProcess returns a copy of the current Process under the
// reader lock. Used by Emit to safely materialize ev.Process.
func (e *Emitter) snapshotProcess() Process {
	e.processMu.RLock()
	defer e.processMu.RUnlock()
	return e.Process
}

// EmitOption mutates an [Event] just before it is handed to the [Sink].
// Options compose; later options override earlier ones on the same field.
type EmitOption func(*Event)

// WithID overrides the auto-generated [Event.ID]. Use carefully —
// duplicate IDs in the same session break correlation and replay.
// Primary use: emitting two correlated events in sequence and using
// the predetermined first ID as the second event's [WithParentID]
// argument.
func WithID(id string) EmitOption {
	return func(e *Event) {
		if id != "" {
			e.ID = id
		}
	}
}

// WithTurnID sets [Event.TurnID].
func WithTurnID(id string) EmitOption { return func(e *Event) { e.TurnID = id } }

// WithParentID sets [Event.ParentID] for correlation with a prior event.
func WithParentID(id string) EmitOption { return func(e *Event) { e.ParentID = id } }

// WithRawOffset sets [Event.RawOffset] for byte-level raw-log replay.
func WithRawOffset(off int64) EmitOption {
	return func(e *Event) { e.RawOffset = &off }
}

// WithProcess overrides the emitter's default [Process] for this event.
// Useful when a single emitter observes multiple child processes.
func WithProcess(p Process) EmitOption { return func(e *Event) { e.Process = p } }

// Emit assembles an [Event] from the emitter's defaults plus the caller's
// kind/source/payload/options and writes it to [Emitter.Sink].
//
// payload may be nil (no payload), a []byte / json.RawMessage (used
// verbatim), or any value json.Marshal accepts (marshaled here).
func (e *Emitter) Emit(ctx context.Context, kind EventKind, source Source, payload any, opts ...EmitOption) error {
	raw, err := marshalPayload(payload)
	if err != nil {
		return fmt.Errorf("runtimeevents: marshal payload for kind %q: %w", kind, err)
	}

	ev := Event{
		SchemaVersion: SchemaVersion,
		ID:            NewEventID(),
		Kind:          kind,
		Time:          e.now(),
		App:           e.App,
		SessionID:     e.SessionID,
		Sequence:      e.Sequencer.Next(e.SessionID),
		Process:       e.snapshotProcess(),
		Source:        source,
		Payload:       raw,
	}
	for _, opt := range opts {
		opt(&ev)
	}
	return e.Sink.Write(ctx, ev)
}

func (e *Emitter) now() time.Time {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now().UTC()
}

func marshalPayload(p any) (json.RawMessage, error) {
	switch v := p.(type) {
	case nil:
		return nil, nil
	case json.RawMessage:
		return v, nil
	case []byte:
		return json.RawMessage(v), nil
	default:
		return json.Marshal(v)
	}
}
