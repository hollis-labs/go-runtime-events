package runtimeevents

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

type captureSink struct {
	events []Event
	err    error
}

func (c *captureSink) Write(_ context.Context, ev Event) error {
	c.events = append(c.events, ev)
	return c.err
}

func newTestEmitter(sink Sink) *Emitter {
	fixed := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	return &Emitter{
		Sink:      sink,
		App:       "test-app",
		SessionID: "ses_test",
		Process:   Process{Provider: "claude", Runtime: "pty"},
		Sequencer: NewSequencer(),
		Now:       func() time.Time { return fixed },
	}
}

func TestEmitterFillsEnvelopeDefaults(t *testing.T) {
	sink := &captureSink{}
	em := newTestEmitter(sink)

	err := em.Emit(context.Background(), KindSessionReady, Source{Channel: ChannelPTY, Confidence: ConfidenceExact}, nil)
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}

	if len(sink.events) != 1 {
		t.Fatalf("sink.events len = %d, want 1", len(sink.events))
	}
	ev := sink.events[0]

	if ev.SchemaVersion != SchemaVersion {
		t.Errorf("SchemaVersion = %q, want %q", ev.SchemaVersion, SchemaVersion)
	}
	if !strings.HasPrefix(ev.ID, "evt_") {
		t.Errorf("ID = %q, want evt_ prefix", ev.ID)
	}
	if ev.App != "test-app" || ev.SessionID != "ses_test" {
		t.Errorf("identity not propagated: app=%q session=%q", ev.App, ev.SessionID)
	}
	if ev.Sequence != 1 {
		t.Errorf("Sequence = %d, want 1 (first emit on this session)", ev.Sequence)
	}
	if ev.Process.Provider != "claude" || ev.Process.Runtime != "pty" {
		t.Errorf("Process not propagated: %#v", ev.Process)
	}
	if ev.Kind != KindSessionReady {
		t.Errorf("Kind = %q, want %q", ev.Kind, KindSessionReady)
	}
	if ev.Source.Channel != ChannelPTY {
		t.Errorf("Source.Channel = %q, want %q", ev.Source.Channel, ChannelPTY)
	}
}

func TestEmitterSequenceMonotonic(t *testing.T) {
	sink := &captureSink{}
	em := newTestEmitter(sink)
	for i := 0; i < 3; i++ {
		if err := em.Emit(context.Background(), KindStdoutLine, Source{Channel: ChannelPTY}, nil); err != nil {
			t.Fatalf("Emit %d: %v", i, err)
		}
	}
	for i, ev := range sink.events {
		want := uint64(i + 1)
		if ev.Sequence != want {
			t.Errorf("event %d sequence = %d, want %d", i, ev.Sequence, want)
		}
	}
}

func TestEmitterPayloadShapes(t *testing.T) {
	cases := []struct {
		name    string
		payload any
		want    string // expected JSON substring of Payload
	}{
		{"nil-payload", nil, ""},
		{"struct-payload", map[string]any{"k": "v"}, `{"k":"v"}`},
		{"raw-message-payload", json.RawMessage(`{"already":"raw"}`), `{"already":"raw"}`},
		{"bytes-payload", []byte(`{"as":"bytes"}`), `{"as":"bytes"}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sink := &captureSink{}
			em := newTestEmitter(sink)
			if err := em.Emit(context.Background(), KindAgentDelta, Source{Channel: ChannelClaudeStreamJSON}, c.payload); err != nil {
				t.Fatalf("Emit: %v", err)
			}
			got := string(sink.events[0].Payload)
			if c.want == "" {
				if got != "" {
					t.Errorf("nil payload produced non-empty Payload: %q", got)
				}
				return
			}
			if got != c.want {
				t.Errorf("Payload = %q, want %q", got, c.want)
			}
		})
	}
}

func TestEmitterOptionsOverrideEnvelope(t *testing.T) {
	sink := &captureSink{}
	em := newTestEmitter(sink)
	otherProc := Process{Provider: "codex", Runtime: "jsonrpc-stdio"}
	if err := em.Emit(
		context.Background(),
		KindInterruptAcknowledged,
		Source{Channel: ChannelJSONRPC, Confidence: ConfidenceExact},
		nil,
		WithTurnID("turn_42"),
		WithParentID("evt_request"),
		WithRawOffset(4096),
		WithProcess(otherProc),
	); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	ev := sink.events[0]
	if ev.TurnID != "turn_42" {
		t.Errorf("TurnID = %q, want turn_42", ev.TurnID)
	}
	if ev.ParentID != "evt_request" {
		t.Errorf("ParentID = %q, want evt_request", ev.ParentID)
	}
	if ev.RawOffset == nil || *ev.RawOffset != 4096 {
		t.Errorf("RawOffset = %v, want 4096", ev.RawOffset)
	}
	if ev.Process != otherProc {
		t.Errorf("Process = %#v, want %#v", ev.Process, otherProc)
	}
}

func TestEmitterWithIDOverridesGenerated(t *testing.T) {
	sink := &captureSink{}
	em := newTestEmitter(sink)
	if err := em.Emit(
		context.Background(),
		KindAgentToolUse,
		Source{Channel: ChannelClaudeStreamJSON},
		nil,
		WithID("evt_predetermined"),
	); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	if got := sink.events[0].ID; got != "evt_predetermined" {
		t.Errorf("ID = %q, want evt_predetermined (override ignored)", got)
	}
}

func TestEmitterWithEmptyIDPreservesGenerated(t *testing.T) {
	sink := &captureSink{}
	em := newTestEmitter(sink)
	if err := em.Emit(
		context.Background(),
		KindAgentToolUse,
		Source{Channel: ChannelClaudeStreamJSON},
		nil,
		WithID(""),
	); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	if got := sink.events[0].ID; got == "" || got == "evt_predetermined" {
		t.Errorf("WithID(\"\") should leave the generated ID intact; got %q", got)
	}
}

func TestEmitterPropagatesSinkError(t *testing.T) {
	wantErr := errors.New("sink failed")
	sink := &captureSink{err: wantErr}
	em := newTestEmitter(sink)
	err := em.Emit(context.Background(), KindSessionReady, Source{Channel: ChannelPTY}, nil)
	if !errors.Is(err, wantErr) {
		t.Errorf("Emit error = %v, want %v", err, wantErr)
	}
}
