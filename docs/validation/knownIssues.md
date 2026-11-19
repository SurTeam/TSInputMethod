# Known issues

| ID | Area | Summary | Status |
| --- | --- | --- | --- |
| KI-001 | dictionary | Default index now includes 8105+41448+base+ext+tencent+others; package size/memory higher than the old 8105+base build | closed |
| KI-002 | platforms | Host IME real-env validation pending for all three adapters; Windows/Linux not compiled on this Mac | open |
| KI-003 | ranking | Real neural backends / weights not shipped; FastRanker + injectable Backend + no-model degrade landed | open (models) |
| KI-004 | windows-ipc | Named Pipe source landed (windows tags) but not executed on a Windows host | open |
| KI-005 | macos-imkit | SurTeam-signed DMG OOBE + ComponentInputModeDict/TIS register+enable landed; custom `NSPanel` candidate row (up to 9 content-sized columns, tighter padding, darker 1–9 labels, wheel paging, `selected_index` highlight; ADR 0003) replaces IMKCandidates; owned `setMarkedText` preedit; arrow move/page, half-shape punct/digit empty-preedit commit, Shift literal letters, Control/Option/Command passthrough, Enter raw-preedit commit, Left Shift CN/EN toggle via `recognizedEvents`/`flagsChanged`, English mode IPC-free passthrough for password/login when ABC is unavailable, wire key_code = macOS CG (ADR 0002), default xiaohe, surteam ids, rounded blue `TS` menu icon + installer AppIcon landed in source; System Settings Add-under-简体中文 browser / TextEdit typing still pending after reinstall | open (UI/typing) |
| KI-009 | macos-keycodes | Earlier ASCII 8/27/32 semantic remap collided with CG `c`/`-`/`u`. Fixed by adopting Karabiner CG codes on the wire (incl. Return=36 / keypad Enter=76) and dropping ASCII edit codes; real-env TextEdit recheck still pending | mitigated (source) |
| KI-007 | macos-signing | SurTeamCode is self-signed (not Developer ID); `spctl` rejects; Gatekeeper may warn; notarization not claimed; hardened-runtime build needs library-validation entitlement for Homebrew protobuf | open |
| KI-006 | shuangpin | Xiaohe/flypy keeps fixed 2-key boundaries (`wo`→我), expands ambiguous keys to all valid readings (`lo`/`lt`/`dv`…), soft-confirm composition (`你好uijp`), short half-code hits (`nih`→你好, single `h` prefers short words), and Backspace undo of soft confirms with cursor restored to preedit end; macOS candidates stay Chinese-only; real-env phrase flow still needs user trial | open (trial) |
| KI-008 | macos-install | Installer now targets `/Library/Input Methods` and prompts for admin; user-level `~/Library/Input Methods` alone was insufficient for stable TIS visibility in this environment. Install/update now kills/refreshes the IMKit process and the menu exposes Restart; logout remains only a last-resort UI fallback | mitigated (source) |
| KI-010 | daemon-rename | Daemon binary/LaunchAgent renamed `imed` → `tsimed` / `com.surteam.tsimed`; installer boots out legacy `com.surteam.tsime.imed` labels, but previously installed leftover agents may still need a reinstall/kickstart | mitigated (source) |
| KI-011 | macos-daemon | After DMG install/Restart, Chinese mode can look like passthrough (no preedit/candidates) when `com.surteam.tsimed` is enabled but not running / socket missing. Menu Restart and install scripts now bootstrap+kickstart the agent; logout is not required | mitigated (source) |

Notes:

- Daemon and continuous pinyin decode are implemented and unit-tested; not yet verified inside host IME UIs.
- Upstream `base.dict.yaml` header comments contain literal `...` substrings; dictBuilder splits only on a whole-line `...` marker.
