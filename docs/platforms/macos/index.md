# Platform: macOS

**Status:** implemented / compiled on this host (IMKit app bundle + SurTeam-signed DMG package). Simplified Chinese Hans mode is TIS-registered, enabled, selectable, and selectable-as-current. **real-env typing: no** (TextEdit typing not verified by agent). System Settings Add-under-简体中文 browser not agent-confirmed.

## Adapter

- InputMethodKit server under `platforms/macos`.
- Links `libtsime_ipc` (framed Protobuf over UDS).
- English menu mode: keys are not forwarded to the daemon (native passthrough).
- Chinese mode: `handle(_:client:)` → `tsime_client_handle_key` → preedit/commit from `KeyEventResult`.

## Input-source registration

`platforms/macos/Info.plist` declares a mode-enabled IMKit source:

- Bundle ID / `TISInputSourceID`: `com.surteam.inputmethod.TSInputMethod`
- Visible Hans mode: `com.surteam.inputmethod.TSInputMethod.Hans` with `TISIntendedLanguage=zh-Hans`
- `ComponentInputModeDict` + `tsInputMethodCharacterRepertoireKey` + custom SurTeam/TS menu/app icons

Installer copies the app into `/Library/Input Methods` (admin prompt) and runs `TSInputMethod --register-input-source` (`TISRegisterInputSource` + `TISEnableInputSource` for Hans + `AppleEnabledThirdPartyInputSources` record).

Observed after install/register on this host:

- `TISCreateInputSourceList` returns parent `com.surteam.inputmethod.TSInputMethod` and Hans mode `com.surteam.inputmethod.TSInputMethod.Hans`
- Parent + Hans are both TIS-enabled; Hans is selectable (`langs=zh-Hans`); parent stays non-selectable
- Installer records both under `com.apple.inputsources` `AppleEnabledThirdPartyInputSources`
- After enable, Hans appears in the enabled-only TIS list and `TISSelectInputSource` can make it current

System Settings → Add Input Source → 简体中文 browser was not agent-confirmed (already-enabled sources may be absent from the Add browser and appear in the enabled list / menu bar instead). TextEdit typing remains pending. If the menu bar does not refresh, reopen Keyboard settings or log out/in.

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
3. Installs LaunchAgent `com.surteam.tsimed` to auto-start embedded `tsimed`.
4. Defaults `TSIME_SCHEME=xiaohe` and points `TSIME_DICT` at the embedded rime-ice dictionary.

If System Settings → Keyboard → Input Sources → 简体中文 does not list it yet, reopen the pane or log out/in; agent has verified TIS listing/enable, not the Settings UI itself.

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
- Candidate UI uses horizontal `IMKCandidates` (`kIMKSingleRowSteppingCandidatePanel`, 9 numbered labels per page) plus marked text/commit; `-`/`=` page previous/next; Space commits the top candidate; Enter/Return commits raw preedit and clears the panel.
- Left Shift alone (`flagsChanged`, keyCode 56) toggles in-IME Chinese/English after committing any current preedit as raw text and clearing candidates; Shift used as a chord modifier does not toggle.
- `KeyEvent.key_code` is the raw macOS CG virtual key from `NSEvent.keyCode` (Karabiner-Elements table: `delete_or_backspace=0x33`, `escape=0x35`, `spacebar=0x31`, `return_or_enter=0x24`, `keypad_enter=0x4c`, `delete_forward=0x75`, letters such as `c=0x08`). No ASCII 8/27/32 remap. Printable compose/paging still uses `characters` ([ADR 0002](../../decisions/0002-semantic-edit-keycodes.md)).
- Menu/palette icon is a rounded blue PDF with white `TS` at **21.75×15.75 pt** (same MediaBox as Fcitx5’s macOS menu icon; landscape, not the 128 pt app mark). Source SVG: `platforms/macos/Resources/menu_icon.svg`. App/installer bundles share `AppIcon.icns`.

## Verification

| implemented | compiled | auto-tested | real-env |
| --- | --- | --- | --- |
| yes | yes (`make app` + packaged DMG) | package/codesign verify + TIS enable/select after register | listing/enable/select yes; Settings Add browser / typing no |
