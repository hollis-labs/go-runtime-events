package runtimeevents

// EventKind identifies the kind of runtime event. Consumers must tolerate
// unknown kinds (treat them as opaque) rather than rejecting the [Event] —
// new kinds will be added over time and old consumers must round-trip them
// cleanly.
type EventKind string

// Process-lifetime kinds.
const (
	KindProcessStarted EventKind = "process.started"
	KindProcessExited  EventKind = "process.exited"
)

// Session-lifetime kinds. A session spans multiple turns.
const (
	KindSessionReady      EventKind = "session.ready"
	KindSessionIdle       EventKind = "session.idle"
	KindSessionProcessing EventKind = "session.processing"
	KindSessionHeartbeat  EventKind = "session.heartbeat"
)

// Turn kinds. A turn is one request/response exchange within a session.
const (
	KindTurnStarted   EventKind = "turn.started"
	KindTurnCompleted EventKind = "turn.completed"
	KindTurnFailed    EventKind = "turn.failed"
)

// Raw and line-buffered IO kinds. Raw kinds carry exact byte spans; line
// kinds carry one logical line each. Both may be emitted for the same
// underlying bytes when the wrapper has both byte-level and line-level
// observers attached.
const (
	KindStdinWrite EventKind = "stdin.write"
	KindStdoutRaw  EventKind = "stdout.raw"
	KindStderrRaw  EventKind = "stderr.raw"
	KindStdoutLine EventKind = "stdout.line"
	KindStderrLine EventKind = "stderr.line"
)

// Agent semantic kinds. These require a semantic observation channel
// (stream-json, JSON-RPC, plugin hook) — PTY-only sessions cannot emit
// them reliably.
const (
	KindAgentDelta               EventKind = "agent.delta"
	KindAgentToolUse             EventKind = "agent.tool_use"
	KindAgentToolResult          EventKind = "agent.tool_result"
	KindAgentSubagentSpawn       EventKind = "agent.subagent_spawn"
	KindAgentPermissionRequested EventKind = "agent.permission_requested"
	KindAgentPermissionResolved  EventKind = "agent.permission_resolved"
)

// Policy action kinds. Emitted whenever the wrapper's policy engine takes
// a non-observe action against a command, tool call, or output.
const (
	KindPolicyNudge   EventKind = "policy.nudge"
	KindPolicyRewrite EventKind = "policy.rewrite"
	KindPolicyBlock   EventKind = "policy.block"
)

// Planting kinds. Emitted around boot-dir / hook / plugin / MCP-config
// planting before and after the wrapped process starts.
const (
	KindPlantStarted   EventKind = "plant.started"
	KindPlantCompleted EventKind = "plant.completed"
)

// Sandbox kind. Emitted once after the sandbox profile has been applied
// (or refused) just before exec.
const KindSandboxApplied EventKind = "sandbox.applied"

// Interrupt kinds. Pair via [Event.ParentID] — the acknowledged event
// references the requested event's ID.
const (
	KindInterruptRequested    EventKind = "interrupt.requested"
	KindInterruptAcknowledged EventKind = "interrupt.acknowledged"
)

// SourceChannel identifies the observation transport that produced the
// event. Open string — adapters may introduce new channels without a
// schema bump; consumers must tolerate unknown channels.
type SourceChannel string

const (
	ChannelClaudeStreamJSON SourceChannel = "claude-stream-json"
	ChannelOpenCodePlugin   SourceChannel = "opencode-plugin"
	ChannelJSONRPC          SourceChannel = "jsonrpc"
	ChannelPTY              SourceChannel = "pty"
	ChannelStdio            SourceChannel = "stdio"
	ChannelHook             SourceChannel = "hook"
	ChannelFilter           SourceChannel = "filter"
)

// Confidence describes how directly the source observation maps to the
// event's semantic meaning. "exact" means the channel reports the event
// natively (e.g., a tool_use JSON message). "derived" means the wrapper
// reconstructed the event from a structured-but-indirect source.
// "inferred" means a text classifier or heuristic produced it; never
// authoritative for policy enforcement on its own.
type Confidence string

const (
	ConfidenceExact    Confidence = "exact"
	ConfidenceDerived  Confidence = "derived"
	ConfidenceInferred Confidence = "inferred"
)
