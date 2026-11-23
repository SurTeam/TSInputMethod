# Platform: macOS

**Status:** implemented / compiled on this host (IMKit app bundle + SurTeam-signed DMG package). Simplified Chinese Hans mode is TIS-registered, enabled, selectable, and selectable-as-current. **real-env typing: no** (TextEdit typing not verified by agent). System Settings Add-under-简体中文 browser not agent-confirmed.

## Adapter

- InputMethodKit server under `platforms/macos`.
- Links `libtsime_ipc` (framed Protobuf over UDS).
- English menu/Left-Shift mode: keys are not forwarded to the daemon (native passthrough). Entering or staying in English does not require a live daemon socket, so password/login fields still accept typing when ABC is unavailable and `tsimed` is down. In-IME Chinese/English mode is process-wide shared across all IMK `TSInputController` / `DaemonClient` instances (including shared alone-modifier tracking and a short toggle debounce) so focusing another client or releasing Shift after a CN→EN switch does not silently revert to Chinese; switching to English sticks the mode bit before committing any local preedit.
- Chinese mode: `handle(_:client:)` → `tsime_client_handle_key` with AppKit modifier bits → preedit/commit/highlight from `KeyEventResult`. Control/Option/Command chords return before IPC so host shortcuts stay native.

## Input-source registration

`platforms/macos/Info.plist` declares a mode-enabled IMKit source. `InputMethodServerControllerClass` / `InputMethodServerDelegateClass` must be the ObjC runtime name `TSInputController` (from `@objc(TSInputController)`), not the Swift-qualified `TSInputMethod.TSInputController`; a mismatch prevents IMKit from creating the controller, so the menu-bar menu (including Settings) is empty and key handling never runs.

- Bundle ID / `TISInputSourceID`: `com.surteam.inputmethod.TSInputMethod`
- Visible Hans mode: `com.surteam.inputmethod.TSInputMethod.Hans` with `TISIntendedLanguage=zh-Hans`
- `ComponentInputModeDict` + `tsInputMethodCharacterRepertoireKey` + custom SurTeam/TS menu/app icons

Installer copies a SurTeam-signed app into `/Library/Input Methods` (admin/root), strips quarantine (`xattr`), and runs `TSInputMethod --register-input-source` / `--enable-input-source` as the console GUI user (Fcitx5-style `su -m <user>` when the installer itself is root). Enabling as root leaves `parent enabled=false` and `TISSelectInputSource(Hans)=-50`; enabling as the real user sets non-selectable parent + Hans enabled and selects Hans. Register still skips `TISRegisterInputSource` when parent+Hans already exist; records parent+Hans under `AppleEnabledThirdPartyInputSources`; refreshes `TextInputMenuAgent`/`TextInputSwitcher` when duplicate TIS IDs are detected. Install/register does not modify `com.apple.HIToolbox.plist` and does not restore ABC.

Observed after install/register on this host:

- `TISCreateInputSourceList` returns parent `com.surteam.inputmethod.TSInputMethod` and Hans mode `com.surteam.inputmethod.TSInputMethod.Hans`
- Parent is TIS-enabled but non-selectable; Hans is enabled, selectable (`langs=zh-Hans`), and selected by the installer so it can appear in the menu/enabled list without a second selectable same-name row
- Installer records parent `Keyboard Input Method` + Hans `Input Mode` under `com.apple.inputsources` `AppleEnabledThirdPartyInputSources` only
- After enable/select, Hans appears in the enabled-only TIS list and can be current
- Repeated unconditional `TISRegisterInputSource` can leave Settings showing many same-name「TS输入法」rows even while prefs stay short; register is now idempotent and TextInput agents are refreshed on duplicate detection

System Settings → Add Input Source → 简体中文 browser was not agent-confirmed (already-enabled sources may be absent from the Add browser and appear in the enabled list / menu bar instead). TextEdit typing remains pending. After install/update the installer refreshes the IMKit process; if the menu bar is still stale, use the IME menu **Restart** item or reopen Keyboard settings, and log out/in only as a last resort.

## Trial package (primary)

```bash
bash scripts/package_macos.sh
# → dist/macos/TSInputMethod-macos.dmg  (primary)
# → dist/macos/TSInputMethod-macos.zip   (secondary fallback)
```

Open the DMG and double-click **Install TSInputMethod**.

That installer:

1. Copies embedded `TSInputMethod.app` into `/Library/Input Methods` (admin privileges).
2. Registers and enables the Hans input mode with Text Input Source Services.
3. Installs LaunchAgent `com.surteam.tsimed` (`RunAtLoad` + `KeepAlive`) so embedded `tsimed` starts and warms at login/session load, not at first IME switch/keystroke.
4. Defaults `TSIME_SCHEME=xiaohe` and points `TSIME_DICT` at the embedded rime-ice dictionary.
5. Refreshes the running IMKit process (`killall TSInputMethod`) so an update does not require logout by default.

If System Settings → Keyboard → Input Sources → 简体中文 does not list it yet, reopen the pane or use the IME menu **Restart** item; log out/in only as a last resort. Agent has verified TIS listing/enable, not the Settings UI itself.

Developer-only local build helpers under `platforms/macos/scripts/` remain available; they are not the primary trial path.

## Signing

`scripts/package_macos.sh` resolves a SurTeam codesign identity from the keychain (default match `SurTeam*`, observed `SurTeamCode`) and signs:

- `TSInputMethod.app` and embedded `tsimed` (with `platforms/macos/TSInputMethod.entitlements` disabling library validation so Homebrew protobuf can load under hardened runtime)
- `Install TSInputMethod.app`
- the DMG

Observed identity type: self-signed `SurTeamCode` (issuer=`SurTeamCode`), not Apple Developer ID Application. `codesign --verify` succeeds; `spctl --assess` rejects the app/DMG. Gatekeeper may still warn for downloaded artifacts. Notarization is not claimed.

Override with `TSIME_SIGN_IDENTITY` if needed.

## Runtime notes

- Packaged trial default scheme: Xiaohe (`TSIME_SCHEME=xiaohe`).
- Socket: `TSIME_SOCKET`, else `$XDG_RUNTIME_DIR/tsime.sock`, else temporary `tsime-$UID.sock`.
- Candidate UI is a custom floating `NSPanel` (`CandidatePanel.swift`) owned by the adapter: window level `.popUpMenu` (above Spotlight/focus search), up to 9 content-sized columns with tighter panel/item padding, darker numeric labels `1`–`9`, highlight synced from session `selected_index`, and mouse-wheel paging; the panel width follows the visible candidates instead of reserving empty equal-width slots. Composition/preedit is owned via `setMarkedText` attributes on the client (underline + accent wash), not via `IMKCandidates`. The adapter strips session-side `N.` text prefixes before drawing so labels are not doubled. Candidates stay Chinese-only; composition alone carries soft-confirmed Chinese plus remaining raw keys (e.g. `你好uijp`). `selectCandidate` returns the same composition/candidates payload as key handling; Left/Right move highlight; Up/Down and `-`/`=` page previous/next; Space soft-confirms/finishes the highlighted candidate; Enter/Return hard-commits the whole composition and clears the panel. On `deactivateServer` / `commitComposition` (e.g. click outside Spotlight), the adapter commits current preedit as raw text, clears marked text/candidates, hides the panel, and sends Escape to the daemon so session compose state is cleared. `deactivateServer` also sends `InvalidateLearningContext` so committed learning context from the previous focus/IME client is not reused. Empty-preedit digits/punctuation commit through the session half-shape map. Column count, font name/size, and highlight color are user-configurable (see Settings). Candidate panel selection uses highlight color without bold/semibold body text. See [ADR 0003](../../decisions/0003-custom-candidate-panel.md) and [ADR 0004](../../decisions/0004-macos-settings-xdg-config.md).
- `recognizedEvents` includes `flagsChanged` as well as `keyDown`. In-IME Chinese/English toggles go through `ModeToggleHotkeys` (`ModeToggleHotkeys.swift`) with default bindings Left Shift alone (`flagsChanged`, keyCode 56) and F18 (`keyDown`, keyCode 79, for Karabiner CapsLock→F18). Both call the same toggle path: flip the process-wide mode bit first (English sticky), then commit any current preedit as raw text and clear candidates. Shift used as a chord modifier does not toggle; F18 with Shift/Control/Option/Command does not toggle. Duplicate `flagsChanged` (same modifier mask) is ignored. Alone-modifier toggles fire on release only; chorded Shift (used with another key) does not toggle. Alone-modifier tracking is process-wide, and a short debounce suppresses duplicate toggles from overlapping IMK controllers. Mode state is process-wide shared so English stays while typing across clients. `TSInputController.modeToggleBindings()` / `setModeToggleBindings(_:)` expose the binding list; Settings can replace it with a recorded multi-hotkey list. Caps Lock LED remains outside the IME (system/Karabiner).
- Daemon lifecycle: LaunchAgent owns `tsimed` across IME switches. Soft ensure (`DaemonAgent.softEnsureRunning`) bootstraps/enables when unloaded and uses `kickstart` without `-k` only when IPC is unreachable; menu **Restart** alone uses force restart (`kickstart -k`). `DaemonWatchdog` periodically hello-probes reachability from IMKit (independent of keystrokes), soft-ensures on failure, and escalates to force restart after repeated soft failures; Chinese-mode IPC connect/handle failures also request the same recovery. Switching away / `deactivateServer` closes only adapter client/session state and leaves the already-initialized daemon running. See [ADR 0005](../../decisions/0005-macos-daemon-login-lifecycle.md).
- The menu-bar IME menu keeps Chinese/English toggles, adds **Settings** (opens the native settings `NSWindow`), and **Restart**: commit/clear local composition, tear down IPC, immediately show the initializing notification, force-restart `com.surteam.tsimed`, notify ready when IPC is reachable again, then terminate the IMKit process so the host reloads the installed binary. If Chinese typing looks like passthrough with no preedit/candidates, the watchdog should already be recovering; **Restart** or `launchctl kickstart -k gui/$UID/com.surteam.tsimed` remains the manual path; logout is not required.
- **Settings** (`SettingsWindowController` + `ConfigStore`): CN/EN mode-toggle hotkeys (multiple; add/remove; capture by recording key input), candidate display column count (1–9; adapter clamps the daemon page to this UI limit), candidate panel font name/size, and highlight color. Persistence path: `$XDG_CONFIG_HOME/tsime/settings.json`, else `~/.config/tsime/settings.json`; the config directory is created on first save/load. See [ADR 0004](../../decisions/0004-macos-settings-xdg-config.md).
- **Local notifications** (`AppNotifier`): IMKit startup and `DaemonWatchdog` recovery reset lifecycle flags and notify initializing/ready around soft-ensure or Restart; if `tsimed` is already up at first probe it notifies ready without an initializing banner. Notification authorization failure is a silent degrade (IME continues without prompts).
- Expected Settings/lifecycle sources under `platforms/macos/Sources/TSInputMethod/`: `ConfigStore.swift`, `SettingsWindowController.swift`, `AppNotifier.swift`, `DaemonAgent.swift`, `DaemonWatchdog.swift`, wired from `InputController.swift`, `CandidatePanel.swift`, `DaemonClient.swift`, `main.swift`, and the macOS `Makefile`.
- `KeyEvent.key_code` is the raw macOS CG virtual key from `NSEvent.keyCode` (Karabiner-Elements table: `delete_or_backspace=0x33`, `escape=0x35`, `spacebar=0x31`, `return_or_enter=0x24`, `keypad_enter=0x4c`, `delete_forward=0x75`, letters such as `c=0x08`). No ASCII 8/27/32 remap. Printable compose/paging still uses `characters` ([ADR 0002](../../decisions/0002-semantic-edit-keycodes.md)).
- Menu/palette icon is a black rounded-rectangle PDF silhouette with hollow/cut-out `TS` letterforms (even-odd path) at **21.75×15.75 pt** (same MediaBox as Fcitx5’s macOS menu icon; landscape, not the 128 pt app mark), so menu-bar template rendering shows the letters instead of a blank pill. Source SVG: `platforms/macos/Resources/menu_icon.svg`. App/installer bundles share `AppIcon.icns`.

## Verification

| Area | implemented | compiled | auto-tested | real-env |
| --- | --- | --- | --- | --- |
| IMKit adapter / package / TIS | yes | yes (`make app` + packaged DMG) | package/codesign verify + TIS enable/select after register | listing/enable/select yes; System Settings Add browser / typing no |
| Settings window / XDG `settings.json` / AppNotifier | yes (`ConfigStore` / `SettingsWindowController` / `AppNotifier` + menu wiring) | yes (`make -C platforms/macos app`) | not claimed | Settings UI / notification auth not agent-confirmed after reinstall |
| Daemon login soft-ensure / watchdog / cross-IME KeepAlive | yes (`DaemonAgent` soft vs force + `DaemonWatchdog` hello probe + `DaemonClient` recovery hooks) | yes (`make -C platforms/macos app`) | not claimed | LaunchAgent/watchdog path compiled; real-env death/Restart notify cycle not agent-claimed this turn |
