# Architecture overview

**Status:** designed / documented — runtime **not yet implemented**.

## Shape

TSInputMethod separates a shared **Go daemon** from thin **platform adapters**:

| Layer | Responsibility |
| --- | --- |
| Platform adapter | Host IME API only: macOS InputMethodKit, Windows Text Services Framework (TSF), Linux Fcitx5 |
| Unified binary IPC | Length-prefixed Protocol Buffers frames over UDS (macOS/Linux) or Named Pipe (Windows) |
| Go daemon (`tsimed`) | Sessions, engine, dictionaries, configuration, logging |

Adapters never own conversion logic. The daemon never calls OS IME frameworks directly.

## Why this split

- One engine implementation across three host stacks.
- Crash isolation: adapter process failures should not always take down the daemon (and vice versa where the OS allows).
- Hot-path IPC stays binary and local; no HTTP/JSON on keystroke path (see [ADR 0001](../decisions/0001-ipc-protobuf-framed.md)).

## Related

- [processModel.md](processModel.md)
- [dataFlow.md](dataFlow.md)
- [lifecycle.md](lifecycle.md)
- [compatibility.md](compatibility.md)
