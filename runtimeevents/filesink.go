package runtimeevents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

// FileSink writes [Event]s as JSONL (one JSON object per line) to a
// file on disk. It is the reference [Sink] implementation that ships
// with this package — a worked example of the contract, suitable for
// tests, local debugging, and simple durable-log use cases where
// network or DB backing is overkill.
//
// FileSink is append-only and safe for concurrent Write calls. It does
// NOT fsync per write — the kernel buffer is flushed on [FileSink.Close]
// (via file close), which is sufficient for crash-tolerance against
// graceful shutdown but NOT against power loss. Callers that need
// stronger durability should wrap FileSink or implement their own Sink.
type FileSink struct {
	mu     sync.Mutex
	f      *os.File
	closed bool
}

// OpenFileSink opens (or creates) path for append and returns a
// FileSink writing to it. The caller owns the path — FileSink does
// not rotate, truncate, or manage retention.
func OpenFileSink(path string) (*FileSink, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644) //nolint:gosec // caller-supplied path
	if err != nil {
		return nil, fmt.Errorf("runtimeevents: open file sink at %q: %w", path, err)
	}
	return &FileSink{f: f}, nil
}

// Write implements [Sink]. It marshals ev to JSON, appends a newline,
// and writes the result atomically with respect to other concurrent
// Write calls on the same FileSink.
//
// Returns [ErrClosed] if [FileSink.Close] has already been called.
func (s *FileSink) Write(_ context.Context, ev Event) error {
	raw, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("runtimeevents: marshal event %s: %w", ev.ID, err)
	}
	raw = append(raw, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}
	if _, err := s.f.Write(raw); err != nil {
		return fmt.Errorf("runtimeevents: write event %s: %w", ev.ID, err)
	}
	return nil
}

// Close flushes and closes the underlying file. Subsequent calls to
// [FileSink.Write] return [ErrClosed]. Close is idempotent.
func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.f.Close()
}

// ErrClosed is returned by [FileSink.Write] after [FileSink.Close]
// has been called. Use [errors.Is] to detect.
var ErrClosed = errors.New("runtimeevents: sink closed")
