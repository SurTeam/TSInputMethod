# Validation index

Honest status only. Do not mark unrun tests as passed. Distinguish implemented / compiled / auto-tested / real-env.

## Legend

See [../index.md](../index.md) for implemented / compiled / auto-tested / real-env.

## Component snapshot

| Component | implemented | compiled | auto-tested | real-env |
| --- | --- | --- | --- | --- |
| Documentation foundation | yes | N/A | N/A | N/A |
| IPC ADR / Proto | yes | yes | N/A | N/A |
| Go daemon (`cmd/imed`) | yes | yes | smoke start | no |
| `internal/ipc` UDS | yes | yes | yes | no |
| `internal/ipc` Named Pipe | yes (windows tags) | not on this host | no | no |
| `platforms/common` libtsime_ipc | yes | yes (`cmake --build`) | frame smoke ok | no |
| `internal/session` / dictionary / pinyin | yes | yes | yes | no |
| Fogong / rime-ice dictBuilder | yes | N/A | build ran | no |
| macOS IMKit adapter | yes | yes (`make app`) | no | **no** |
| Windows TSF adapter | yes (source + scripts) | **no** (needs MSVC) | no | **no** |
| Linux Fcitx5 adapter | yes (source + scripts) | **no** (needs Fcitx5) | no | **no** |
| Ranking / Inference | yes (API + FastRanker + stubs) | yes | yes (unit) | no |
| Xiaohe / flypy shuangpin | yes | yes | yes (unit) | no |
| macOS trial package (`scripts/package_macos.sh`) | yes (DMG OOBE + SurTeam codesign) | yes when packaged | codesign verify | user installs; real-env typing **no** |

## Observed packaging / signing

- Primary artifact: `dist/macos/TSInputMethod-macos.dmg` with double-click **Install TSInputMethod**
- Installer embeds `imed` + `rime-ice.tsdict`, installs into `~/Library/Input Methods`, LaunchAgent auto-starts daemon with `TSIME_SCHEME=xiaohe`
- Codesign identity resolved: `SurTeamCode` (self-signed; not Apple Developer ID Application)
- `codesign --verify` passed for IME app, embedded `imed`, and Installer; `spctl --assess` rejected app/DMG
- Notarization not claimed; Gatekeeper may still warn for downloaded artifacts
- Zip remains secondary fallback only
- Host Input Source enable / TextEdit typing still **real-env: no**

## Pages

- [compatibilityMatrix.md](compatibilityMatrix.md)
- [performance.md](performance.md)
- [knownIssues.md](knownIssues.md)
