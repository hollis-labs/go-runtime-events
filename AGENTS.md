# go-runtime-events

The shared on-the-wire event envelope for CLI-wrapped agent subprocesses. It
owns the `Event` envelope, the kind and source vocabularies, per-session
sequencing, ID generation and the `Sink` surface — and nothing else. Per-kind
payloads stay opaque `json.RawMessage` so producers and consumers can evolve
payload conventions without moving the schema. `go-agent-wrapper` produces
these; Nanite, Tether, Torque, Hadron and Stack Explorer consume them.

## Start Here

- `README.md` enumerates the envelope fields and the current constant counts.
- `ROADMAP.md` records deferred scope.
- `runtimeevents/schema.go` defines the `Event` envelope.
- `runtimeevents/kinds.go` holds the `EventKind`, `SourceChannel` and
  `Confidence` vocabularies.
- `runtimeevents/emitter.go` and `runtimeevents/sequencer.go` own emission and
  per-session monotonic sequence.
- `runtimeevents/ids.go` owns the `evt_` / `ses_` / `turn_` prefixed IDs.
- `runtimeevents/sink.go` and `runtimeevents/filesink.go` own `Sink`,
  `MultiSink` and the reference JSONL sink.

## Commands

```bash
go vet ./...
go test -race -count=1 ./...
```

CI runs both.

## Boundaries

This is a wire schema with independent producers and consumers across the
portfolio, so envelope fields and `EventKind` values are compatibility surface.
Renaming a field, changing a JSON tag or repurposing a kind breaks readers of
already-written JSONL, which `schema_version` exists to manage — bump it rather
than redefining a field in place.

Payloads stay opaque here on purpose. Adding typed per-kind payload structs
would pull every producer's vocabulary into this module and end the
independence that makes it safe to depend on.

Sequence is per-session and monotonic, and the sequencer is concurrent-safe
(`TestEmitterSequenceMonotonic`). Consumers order and de-duplicate on it.

Sinks must be robust in the ways the tests name: concurrent writes to
`FileSink` must not interleave, close is idempotent, a write after close
returns `ErrClosed`, appends survive reopening, and `MultiSink` calls every
sink even when one errors (joining the errors) rather than stopping at the
first. That last one matters — a failing sink must not silently cost you the
others.
