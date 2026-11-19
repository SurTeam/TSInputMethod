# Engineering principles

1. **One shared core.** Conversion logic lives in the Go daemon; platforms are adapters only.
2. **Small durable interfaces.** Prefer stable IPC and module boundaries over leaky cross-imports.
3. **Evidence over aspiration.** Docs and validation pages record what exists and what was measured.
4. **Incremental delivery.** Ship vertical slices (IPC → session → engine → one adapter) rather than unfinished sprawl.
5. **Operability.** Failures are typed, logged, and recoverable at the adapter boundary without crashing host apps.
