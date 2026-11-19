# Security principles

1. **Local IPC only.** No network listeners on the input hot path.
2. **Same-user endpoints.** UDS and Named Pipe permissions restrict access to the interactive user session.
3. **Minimize data retention.** Do not log raw keystreams or committed passwords; prefer counts and error codes.
4. **Least privilege adapters.** Platform adapters request only the host capabilities required for IME duties.
5. **Treat IPC as a trust boundary.** Validate frame lengths and protobuf sizes; reject oversize or malformed frames.
