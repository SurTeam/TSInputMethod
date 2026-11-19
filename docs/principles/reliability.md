# Reliability principles

1. **Isolate faults.** Adapter crashes should not corrupt daemon state; daemon restarts should leave adapters able to reconnect.
2. **Session hygiene.** Destroy sessions on disconnect; no leaked compose buffers across clients.
3. **Degrade gracefully.** If the daemon is unavailable, disable or soft-fail in the host IME rather than hang.
4. **Idempotent control ops** where practical (Clear, DestroySession).
5. **Observable failures.** Errors cross IPC as typed messages; logs include session and client identifiers without leaking keystroke contents by default.
