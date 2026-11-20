# Security principles

1. **Local IPC only.** No network listeners on the input hot path.
2. **Same-user endpoints.** UDS and Named Pipe permissions restrict access to the interactive user session.
3. **Minimize data retention.** Do not log raw keystreams or committed passwords; prefer counts and error codes.
4. **Least privilege adapters.** Platform adapters request only the host capabilities required for IME duties.
5. **Treat IPC as a trust boundary.** Validate frame lengths and protobuf sizes; reject oversize or malformed frames.
6. **Personalization stays local.** User dictionary and SGD weights live under the OS user config directory with `0700` dirs and `0600` files. Persistence stores weights and user entries only—never raw context training logs.
7. **PrivacyMode and clear/reset.** Explicit PrivacyMode disables learning enqueue and fused personalization. Operators can clear personal weights, user-dictionary entries, and short learning context. Feature hashing is not encryption.
8. **Secure-input limits (honest).** Platform adapters do not currently expose a reliable “secure/password field” signal into the daemon. Do **not** claim automatic filtering of all sensitive fields. Until that exists, users should enable PrivacyMode or clear personal data after sensitive sessions.
9. **Learning-context boundaries.** IME deactivate / app focus loss and CN↔EN mode switches invalidate session learning context so unrelated prior text is not mixed. Post-commit in-app Backspace and other external buffer edits are still largely unobservable; treat accurate Backspace-aware committed-context tracking as a follow-up, not a present guarantee.
