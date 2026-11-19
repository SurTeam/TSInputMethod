# platforms/macos

InputMethodKit adapter for TSInputMethod.

## Build

```bash
cd platforms/macos
make app
```

Requires: Xcode/swiftc, cmake, protobuf, built `platforms/common` (`libtsime_ipc`).

## Install / uninstall

```bash
./scripts/install.sh
./scripts/uninstall.sh
```

Installs to `~/Library/Input Methods/TSInputMethod.app`.

## Runtime

- Start `imed` first (`TSIME_SCHEME=xiaohe` for Xiaohe/flypy; default quanpin if unset).
- Socket: `TSIME_SOCKET`, else `$XDG_RUNTIME_DIR/tsime.sock`, else `/tmp/tsime-$UID.sock`.
- English mode: menu toggle; ordinary keys are not forwarded to the daemon.
- Chinese mode: keys go through framed Protobuf IPC; preedit/candidates/commit follow Go session results.
- Packaged trial: `bash scripts/package_macos.sh` → `dist/macos/TSInputMethod-macos.zip`.

## Verification

See [docs/platforms/macos/index.md](../../docs/platforms/macos/index.md). Real-env host typing is not claimed until exercised.
