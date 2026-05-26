package runtimeevents

import "sync"

// Sequencer assigns per-session monotonic [Event.Sequence] values. It is
// safe for concurrent use. A single Sequencer should be shared by all
// emitters writing into the same session-ID space (typically: one per
// wrapper process).
type Sequencer struct {
	mu       sync.Mutex
	counters map[string]uint64
}

// NewSequencer returns a fresh Sequencer with no recorded sessions.
func NewSequencer() *Sequencer {
	return &Sequencer{counters: make(map[string]uint64)}
}

// Next returns the next sequence number for sessionID. Sequence numbers
// start at 1 and increase by 1 per call within the same session. Sessions
// are independent — Next("a") and Next("b") progress separately.
func (s *Sequencer) Next(sessionID string) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[sessionID]++
	return s.counters[sessionID]
}

// Forget drops the counter for sessionID. Use after a session has ended
// and no further events will be emitted for it.
func (s *Sequencer) Forget(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.counters, sessionID)
}
