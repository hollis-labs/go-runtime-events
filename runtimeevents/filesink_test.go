package runtimeevents

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func newTestEvent(id string, seq uint64) Event {
	return Event{
		SchemaVersion: SchemaVersion,
		ID:            id,
		Kind:          KindSessionReady,
		Time:          time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC),
		App:           "test",
		SessionID:     "ses_test",
		Sequence:      seq,
		Source:        Source{Channel: ChannelPTY},
	}
}

func TestFileSinkRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	sink, err := OpenFileSink(path)
	if err != nil {
		t.Fatalf("OpenFileSink: %v", err)
	}

	want := []Event{
		newTestEvent("evt_1", 1),
		newTestEvent("evt_2", 2),
		newTestEvent("evt_3", 3),
	}
	for _, ev := range want {
		if err := sink.Write(context.Background(), ev); err != nil {
			t.Fatalf("Write %s: %v", ev.ID, err)
		}
	}
	if err := sink.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %q: %v", path, err)
	}
	defer f.Close()

	var got []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var ev Event
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			t.Fatalf("unmarshal line %q: %v", scanner.Text(), err)
		}
		got = append(got, ev)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("read %d lines, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i].ID || got[i].Sequence != want[i].Sequence {
			t.Errorf("event %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestFileSinkAppendAcrossOpens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")

	s1, err := OpenFileSink(path)
	if err != nil {
		t.Fatalf("OpenFileSink #1: %v", err)
	}
	if err := s1.Write(context.Background(), newTestEvent("evt_a", 1)); err != nil {
		t.Fatalf("Write #1: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close #1: %v", err)
	}

	s2, err := OpenFileSink(path)
	if err != nil {
		t.Fatalf("OpenFileSink #2: %v", err)
	}
	if err := s2.Write(context.Background(), newTestEvent("evt_b", 2)); err != nil {
		t.Fatalf("Write #2: %v", err)
	}
	if err := s2.Close(); err != nil {
		t.Fatalf("Close #2: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	lines := 0
	for _, c := range b {
		if c == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Errorf("lines = %d, want 2 (append across opens lost data)", lines)
	}
}

func TestFileSinkWriteAfterCloseReturnsErrClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	sink, err := OpenFileSink(path)
	if err != nil {
		t.Fatalf("OpenFileSink: %v", err)
	}
	if err := sink.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	err = sink.Write(context.Background(), newTestEvent("evt_x", 1))
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("Write after Close: err=%v, want ErrClosed", err)
	}
}

func TestFileSinkCloseIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	sink, err := OpenFileSink(path)
	if err != nil {
		t.Fatalf("OpenFileSink: %v", err)
	}
	if err := sink.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := sink.Close(); err != nil {
		t.Errorf("second Close: %v (want nil — Close should be idempotent)", err)
	}
}

func TestFileSinkConcurrentWritesDontInterleave(t *testing.T) {
	const goroutines = 8
	const perGoroutine = 100

	path := filepath.Join(t.TempDir(), "events.jsonl")
	sink, err := OpenFileSink(path)
	if err != nil {
		t.Fatalf("OpenFileSink: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				ev := newTestEvent("evt_g_x", uint64(g*perGoroutine+i+1))
				if err := sink.Write(context.Background(), ev); err != nil {
					t.Errorf("Write: %v", err)
					return
				}
			}
		}(g)
	}
	wg.Wait()
	if err := sink.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Every line must parse as a complete Event. If writes interleaved,
	// json.Unmarshal would fail on at least one line.
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1<<16), 1<<20)
	count := 0
	for scanner.Scan() {
		var ev Event
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			t.Fatalf("interleaved write detected at line %d: %v\n  raw: %q", count, err, scanner.Text())
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if want := goroutines * perGoroutine; count != want {
		t.Errorf("lines = %d, want %d", count, want)
	}
}
