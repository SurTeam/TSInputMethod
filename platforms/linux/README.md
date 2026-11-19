# platforms/linux

Fcitx5 addon for TSInputMethod.

## Build (Linux with Fcitx5)

```bash
cmake -S platforms/linux -B platforms/linux/build
cmake --build platforms/linux/build
```

Requires Fcitx5 development packages (`Fcitx5Core`, `Fcitx5Module`) and protobuf.

## Install / uninstall

```bash
./scripts/install.sh
./scripts/uninstall.sh
```

## Runtime

- Start `imed`.
- Socket: `TSIME_SOCKET` / `$XDG_RUNTIME_DIR/tsime.sock` / `/tmp/tsime-$UID.sock`.
- English mode does not consume keys; Chinese mode uses daemon composition/candidates via InputPanel.

## Verification

This macOS host lacks Fcitx5 headers; configure/build was not completed here. Treat as source-implemented until built on a Linux Fcitx5 system.
