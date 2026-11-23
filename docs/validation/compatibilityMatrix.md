# Compatibility matrix

Honest status only. Do not mark real-env without host IME evidence.

| Target | Daemon IPC | Adapter | Smoke (host IME) | Notes |
| --- | --- | --- | --- | --- |
| macOS (IMKit) | implemented / compiled (UDS) | implemented / compiled (`make app` + SurTeam-signed DMG) | TIS listing/enable/select yes; Settings Add browser / typing no | DMG Installer → `/Library/Input Methods` + register/enable Hans-only + scrub parent prefs; Add-under-简体中文 UI still unconfirmed |
| Windows (TSF) | implemented in source (Named Pipe, windows build tag) | source implemented | not run | not MSVC-compiled on this Mac |
| Linux (Fcitx5) | implemented / compiled (UDS client lib) | source implemented | not run | Fcitx5 headers missing on this Mac |

Shared native client: `platforms/common` `libtsime_ipc` built here; `tsime_frame_smoke` passed.
