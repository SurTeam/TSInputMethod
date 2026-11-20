# ADR 0005: macOS daemon login-time lifecycle and soft ensure

- Status: Accepted
- Date: 2026-03-21

## Context

`tsimed` is owned by LaunchAgent `com.surteam.tsimed` with `RunAtLoad` and `KeepAlive`, but the IMKit adapter previously treated daemon readiness as an IME-controller concern: ensure/kickstart lived on Restart and always preferred `kickstart -k`, while readiness probing opened a full session from `TSInputController` init. That made initialization feel gated by switching to the input method or typing, and a healthy daemon could be killed during ensure.

## Decision

1. Login/session start remains the LaunchAgent path (`RunAtLoad` + `KeepAlive`). Dictionary/engine warm happens when `tsimed` starts, not when the user first types.
2. Adapter-side LaunchAgent management splits into:
   - soft ensure: bootstrap/enable if unloaded, then `kickstart` without `-k` only when IPC is unreachable
   - force restart: `kickstart -k` (with bootstrap fallback) only for menu **Restart** recovery
3. Soft ensure and hello-only reachability probing run from IMKit `main` as a safety net/notification path. Controller activation no longer owns first init.
4. Switching away from TSInputMethod tears down only adapter client/session state. The LaunchAgent-managed daemon and its loaded dictionary/engine stay alive.

## Consequences

- Healthy daemons are no longer force-killed by ordinary ensure.
- Ready notifications can fire without opening a typing session.
- Login readiness still depends on the LaunchAgent remaining loaded after install; IMKit itself still launches when the input source is used.
