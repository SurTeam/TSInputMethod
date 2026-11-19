# IPC Framing and Envelope

## Transport

- macOS / Linux: Unix domain socket (`internal/ipc.ListenUnix`, native client `connectPath`)
- Windows: Named Pipe (`internal/ipc.ListenNamedPipe` / `DialNamedPipe`, windows build tags; daemon `cmd/tsimed/main_windows.go`)
- Hot path: Protocol Buffers payload only; no HTTP/JSON

## Frame layout

```text
| 4-byte big-endian length | protobuf Envelope bytes |
```

- Maximum payload size: 4 MiB
- Empty frames are rejected

## Native client

- `platforms/common/native` C++ client + `tsime_c_api.h` for adapters
- Built as `libtsime_ipc` via `platforms/common/CMakeLists.txt`
- Smoke: `tsime_frame_smoke` (UDS socketpair frame roundtrip)

## Envelope fields

| Field | Role |
| --- | --- |
| `request_id` | Correlate request/response |
| `sequence_number` | Client ordering hint |
| `generation` | Composition/candidate generation for stale discard |
| `payload` | oneof Hello, session, key, candidate, commit, error |

## Negotiation

1. Client sends `Hello` with major/minor.
2. Server replies `HelloAck`; major mismatch rejects further RPCs.
3. Client opens a session, then sends key/mode/candidate messages.

## Source

- Protocol: `proto/ime/v1/ime.proto`
- Framing/server: `internal/ipc/`
- Native client: `platforms/common/`
- Daemon entry: `cmd/tsimed/`
