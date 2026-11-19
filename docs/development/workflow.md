# Development workflow

1. Read `docs/index.md` and the module/platform pages you will touch.
2. Implement behind existing boundaries (`internal/ipc`, `internal/session`, `internal/engine`, `platforms/...`).
3. Keep handwritten source comment-free; update `docs/` for intent and status.
4. Run through [documentationReview.md](documentationReview.md) before commit.
5. Report completion using the format in [../../Agent.md](../../Agent.md).

## Parallel ownership

- Docs foundation and Go daemon/adapters may land from different agents/tasks.
- Do not claim another owner's component is implemented, compiled, or tested without evidence in-tree.
