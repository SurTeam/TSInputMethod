# Platform compatibility notes

**Status:** adapters implemented in-tree; macOS packaged/compiled on this host; Windows/Linux source present; real-env typing still pending.

## Shared contract

All platforms speak the same protobuf + framed IPC dialect. Compatibility means:

- Same message semantics for keys, preedit, candidates, and commit.
- Transport differs (UDS vs Named Pipe) but framing does not.
- Engine behavior is OS-agnostic; only adapters interpret host key codes / modifier maps.
- Edit `key_code` values are macOS CG virtual keys (Delete=51, Escape=53, Space=49, ForwardDelete=117). Printable input uses `characters` ([ADR 0002](../decisions/0002-semantic-edit-keycodes.md)).

## macOS (InputMethodKit)

- Adapter under `platforms/macos`.
- Must respect IMKit server lifecycle and text input client callbacks.
- Host CG virtual keys follow Karabiner-Elements `pqrs::osx::cg_event::key_code` and are forwarded unchanged.
- Sandbox / TCC constraints may affect socket path placement; prefer per-user runtime directories.

## Windows (TSF)

- Planned adapter under `platforms/windows`.
- TSF text service and UI thread rules apply; heavy work stays in the daemon.
- Named Pipe security descriptors must limit access to the interactive user.

## Linux (Fcitx5)

- Planned addon under `platforms/linux`.
- Follow Fcitx5 addon ABI and event loop conventions.
- UDS under `$XDG_RUNTIME_DIR` when available.

## Version skew

- Adapters and daemon advertise protocol version at Connect.
- Older adapters against newer daemons: allow only if the daemon keeps a documented compatibility window.
- Breaking IPC changes require a new ADR and version bump.

## Tracking

Live matrix: [../validation/compatibilityMatrix.md](../validation/compatibilityMatrix.md).
