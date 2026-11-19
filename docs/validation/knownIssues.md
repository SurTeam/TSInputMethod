# Known issues

| ID | Area | Summary | Status |
| --- | --- | --- | --- |
| KI-001 | dictionary | Optional `ext` rime-ice table not built by default (size/time); 8105+base is the default index | open |
| KI-002 | platforms | Host IME real-env validation pending for all three adapters; Windows/Linux not compiled on this Mac | open |
| KI-003 | ranking | Real neural backends / weights not shipped; FastRanker + injectable Backend + no-model degrade landed | open (models) |
| KI-004 | windows-ipc | Named Pipe source landed (windows tags) but not executed on a Windows host | open |
| KI-005 | macos-imkit | SurTeam-signed DMG OOBE + ComponentInputModeDict/TIS register+enable landed; horizontal IMKCandidates (`kIMKSingleRowSteppingCandidatePanel`, 9 numbered labels), Enter raw-preedit commit, Left Shift CN/EN toggle, wire key_code = macOS CG (ADR 0002), default xiaohe, surteam ids, rounded blue `TS` menu icon + installer AppIcon landed in source; System Settings Add-under-简体中文 browser / TextEdit typing still pending after reinstall | open (UI/typing) |
| KI-009 | macos-keycodes | Earlier ASCII 8/27/32 semantic remap collided with CG `c`/`-`/`u`. Fixed by adopting Karabiner CG codes on the wire (incl. Return=36 / keypad Enter=76) and dropping ASCII edit codes; real-env TextEdit recheck still pending | mitigated (source) |
| KI-007 | macos-signing | SurTeamCode is self-signed (not Developer ID); `spctl` rejects; Gatekeeper may warn; notarization not claimed; hardened-runtime build needs library-validation entitlement for Homebrew protobuf | open |
| KI-006 | shuangpin | Xiaohe/flypy keeps fixed 2-key boundaries (uijp→shi'jie) and progressive N…1 prefix candidates (incl. odd-length remainder); real-env phrase flow still needs user trial | open (trial) |
| KI-008 | macos-install | Installer now targets `/Library/Input Methods` and prompts for admin; user-level `~/Library/Input Methods` alone was insufficient for stable TIS visibility in this environment | open (document) |
| KI-010 | daemon-rename | Daemon binary/LaunchAgent renamed `imed` → `tsimed` / `com.surteam.tsimed`; installer boots out legacy `com.surteam.tsime.imed` and `com.surteam.tsime.imed` labels, but previously installed leftover agents may still need a reinstall/kickstart | mitigated (source) |

Notes:

- Daemon and continuous pinyin decode are implemented and unit-tested; not yet verified inside host IME UIs.
- Upstream `base.dict.yaml` header comments contain literal `...` substrings; dictBuilder splits only on a whole-line `...` marker.
