# go-runtime-events Roadmap

Status as of v0.1.0 (2026-05-26). See
[CHANGELOG.md](./CHANGELOG.md) for what landed.

## Publish blockers

- None internal to this module — `go-runtime-events` has no Hollis Labs
  dependencies. It's the leaf of the dependency tree and should be the
  first tagged of the three new libs (`go-runtime-events` →
  `go-harness-filters` → `go-agent-wrapper`).
- Standard pre-tag polish: `examples/` is empty, drop a runnable
  example before publishing.

## Deferred this pass

### Reference sinks

- `FileSink` is the only reference Sink implementation. Useful
  additions, in priority order:
  - **`RingSink`** — in-memory ring buffer for tests and live-tailing
    UIs.
  - **`HTTPSink`** — POST each event to a URL with retry/backoff.
    Useful for Tether's `/events/stream` ingestion or any centralized
    sink.

  Add these in a follow-up rather than now — the architecture doc
  explicitly avoids prescribing durability, so reference impls should
  only land when there's a real consumer.

### Schema extensions landed

- **`KindPolicyApprovalRequested`** — added so `policy.ModeApproval`
  no longer overloads `policy.block`. The full pause/resume operator
  approval flow still belongs in the wrapper/app layer.

### Schema extensions on the table

- **`session.idle` / `session.processing` / `session.heartbeat` payload
  shapes** — kinds exist, but no concrete payload conventions yet.
  Drop a documented shape (last-activity-time, processing token,
  heartbeat counter) once the wrapper starts emitting them.

## Open design questions

1. **Per-kind payload validation.** Right now `payload` is opaque. If
   too many consumers grow ad-hoc shapes for the same kind, the
   wrapper schema effectively splinters. Consider standardizing payload
   shapes per-kind (still as `json.RawMessage` on the wire but with a
   companion `runtimeevents/payloads` subpackage that owns the Go
   types). Wait until at least two consumers diverge before paying that
   tax.
2. **`WithID` safety.** Lets callers override the auto-generated event
   ID — necessary for some `ParentID` correlation chains, but duplicate
   IDs break replay. `Emitter.EmitReturning` now covers the safer common
   case where callers only need the assigned ID after emission. Keep
   `WithID` for pre-generated correlation IDs and document sparingly.
3. **Schema versioning policy.** `SchemaVersion = "1"` — when do we
   bump? Adding kinds, adding source channels, adding optional fields
   are all backwards-additive. Removing or renaming anything breaks
   consumers and warrants a bump. Document this in the package doc
   before v1.

## Related docs

- Original architecture: `chrispian/inbox/cli-runner-wrapper-architecture-2026-05-26.md`
- Next-session handoff: `chrispian/inbox/cli-wrapper-implementation-followups-2026-05-26.md`
