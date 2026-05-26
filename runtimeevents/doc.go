// Package runtimeevents defines the runtime activity event envelope emitted
// by the go-agent-wrapper harness when a CLI-agent subprocess is launched,
// observed, and governed.
//
// The package is intentionally minimal: it owns the on-the-wire schema and
// nothing else. Per-kind payload shapes are intentionally opaque
// (json.RawMessage) so the schema stays stable while individual apps and
// adapters evolve their own payload conventions independently.
//
// Apps that observe these events should:
//
//   - Treat any unknown [EventKind] as opaque rather than rejecting it. New
//     kinds will be added; old consumers must not break.
//   - Trust [Event.Sequence] for per-session ordering. Wall-clock time is
//     informative; sequence is authoritative.
//   - Use [Event.RawOffset] (when set) for byte-level replay of stdin/stdout/
//     stderr streams against the durable raw log.
//
// The wrapper assigns sequence numbers and IDs; downstream consumers should
// not regenerate them. See [Sequencer], [Emitter], and [Sink] for the
// producer-side helpers.
package runtimeevents
