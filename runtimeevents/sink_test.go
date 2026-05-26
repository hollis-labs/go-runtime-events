package runtimeevents

import (
	"context"
	"errors"
	"testing"
)

func TestMultiSinkFanOut(t *testing.T) {
	var a, b captureSink
	multi := MultiSink{&a, &b}

	ev := Event{ID: "evt_1"}
	if err := multi.Write(context.Background(), ev); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(a.events) != 1 || len(b.events) != 1 {
		t.Fatalf("fan-out incomplete: a=%d b=%d", len(a.events), len(b.events))
	}
}

func TestMultiSinkCallsAllEvenAfterError(t *testing.T) {
	failErr := errors.New("a failed")
	a := &captureSink{err: failErr}
	b := &captureSink{}
	multi := MultiSink{a, b}

	err := multi.Write(context.Background(), Event{ID: "evt_1"})
	if !errors.Is(err, failErr) {
		t.Errorf("Write error = %v, want it to wrap %v", err, failErr)
	}
	if len(b.events) != 1 {
		t.Errorf("second sink was skipped after first failed; len(b.events) = %d", len(b.events))
	}
}

func TestSinkFunc(t *testing.T) {
	var got Event
	var s Sink = SinkFunc(func(_ context.Context, ev Event) error {
		got = ev
		return nil
	})
	if err := s.Write(context.Background(), Event{ID: "evt_x"}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got.ID != "evt_x" {
		t.Errorf("SinkFunc did not receive event: got ID %q", got.ID)
	}
}
