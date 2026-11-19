# Process and IPC transport model

**Status:** designed / documented — runtime **not yet implemented**.

## Processes

1. **Daemon (`imed`)** — long-lived user-session service. Owns engine state, dictionaries, and session table.
2. **Adapter process(es)** — one per platform integration surface (IMKit server, TSF text service, Fcitx5 addon). Translate host callbacks into IPC messages.

Multiple adapter clients may attach to one daemon. Each attachment maps to one or more **sessions** (see [lifecycle.md](lifecycle.md)).

## Transport

| Platform | Transport | Endpoint convention (planned) |
| --- | --- | --- |
| macOS | Unix domain socket (UDS) | Per-user runtime path under the login session |
| Linux | Unix domain socket (UDS) | Per-user runtime path (e.g. `$XDG_RUNTIME_DIR`) |
| Windows | Named Pipe | Per-user pipe name scoped to the interactive session |

## Framing

- Protocol Buffers payloads.
- Length-prefixed binary frames (fixed-width length header + message body).
- Request/response and optional push events share the same frame format; message type is inside the protobuf envelope.

Details and rejected alternatives: [ADR 0001](../decisions/0001-ipc-protobuf-framed.md).

## Trust boundary

IPC is **local-only**. Adapters and daemon assume same-user endpoints. Authentication relies on OS socket/pipe credentials and filesystem/pipe ACL; there is no network listener on the hot path.
