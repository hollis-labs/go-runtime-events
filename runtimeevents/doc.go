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
//   - Treat the KindPolicy* constants as legacy compatibility labels for
//     observed policy findings or recommendations. A policy event alone is not
//     evidence that the producer changed, blocked, or paused an operation.
//
// The wrapper assigns sequence numbers and IDs; downstream consumers should
// not regenerate them. See [Sequencer], [Emitter], and [Sink] for the
// producer-side helpers.
//
// # Payload conventions
//
// Payloads stay opaque to this package, but producers share a few field
// conventions so consumers can read every runtime the same way. All fields
// are optional and additive; a consumer that does not know a field ignores
// it.
//
//   - agent.delta: content (the text), block_id (stable for every delta of
//     one content block, different for the next block, so a consumer can
//     separate blocks without knowing the provider), phase ("thought" for
//     thinking text on every runtime; "narration" or "final" when a native
//     provider classifies its text; "message" for an ACP agent message;
//     absent when nothing classified it).
//   - turn.completed and turn.failed: usage (the turn's accumulated usage;
//     native runtimes emit go-llm-types' Usage with its Go field names,
//     InputTokens, OutputTokens, CacheCreationTokens, CacheReadTokens,
//     StopReason and CostUSD, a per-turn figure consumers may sum across
//     turns; ACP runtimes pass the agent's own usage object through),
//     stop_reason (normalised: end_turn, max_tokens, tool_use, turn_limit,
//     refusal, cancelled or error; a value outside that list is the
//     provider's own word, passed through), error (turn.failed).
//   - session.lost: requested_id, actual_id, reason.
//   - session.auth_failed: error.
//   - agent.permission_denied: action, display_name.
//
// These names are chosen so a chat-stream consumer can map them directly:
// block_id to a message part id, stop_reason to a finish reason.
package runtimeevents
