# Module: ipc

**Status:** designed / documented — code **not yet implemented**.

## Goals

- Define the protobuf message envelope and length-prefixed frame codec.
- Provide daemon-side accept/serve and adapter-side dial/client helpers.
- Negotiate protocol version at Connect; enforce max frame size.

## Boundaries

- **In scope:** transport glue (UDS / Named Pipe), framing, encode/decode, connection lifecycle hooks.
- **Out of scope:** compose/convert logic (engine), session policy beyond connection multiplexing, host IME APIs.

## Source locations (planned)

| Path | Role |
| --- | --- |
| `cmd/imed` | Opens IPC listener via this module |
| `internal/ipc` | Server, client, frame codec |
| proto schemas (TBD under `internal/ipc` or `api/`) | Message definitions |

## Related

- [ADR 0001](../../decisions/0001-ipc-protobuf-framed.md)
- [processModel.md](../../architecture/processModel.md)
