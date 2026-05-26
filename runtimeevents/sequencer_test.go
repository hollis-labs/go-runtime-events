package runtimeevents

import (
	"sync"
	"testing"
)

func TestSequencerMonotonicPerSession(t *testing.T) {
	s := NewSequencer()
	for i := uint64(1); i <= 5; i++ {
		got := s.Next("ses_a")
		if got != i {
			t.Fatalf("Next(\"ses_a\") iteration %d = %d, want %d", i, got, i)
		}
	}
}

func TestSequencerSessionsAreIndependent(t *testing.T) {
	s := NewSequencer()
	if got := s.Next("ses_a"); got != 1 {
		t.Fatalf("first Next(\"ses_a\") = %d, want 1", got)
	}
	if got := s.Next("ses_b"); got != 1 {
		t.Fatalf("first Next(\"ses_b\") = %d, want 1", got)
	}
	if got := s.Next("ses_a"); got != 2 {
		t.Fatalf("second Next(\"ses_a\") = %d, want 2", got)
	}
}

func TestSequencerForget(t *testing.T) {
	s := NewSequencer()
	s.Next("ses_a")
	s.Next("ses_a")
	s.Forget("ses_a")
	if got := s.Next("ses_a"); got != 1 {
		t.Fatalf("Next after Forget = %d, want 1 (counter reset)", got)
	}
}

func TestSequencerConcurrent(t *testing.T) {
	const goroutines = 8
	const perGoroutine = 1000
	s := NewSequencer()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				s.Next("ses_shared")
			}
		}()
	}
	wg.Wait()

	want := uint64(goroutines*perGoroutine + 1)
	if got := s.Next("ses_shared"); got != want {
		t.Fatalf("Next after concurrent emissions = %d, want %d (sequence races dropped values)", got, want)
	}
}
