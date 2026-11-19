# Cross-platform principles

1. **Shared semantics, local transport.** Same protobuf contract; UDS on macOS/Linux, Named Pipe on Windows.
2. **Adapters own host quirks.** Keycode maps, modifier semantics, and UI chrome stay outside the engine.
3. **Feature parity by contract.** New engine features land behind IPC messages available to all platforms before platform-specific UI polish.
4. **Compatibility matrix is authoritative.** Track per-OS status in `docs/validation/compatibilityMatrix.md`.
5. **No `#ifdef` engine logic.** Platform conditionals belong in adapters or thin transport glue, not conversion core.
