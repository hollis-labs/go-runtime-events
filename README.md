# go-runtime-events

Shared runtime-activity event schema for CLI-wrapped agent subprocesses
and other managed processes in the Hollis Labs portfolio.

This package owns the on-the-wire envelope (`Event`) and nothing else.
Per-kind payload shapes stay opaque (`json.RawMessage`) so the schema
can remain stable while individual apps and adapters evolve their own
payload conventions independently. The companion `go-agent-wrapper`
library produces these events; Nanite, Tether, Torque, Hadron, and Stack
Explorer consume them.

## Status (v0.1.0, 2026-05-26)

Production-shaped schema with:

- `Event` envelope: `schema_version`, `id`, `kind`, `time`, `app`,
  `session_id`, `turn_id`, `sequence`, `parent_id`, `raw_offset`,
  `process`, `source`, `payload`.
- 31 `EventKind` constants covering process / session / turn / stdio /
  agent / policy / plant / sandbox / interrupt lifecycle.
- 7 `SourceChannel` constants + 3 `Confidence` levels.
- `Sequencer` (per-session monotonic, concurrent-safe), ID generators
  with stable prefixes (`evt_`, `ses_`, `turn_`), `Emitter` with options
  (`WithID`, `WithTurnID`, `WithParentID`, `WithRawOffset`,
  `WithProcess`), and a thread-safe `SetProviderSessionID` mutator.
- `Sink` interface plus `SinkFunc`, `MultiSink` (fan-out with joined
  errors), and a reference `FileSink` (append-only JSONL with
  mutex-guarded writes).
- 24 tests, all `-race` clean.

See [ROADMAP.md](./ROADMAP.md) for deferred scope.

Module path: `github.com/hollis-labs/go-runtime-events`
Library package: `github.com/hollis-labs/go-runtime-events/runtimeevents`

## Install

```sh
go get github.com/hollis-labs/go-runtime-events/runtimeevents
```

## Producing events

```go
import (
    "context"
    "github.com/hollis-labs/go-runtime-events/runtimeevents"
)

em := &runtimeevents.Emitter{
    Sink:      mySink,                              // implements runtimeevents.Sink
    App:       "nanite",
    SessionID: runtimeevents.NewSessionID(),
    Process:   runtimeevents.Process{Provider: "claude", Runtime: "streaming-stdio"},
    Sequencer: runtimeevents.NewSequencer(),
}

_ = em.Emit(context.Background(),
    runtimeevents.KindAgentToolUse,
    runtimeevents.Source{Channel: runtimeevents.ChannelClaudeStreamJSON, Confidence: runtimeevents.ConfidenceExact},
    map[string]any{"tool": "Read", "args": map[string]any{"path": "/tmp/x"}},
    runtimeevents.WithTurnID("turn_1"),
)
```

## Consuming events

```go
type myStore struct{}
func (myStore) Write(_ context.Context, ev runtimeevents.Event) error {
    // Switch on ev.Kind; treat unknown kinds as opaque rather than rejecting.
    return nil
}
```

For a worked sink: `runtimeevents.OpenFileSink(path)` returns a
JSONL-append `Sink` that closes cleanly via `Close()`.

## Envelope shape

| Field | Required | Notes |
|---|---|---|
| `schema_version` | yes | Bump only on non-additive envelope changes; starts at `"1"`. |
| `id` | yes | `evt_<32-hex>`; use `NewEventID()`. |
| `kind` | yes | Open string; `EventKind` constants enumerate the kinds the wrapper emits today. |
| `time` | yes | RFC3339 timestamp. |
| `session_id` | yes | Wrapper-session identity. Distinct from `process.provider_session_id`. |
| `sequence` | yes | Per-session monotonic. Authoritative for ordering — wall-clock is informative only. |
| `source.channel` | yes | Observation transport. Open string; constants name the channels the wrapper supports today. |
| `app`, `turn_id`, `parent_id`, `raw_offset`, `payload`, `process.*`, `source.confidence` | no | Omitted from JSON when unset. |

## Layout

Module-root package, no `cmd/`, no `internal/`.

```
.
├── runtimeevents/   # Library package — importable by other modules
├── examples/        # Runnable usage examples (TBD)
├── go.mod
├── CHANGELOG.md
├── ROADMAP.md
└── README.md (this file)
```

## Architecture notes

- `chrispian/inbox/cli-runner-wrapper-architecture-2026-05-26.md`
- `chrispian/inbox/cli-wrapper-implementation-followups-2026-05-26.md`
  (handoff for the next session)

## Development

```sh
go test -race ./...   # tests
go vet ./...          # vet
gofmt -l .            # formatting check (no output = clean)
golangci-lint run     # lint
govulncheck ./...     # vulnerability scan
```

CI (`.github/workflows/check.yml`) runs the same checks on push and pull
request to `main`.

## License

MIT — see [LICENSE](./LICENSE).
