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

Open the DMG → double-click **Install TSInputMethod** → enable TSInputMethod in System Settings.

The installer places the app under `~/Library/Input Methods` and loads a LaunchAgent that auto-starts embedded `imed` with `TSIME_SCHEME=xiaohe`.

Signed with the SurTeam keychain identity (`SurTeamCode` on this machine). Self-signed / non-Developer-ID limits apply; notarization is not claimed.

## Developer install helpers

```bash
./scripts/install.sh
./scripts/uninstall.sh
```

These copy `build/TSInputMethod.app` only and do not replace the DMG OOBE path.

## Runtime

- English mode: menu toggle; ordinary keys are not forwarded to the daemon.
- Chinese mode: keys go through framed Protobuf IPC; preedit/candidates/commit follow Go session results.
- Socket: `TSIME_SOCKET`, else `$XDG_RUNTIME_DIR/tsime.sock`, else temporary `tsime-$UID.sock`.

## Verification

See [docs/platforms/macos/index.md](../../docs/platforms/macos/index.md). Real-env host typing is not claimed until exercised.
