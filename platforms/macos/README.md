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

The installer places the app under `/Library/Input Methods` (admin prompt), registers/enables the Hans mode with Text Input Source Services, and loads a LaunchAgent that auto-starts embedded `imed` with `TSIME_SCHEME=xiaohe`.

Signed with the SurTeam keychain identity (`SurTeamCode` on this machine). Self-signed / non-Developer-ID limits apply; notarization is not claimed.

## Developer install helpers

```bash
./scripts/install.sh
./scripts/uninstall.sh
```

These install `build/TSInputMethod.app` into `/Library/Input Methods` and register the input source; they do not replace the DMG OOBE path.

## Runtime

- English mode: menu toggle; ordinary keys are not forwarded to the daemon.
- Chinese mode: keys go through framed Protobuf IPC; preedit/candidates/commit follow Go session results.
- Socket: `TSIME_SOCKET`, else `$XDG_RUNTIME_DIR/tsime.sock`, else temporary `tsime-$UID.sock`.

## Verification

See [docs/platforms/macos/index.md](../../docs/platforms/macos/index.md). TIS listing for Hans/`zh-Hans` is verified after register; host typing is not claimed until exercised.
