# Agent protocol

## First read

1. Read [docs/index.md](docs/index.md) before changing code or docs.
2. Follow linked architecture, module, platform, principle, and decision pages that apply to the task.
3. Prefer current repository files over memory of earlier chats.

## Truth and comments

- **Code is truth.** Behavior, APIs, and layouts come from source and verified builds/tests, not from stale prose.
- **No comments in handwritten source.** Do not add `//`, `/* */`, `#`, or equivalent comments in project-owned source. Intent belongs in `docs/` and clear names.
- Generated or vendored code may keep upstream comments; do not strip third-party files unless a task requires it.

## Documentation home

- Technical documentation lives only under `docs/`.
- Root `README.md` is a short intro and navigation pointer, not a second architecture document.
- Architecture Decision Records live under `docs/decisions/`.
- Validation status lives under `docs/validation/` and must not invent pass results.

## Commit-time documentation review checklist

Before finishing a change set:

- [ ] `docs/index.md` still points to every topic the change affects.
- [ ] Module and platform pages list accurate goals, boundaries, and source paths.
- [ ] Architecture pages match process/IPC/lifecycle behavior if those changed.
- [ ] New durable choices have an ADR under `docs/decisions/` when warranted.
- [ ] Verification status (implemented / compiled / auto-tested / real-env) is updated honestly; never mark auto-tested or real-env without evidence.
- [ ] Handwritten source has no new comments.
- [ ] Directory keepers remain for empty planned trees that should stay in git.

## Task completion report format

When a task is done, report in this shape:

```text
## Summary
<one short paragraph of what changed and why>

## Files
- path — one-line purpose

## Verification
- implemented: <yes/no + what>
- compiled: <yes/no + command, or N/A>
- auto-tested: <yes/no + command/result, or N/A>
- real-env: <yes/no + environment, or N/A>

## Docs touched
- list of docs paths updated or confirmed still accurate

## Follow-ups
- remaining work, risks, or blockers (or "none")
```

## Scope discipline

- Do not invent fake test or real-environment success.
- Do not implement unrelated modules "while here" unless the task asks for them.
- Platform adapters and Go runtime work are owned by their own tasks; docs agents must not claim those are implemented without evidence.
