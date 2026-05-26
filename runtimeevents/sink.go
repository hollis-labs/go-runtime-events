package runtimeevents

import (
	"context"
	"errors"
)

// Sink consumes runtime events. It is the single neutral seam between
// producers (the wrapper, or any app emitting normalized activity) and
// the storage / transport / UI choices of the surrounding system.
//
// Producers have zero opinion about what a Sink does with an Event —
// fan out, persist, drop, replay, route to a network service, or
// nothing. That choice belongs to the Sink implementation.
//
// Example Sink shapes (none are required; the wrapper ships none of
// these except [FileSink]):
//
//   - An in-memory ring buffer for tests or live UI tailing.
//   - A JSONL file writer for durable local logs — see [FileSink].
//   - A network publisher that ships Events to a remote consumer
//     (Tether's /events/stream, a Kafka topic, an SSE bridge, ...).
//   - A fan-out splitter sending the same Event to multiple downstream
//     sinks — see [MultiSink].
//   - A per-app adapter that translates Events into the app's native
//     event vocabulary (Tether session events, Nanite chat events,
//     Torque task-stream messages, Hadron timeline ticks).
//
// Durability semantics — at-least-once, at-most-once, fsync on write,
// replay on restart — are entirely per-implementation. Sinks that need
// durability should provide their own guarantees and document them;
// producers cannot assume any specific durability posture from the
// Sink interface alone.
//
// Sink.Write is invoked synchronously from the emitter's goroutine.
// Implementations that may block (network, disk fsync, downstream
// queue) should buffer internally so they don't stall the producer's
// IO loop. Implementations that drop on overflow MUST do so silently
// without returning an error — returning an error is reserved for
// conditions the producer can act on.
type Sink interface {
	Write(ctx context.Context, ev Event) error
}

// SinkFunc adapts a plain function to the [Sink] interface.
type SinkFunc func(ctx context.Context, ev Event) error

// Write implements [Sink].
func (f SinkFunc) Write(ctx context.Context, ev Event) error { return f(ctx, ev) }

// MultiSink fans an [Event] out to multiple sinks. Errors from individual
// sinks are joined into a single error using [errors.Join]; all sinks are
// always called even when an earlier sink returned an error.
type MultiSink []Sink

// Write implements [Sink].
func (m MultiSink) Write(ctx context.Context, ev Event) error {
	var errs []error
	for _, s := range m {
		if err := s.Write(ctx, ev); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
