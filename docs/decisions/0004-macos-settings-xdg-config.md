# ADR 0004: macOS Settings window + XDG config persistence

- **Status:** Accepted
- **Date:** 2026-03-22

## Context

The macOS IMKit adapter already exposes CN/EN mode-toggle bindings via `ModeToggleHotkeys` and draws a custom candidate `NSPanel`. Users need a first-party way to change hotkeys, candidate column count, panel font, and highlight color without rebuilding. Host System Settings is a poor fit for IME-process-owned state, and preferences should remain portable across Unix-style layouts used by the rest of the project (`$XDG_*`).

## Decision

1. **UI entry:** The menu-bar IME menu gains a **Settings** item that opens a native `NSWindow` owned by `SettingsWindowController` (not a SwiftUI Preferences scene or System Settings pane).
2. **Settings surface:**
   - CN/EN mode-toggle hotkeys: multiple bindings; add/remove; values captured by recording user key input (same binding model as `ModeToggleHotkeys`).
   - Candidate display column count: integer 1–9; the adapter clamps the daemon page size to this UI limit when drawing.
   - Candidate panel font name and size.
   - Highlight color for the selected candidate.
3. **Persistence:** JSON at `$XDG_CONFIG_HOME/tsime/settings.json`, falling back to `~/.config/tsime/settings.json` when `XDG_CONFIG_HOME` is unset. `ConfigStore` creates the directory on first save/load.
4. **Lifecycle notifications:** `AppNotifier` posts local notifications when IMKit launches (initializing) and when `tsimed` IPC becomes reachable (ready). Authorization failure degrades silently (no prompt loop, no hard failure).
5. **Source ownership:** Expected under `platforms/macos/Sources/TSInputMethod/`: `ConfigStore.swift`, `SettingsWindowController.swift`, `AppNotifier.swift`, with wiring in `InputController.swift`, `CandidatePanel.swift`, `main.swift`, and the macOS `Makefile`.

## Consequences

- Mode-toggle bindings are no longer compile-time-only defaults; Settings can replace the list persisted in `settings.json`.
- Candidate panel appearance (columns, font, highlight) is adapter-local UI state, not a daemon IPC setting.
- Config path follows XDG on macOS even though Apple defaults to `~/Library/Preferences`; this keeps one documented layout with Linux and tooling.
- Notification permission is optional; missing authorization must not block IME startup or IPC.

## Alternatives considered

| Alternative | Why not |
| --- | --- |
| `UserDefaults` / AppKit preferences domain | Diverges from XDG layout used elsewhere; harder to inspect/edit as plain files |
| System Settings / Preference pane extension | Heavyweight for an Input Method; lifecycle tied to System Settings, not IMKit |
| Daemon-owned settings over IPC | UI chrome (font/color/columns) is adapter-side; hotkeys must work before IPC is up |
| Hard-coded hotkeys only | Blocks Karabiner / user remaps beyond the shipped Left Shift + F18 defaults |

## References

- [../platforms/macos/index.md](../platforms/macos/index.md)
- [0003-custom-candidate-panel.md](0003-custom-candidate-panel.md)
- [../validation/knownIssues.md](../validation/knownIssues.md) (KI-005)
