# Key input path data flow

**Status:** designed / documented — runtime **not yet implemented**.

## Happy path (key → candidates → commit)

```text
Host app
  → OS IME framework (IMKit / TSF / Fcitx5)
    → Platform adapter
      → IPC frame (KeyEvent / Update)
        → Daemon session
          → Engine (compose / convert / rank)
        ← CandidateList / Preedit / Commit
      ← IPC frame
    ← Adapter maps to host UI / commit string
  ← Text inserted / preedit updated
```

## Message kinds (conceptual)

| Direction | Examples |
| --- | --- |
| Adapter → daemon | Connect, CreateSession, KeyEvent, SelectCandidate, Clear, Disconnect |
| Daemon → adapter | Preedit, Candidates, Commit, Status, Error |

Exact protobuf schemas live with the ipc module once code lands (`internal/ipc` / proto files).

## Latency expectations

- Hot path must stay in-process-local IPC; no HTTP, no JSON parse on each key (see principles and ADR 0001).
- Engine work for a single key should be bounded; heavy dictionary loads happen at daemon start or explicit reload, not per key.

## Failure paths

- Transport drop → adapter surfaces disabled state to host; daemon drops or times out the session.
- Engine error → typed Error frame; adapter must not crash the host IME process.
