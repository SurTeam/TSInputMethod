# Known issues

| ID | Area | Summary | Status |
| --- | --- | --- | --- |
| KI-001 | dictionary | Optional `ext` rime-ice table not built by default (size/time); 8105+base is the default index | open |
| KI-002 | platforms | Host IME real-env validation pending for all three adapters; Windows/Linux not compiled on this Mac | open |
| KI-003 | ranking | Real neural backends / weights not shipped; FastRanker + injectable Backend + no-model degrade landed | open (models) |
| KI-004 | windows-ipc | Named Pipe source landed (windows tags) but not executed on a Windows host | open |
| KI-005 | macos-imkit | SurTeam-signed DMG OOBE + ComponentInputModeDict/TIS register+enable landed; parent/Hans TIS-enabled and selectable Hans can become current; System Settings Add-under-简体中文 browser / TextEdit typing still pending | open (UI/typing) |
| KI-007 | macos-signing | SurTeamCode is self-signed (not Developer ID); `spctl` rejects; Gatekeeper may warn; notarization not claimed; hardened-runtime build needs library-validation entitlement for Homebrew protobuf | open |
| KI-006 | shuangpin | Xiaohe map covers common finals/zero-initials; rare edge syllables may need iteration after user trial | open |
| KI-008 | macos-install | Installer now targets `/Library/Input Methods` and prompts for admin; user-level `~/Library/Input Methods` alone was insufficient for stable TIS visibility in this environment | open (document) |

Notes:

- Daemon and continuous pinyin decode are implemented and unit-tested; not yet verified inside host IME UIs.
- Upstream `base.dict.yaml` header comments contain literal `...` substrings; dictBuilder splits only on a whole-line `...` marker.
