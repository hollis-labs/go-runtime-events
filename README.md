# go-runtime-events

Shared runtime-activity event schema for CLI-wrapped agent subprocesses
and other managed processes in the Hollis Labs portfolio.

This package owns the on-the-wire envelope (`Event`) and nothing else.
Per-kind payload shapes stay opaque (`json.RawMessage`) so the schema
can remain stable while individual apps and adapters evolve their own
payload conventions independently. The companion `go-agent-wrapper`
library produces these events; Nanite, Tether, Torque, Hadron, and Stack
Explorer consume them.

## Status (v0.1.2, 2026-09-04)

Production-shaped schema with:

- `Event` envelope: `schema_version`, `id`, `kind`, `time`, `app`,
  `session_id`, `turn_id`, `sequence`, `parent_id`, `raw_offset`,
  `process`, `source`, `payload`.
- 32 `EventKind` constants covering process / session / turn / stdio /
  agent / policy / plant / sandbox / interrupt lifecycle.
- 7 `SourceChannel` constants + 3 `Confidence` levels.
- `Sequencer` (per-session monotonic, concurrent-safe), ID generators
  with stable prefixes (`evt_`, `ses_`, `turn_`), `Emitter` with
  `EmitReturning` and options (`WithID`, `WithTurnID`, `WithParentID`,
  `WithRawOffset`, `WithProcess`), plus thread-safe `SetProcess` and
  `SetProviderSessionID` mutators.
- `Sink` interface plus `SinkFunc`, `MultiSink` (fan-out with joined
  errors), and a reference `FileSink` (append-only JSONL with
  mutex-guarded writes).
- 27 tests, all `-race` clean.

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

The action-shaped `KindPolicyNudge`, `KindPolicyRewrite`,
`KindPolicyBlock`, and `KindPolicyApprovalRequested` names and their
`policy.*` wire values are legacy compatibility labels. They carry
observed policy findings or recommendations; receiving one does not by
itself prove that the producer changed, blocked, or paused the underlying
operation.

## Payload conventions

Payloads stay opaque to this module, but producers share these field
conventions (all optional and additive) so consumers read every runtime
the same way. The package doc carries the same list.

| Kind | Fields |
|---|---|
| `agent.delta` | `content`; `block_id` (stable within one content block, different for the next, so blocks can be separated without per-provider rules); `phase` (`thought` for thinking text on every runtime; `narration` / `final` when a native provider classifies its text; `message` for an ACP agent message; absent otherwise) |
| `turn.completed`, `turn.failed` | `usage` (native runtimes: go-llm-types `Usage` with its Go field names — `InputTokens`, `OutputTokens`, `CacheCreationTokens`, `CacheReadTokens`, `StopReason`, `CostUSD`; ACP runtimes: the agent's own usage object); `stop_reason` (`end_turn`, `max_tokens`, `tool_use`, `turn_limit`, `refusal`, `cancelled`, `error`, or the provider's own word when it is none of those); `error` (`turn.failed`) |
| `turn.completed` | also `text`: the turn's final message when the runtime reports one on its terminal event (Claude's `result.result`, agy's `result.response`); absent when it reports none |
| `turn.failed` | also `reason`: `interrupted` (the host cancelled the turn) or `process_exited` (the process went away mid-turn) when the producer knows; absent for an ordinary failure |
| `agent.tool_use` | `tool_use` (`id`, `name`, `input`) on native runtimes; ACP runtimes carry the tool call's own fields (`tool_call_id`, `title`, `kind`, `status`, `raw_input`) |
| `agent.permission_requested`, `agent.permission_resolved` | request: `request_id` (when the agent has one), `method`, `params`; resolution: `allowed`, `reason`, `request_id`, and `ParentID` set to the request event's ID, so the pair matches either way |
| `session.lost` | `requested_id`, `actual_id`, `reason` |
| `session.auth_failed` | `error` |
| `agent.permission_denied` | `action`, `display_name` |

`block_id` and `stop_reason` are named so a chat-stream consumer can map
them straight onto a message part id and a finish reason.

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
