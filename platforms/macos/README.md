# platforms/macos

InputMethodKit adapter for TSInputMethod.

## Build

```bash
cd platforms/macos
make app
```

Requires: Xcode/swiftc, cmake, protobuf, built `platforms/common` (`libtsime_ipc`).

## Trial package (preferred)

```bash
bash ../../scripts/package_macos.sh
```

Primary artifact: `dist/macos/TSInputMethod-macos.dmg`.

Open the DMG → double-click **Install TSInputMethod** → enable TSInputMethod under System Settings → Keyboard → Input Sources → 简体中文.

The installer places the app under `/Library/Input Methods` (admin prompt), registers/enables the Hans mode with Text Input Source Services, refreshes the running IMKit process, and loads a LaunchAgent that auto-starts embedded `tsimed` with `TSIME_SCHEME=xiaohe`. If the menu bar is still stale afterward, use the IME menu **Restart** item or reopen Keyboard settings; log out/in only as a last resort.

Signed with the SurTeam keychain identity (`SurTeamCode` on this machine). Self-signed / non-Developer-ID limits apply; notarization is not claimed.

## Developer install helpers

```bash
./scripts/install.sh
./scripts/uninstall.sh
```

These install `build/TSInputMethod.app` into `/Library/Input Methods`, register the input source, and refresh the IMKit process; they do not replace the DMG OOBE path.

## Runtime

- English mode: menu/Left-Shift toggle; ordinary keys are not forwarded to the daemon and do not require IPC (works offline for password/login when ABC is gone).
- Menu-bar **Restart** commits/clears composition, bootstraps/`kickstart`s `tsimed` if needed, and terminates the IMKit process so an update can load without logout. Passthrough with no preedit/candidates usually means the daemon is down, not that a logout is required.
- Chinese mode: keys go through framed Protobuf IPC; preedit/candidates/commit follow Go session results; a custom `NSPanel` candidate row draws 9 equal-width columns with darker digit labels and `selected_index` highlight (mouse wheel pages); composition marked text alone shows soft-confirmed Chinese + remaining raw keys; Left Shift alone toggles Chinese/English via `recognizedEvents`+`flagsChanged`; `-`/`=` page; Delete/Escape/Space use daemon CG keycodes.
- Socket: `TSIME_SOCKET`, else `$XDG_RUNTIME_DIR/tsime.sock`, else temporary `tsime-$UID.sock`.

## Verification

See [docs/platforms/macos/index.md](../../docs/platforms/macos/index.md). TIS listing for Hans/`zh-Hans` is verified after register; host typing is not claimed until exercised.
