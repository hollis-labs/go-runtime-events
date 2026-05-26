# Changelog

All notable changes to go-runtime-events are documented here. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.1.0 — 2026-05-26

Initial cut. 24 tests, all `-race` clean.

### Added

- **`Event` envelope** with all required fields locked in from day one:
  schema_version, id, kind, time, app, session_id, turn_id, sequence,
  parent_id, raw_offset, process, source, payload.
- **31 `EventKind` constants** covering process lifecycle, session
  lifecycle, turn lifecycle, stdin/stdout/stderr raw + line, agent
  semantic events (delta, tool_use, tool_result, subagent_spawn,
  permission_*), policy actions (nudge/rewrite/block), planting,
  sandbox, and interrupt request/acknowledged.
- **`SourceChannel` constants** for claude-stream-json, opencode-plugin,
  jsonrpc, pty, stdio, hook, filter. Open string — adapters may
  introduce new channels without a schema bump.
- **`Confidence` levels** (exact / derived / inferred) so consumers can
  trust enforcement decisions from semantic channels and treat text
  classifier hits as advisory.
- **`Sequencer`** — per-session monotonic counter, concurrent-safe.
  Authoritative ordering source per the architecture doc.
- **ID generators** — `NewEventID` (`evt_`), `NewSessionID` (`ses_`),
  `NewTurnID` (`turn_`); 16 random bytes hex-encoded via `crypto/rand`.
- **`Emitter`** — convenience producer with `Now` override (test
  clock), per-session identity binding, and `EmitOption`s: `WithID`,
  `WithTurnID`, `WithParentID`, `WithRawOffset`, `WithProcess`.
- **Thread-safe `SetProcess` / `SetProviderSessionID` mutators** on
  Emitter for late-arriving provider session IDs while concurrent
  emitters are running.
- **`Sink` interface** — neutral about durability; producers have zero
  opinion about what sinks do with events.
- **`SinkFunc`** for plain-function adapters.
- **`MultiSink`** for fan-out; errors joined via `errors.Join`.
- **`FileSink`** — reference JSONL append-only sink, mutex-guarded,
  idempotent `Close`, `ErrClosed` sentinel.
- **Initial module scaffold from folio's `go-lib` preset** — CI
  workflow, MIT license.

### Notes

- The schema is intentionally minimal: per-kind payload shapes stay
  opaque (`json.RawMessage`) so the envelope can remain stable while
  apps and adapters evolve their own payload conventions independently.
- Producers are expected to treat unknown kinds as round-trip-able;
  consumers must not reject unknown kinds.
