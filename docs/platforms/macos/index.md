# Platform: macOS

**Status:** implemented / compiled on this host (IMKit app bundle). **real-env: no** (not registered/typed in TextEdit this turn).

## Adapter

- InputMethodKit server under `platforms/macos`.
- Links `libtsime_ipc` (framed Protobuf over UDS).
- English menu mode: keys are not forwarded to the daemon (native passthrough).
- Chinese mode: `handle(_:client:)` → `tsime_client_handle_key` → preedit/commit from `KeyEventResult`.

## Build / install

```bash
cd platforms/macos && make app
./scripts/install.sh
./scripts/uninstall.sh
```

Trial package (app + imed + dict + scripts):

```bash
bash scripts/package_macos.sh
# → dist/macos/TSInputMethod-macos.zip
```

Default daemon scheme for trial: Xiaohe (`TSIME_SCHEME=xiaohe`).

Observed: `make app` produced `platforms/macos/build/TSInputMethod.app`.

## Verification

| implemented | compiled | auto-tested | real-env |
| --- | --- | --- | --- |
| yes | yes (`make app`) | no host IME automation | no |
