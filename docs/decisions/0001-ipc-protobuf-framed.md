# ADR 0001: Protocol Buffers + length-prefixed binary frames over UDS / Named Pipe

- **Status:** Accepted
- **Date:** 2026-03-21

## Context

TSInputMethod needs a cross-platform hot path between thin OS adapters and a shared Go daemon. Keystroke latency, local-only trust, and a single schema across macOS, Windows, and Linux matter more than browser-friendly debugging.

## Decision

1. **Payload:** Protocol Buffers for request/response and event messages.
2. **Framing:** Length-prefixed binary frames (fixed-width length header + protobuf body).
3. **Transport:**
   - macOS / Linux: Unix domain sockets (UDS)
   - Windows: Named Pipes
4. **Reject on the hot path:** HTTP and JSON (including gRPC-over-HTTP/2 style stacks that add unnecessary overhead and complexity for local keystroke IPC).

## Consequences

- One schema and codec shared by all adapters.
- Adapters stay thin; host frameworks never dictate the wire format.
- Debugging requires protobuf-aware tools rather than curl/browser inspection.
- Frame size limits and credential checks are mandatory at the transport boundary.

## Alternatives considered (future re-evaluation)

| Alternative | Why not now | Revisit when |
| --- | --- | --- |
| HTTP + JSON | Higher per-key parse cost; easy to accidentally expose a network listener; poor fit for keystroke IPC | Never for hot path; maybe for out-of-band admin only |
| FlatBuffers | Excellent zero-copy reads; different ecosystem/tooling tradeoffs | Benchmarks show protobuf encode/decode dominating measured latency |
| Cap’n Proto | Strong RPC/performance story; smaller community in this stack | Benchmarks and maintenance cost justify a migration ADR |

Any switch requires new benchmarks under `tests/performance` and a superseding ADR.

## References

- [../architecture/processModel.md](../architecture/processModel.md)
- [../modules/ipc/index.md](../modules/ipc/index.md)
