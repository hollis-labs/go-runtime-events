# Changelog

All notable changes to go-runtime-events are documented here. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.2.2 — 2026-10-02

Documentation only: the payload conventions a turn reducer reads
(CW-20261002-0061, ADR 0049). Envelope, kinds, `SchemaVersion` and wire values
are unchanged.

### Added

- `turn.completed` `text`: the turn's final message when the runtime reports
  one on its terminal event. No producer emits it yet; go-agent-wrapper starts
  once go-providers carries Claude's `result.result` and agy's `result.response`
  on the stream's done event.
- `turn.failed` `reason` (`interrupted`, `process_exited`), which
  go-agent-wrapper already emits, so a consumer can tell a cut-short turn from a
  failed one.
- `agent.tool_use` and `agent.permission_requested` / `agent.permission_resolved`
  field conventions, written down from what go-agent-wrapper and its ACP
  clients emit today.

## v0.2.1 — 2026-10-01

Documentation only: the payload conventions now match what producers emit
(CW-20260930-0137).

### Changed

- `agent.delta` `phase`: thinking text is `thought` (the value the ACP
  translators already emit, and that consumers key on), not `thinking`; ACP
  agent text is `message`; native text is `narration` / `final` when the
  provider classifies it.
- `usage` on terminal events: native runtimes emit go-llm-types' `Usage` with
  its Go field names (`CostUSD`, not `cost_usd`); ACP runtimes pass the
  agent's usage object through. v0.2.0's wording implied snake_case keys.

## v0.2.0 — 2026-10-01

Additive: three kinds and documented payload conventions (CW-20260930-0137).
Envelope, `SchemaVersion` and existing wire values are unchanged.

### Added

- `KindSessionLost` (`session.lost`): a resume did not continue the provider
  session it asked for.
- `KindSessionAuthFailed` (`session.auth_failed`): the provider CLI is not
  signed in or its credentials were refused.
- `KindAgentPermissionDenied` (`agent.permission_denied`): the agent CLI
  refused a tool action on its own because it could not ask for approval.
- Payload conventions, in the package doc and README: `block_id` and `phase`
  on `agent.delta`; `usage` (with `cost_usd`) and a normalised `stop_reason`
  on `turn.completed` / `turn.failed`; the fields of the three new kinds.
  Payloads stay opaque; these are conventions, not typed structs.

### Changed

- Raised the module's `go` directive to `1.26.6` (Go floor across the portfolio); CI now uses `go-version-file: go.mod`.

## v0.1.2 — 2026-09-04

Documentation and test-hygiene patch release. Public Go declarations, wire
values, `SchemaVersion`, and runtime behavior are unchanged from v0.1.1. 27
tests, all `-race` clean.

### Changed

- Corrected the public documentation for the action-shaped `KindPolicy*`
  constants and `policy.*` wire values. They are legacy compatibility labels
  carrying observed policy findings or recommendations, not evidence that an
  engine changed, blocked, or paused an operation. Constant names and wire
  values are unchanged.

### Notes

- No new module dependencies — `go.mod` still has zero requires.
- Test file-reader cleanup now checks close errors so the repository's pinned
  CI lint gate is clean; library code and behavior are unaffected.

## v0.1.1 — 2026-08-25

Additive release of commit `8756744` ("Add policy approval event helpers").
No breaking changes; every v0.1.0 API keeps its signature and `SchemaVersion`
stays `"1"`. 27 tests, all `-race` clean. Tag cut after the fact, so the
tagged tree does not contain this entry.

### Added

- **`Emitter.EmitReturning(ctx, kind, source, payload, opts...) (string, error)`**
  — emits and returns the assigned event ID once the sink write succeeds.
  This is the safe way to build a `WithParentID` correlation chain: callers
  that only need to know the ID after the fact no longer have to reach for
  `WithID`'s override power. It returns `""` on both marshal failure and sink
  failure, and honors every `EmitOption`, so an explicit `WithID` is reflected
  in the returned value. `Emit` is now a thin wrapper over it and its
  behavior, including the `runtimeevents: marshal payload for kind %q` error
  wrapping, is unchanged.

- **`KindPolicyApprovalRequested` (`"policy.approval_requested"`)** — a
  dedicated legacy compatibility label for an observed approval
  recommendation, which previously had to be overloaded onto `policy.block`.
  Backwards-additive: 29 `EventKind` constants at this tag, and consumers
  already have to round-trip unknown kinds. The event does not itself pause
  execution; the pause/resume operator-approval flow still belongs to the
  wrapper/app layer.

### Notes

- No new module dependencies — `go.mod` still has zero requires.

## v0.1.0 — 2026-05-26

Initial cut. 24 tests, all `-race` clean.

### Added

- **`Event` envelope** with all required fields locked in from day one:
  schema_version, id, kind, time, app, session_id, turn_id, sequence,
  parent_id, raw_offset, process, source, payload.
- **28 `EventKind` constants** covering process lifecycle, session
  lifecycle, turn lifecycle, stdin/stdout/stderr raw + line, agent
  semantic events (delta, tool_use, tool_result, subagent_spawn,
  permission_*), policy observation compatibility labels
  (nudge/rewrite/block), planting,
  sandbox, and interrupt request/acknowledged.
- **`SourceChannel` constants** for claude-stream-json, opencode-plugin,
  jsonrpc, pty, stdio, hook, filter. Open string — adapters may
  introduce new channels without a schema bump.
- **`Confidence` levels** (exact / derived / inferred) so consumers can
  distinguish direct semantic observations from advisory text-classifier
  findings.
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
