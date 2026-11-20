# Validation index

Honest status only. Do not mark unrun tests as passed. Distinguish implemented / compiled / auto-tested / real-env.

## Legend

See [../index.md](../index.md) for implemented / compiled / auto-tested / real-env.

## Component snapshot

| Component | implemented | compiled | auto-tested | real-env |
| --- | --- | --- | --- | --- |
| Documentation foundation | yes | N/A | N/A | N/A |
| IPC ADR / Proto | yes | yes | N/A | N/A |
| Go daemon (`cmd/tsimed`) | yes | yes | smoke start | no |
| `internal/ipc` UDS | yes | yes | yes | no |
| `internal/ipc` Named Pipe | yes (windows tags) | not on this host | no | no |
| `platforms/common` libtsime_ipc | yes | yes (`cmake --build`) | frame smoke ok | no |
| `internal/session` / dictionary / pinyin | yes | yes | yes | no |
| Fogong / rime-ice dictBuilder | yes | N/A | build ran (8105+41448+base+ext+tencent+others + emoji.txt) | no |
| macOS IMKit adapter | yes | yes (`make app`) | TIS list/enable after register | listing+enable yes; Settings Add browser / typing **no** |
| Windows TSF adapter | yes (source + scripts) | **no** (needs MSVC) | no | **no** |
| Linux Fcitx5 adapter | yes (source + scripts) | **no** (needs Fcitx5) | no | **no** |
| Ranking / Inference | yes (API + FastRanker + stubs) | yes | yes (unit) | no |
| Online SGD personalization + user dict | yes | yes | yes (`go test ./...`; `-race` on personal/ranking/engine/session/dictionary/pinyin) | no |
| `cmd/tsime-personal-eval` | yes | yes | fixture replay smoke | no |
| Xiaohe / flypy shuangpin | yes | yes | yes (unit) | no |
| macOS trial package (`scripts/package_macos.sh`) | yes (DMG OOBE + SurTeam codesign + TIS register/enable) | yes when packaged | codesign verify + TIS parent/Hans enable + select | Settings Add browser / typing **no** |

## Observed packaging / signing

- Primary artifact: `dist/macos/TSInputMethod-macos.dmg` with double-click **Install TSInputMethod**
- Installer embeds `tsimed` + `rime-ice.tsdict`, installs into `/Library/Input Methods` (admin), runs `TISRegisterInputSource`, LaunchAgent auto-starts daemon with `TSIME_SCHEME=xiaohe`
- `Info.plist` includes `ComponentInputModeDict` Hans mode (`com.surteam.inputmethod.TSInputMethod.Hans`, `zh-Hans`)
- Observed after register/enable: parent + Hans both TIS-enabled; Hans selectable/`zh-Hans`; appears in enabled-only list; `TISSelectInputSource` can make Hans current; prefs record present; System Settings Add-under-简体中文 browser / TextEdit typing still pending
- Codesign identity resolved: `SurTeamCode` (self-signed; not Apple Developer ID Application)
- `codesign --verify` passed for IME app, embedded `tsimed`, and Installer; `spctl --assess` rejected app/DMG
- Notarization not claimed; Gatekeeper may still warn for downloaded artifacts
- Zip remains secondary fallback only

## Pages

- [compatibilityMatrix.md](compatibilityMatrix.md)
- [performance.md](performance.md)
- [knownIssues.md](knownIssues.md)
