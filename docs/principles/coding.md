# Coding principles

1. **No comments in handwritten source.** Names, types, and `docs/` carry intent.
2. **Code is truth.** Update docs when behavior changes; do not leave contradictory prose.
3. **Clear package boundaries.** `internal/ipc`, `internal/session`, `internal/engine` stay separable; adapters depend on IPC contracts, not engine internals.
4. **Explicit errors.** Return and propagate typed errors on the IPC and Go APIs; do not swallow failures on the hot path.
5. **Generated code is owned by generators.** Edit `.proto` / schemas, not generated stubs, unless the task is generator repair.
