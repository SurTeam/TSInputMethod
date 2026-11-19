# Platform: Linux

**Status:** source implemented (Fcitx5 addon + install/uninstall). **compiled/real-env: no** on this macOS host (Fcitx5 headers missing).

## Adapter

- Fcitx5 `InputMethodEngineV2` under `platforms/linux`.
- UDS IPC via `tsime_c_api`.
- English mode does not consume keys; Chinese mode updates preedit + CandidateList from daemon results.
- Edit keys translate into macOS CG codes on the wire (BackSpace→51, Delete→117, Escape→53, space→49); printable input uses `characters` ([ADR 0002](../../decisions/0002-semantic-edit-keycodes.md)).

## Build / install (on Linux)

```bash
cmake -S platforms/linux -B platforms/linux/build
cmake --build platforms/linux/build
./platforms/linux/scripts/install.sh
```

## Verification

| implemented | compiled | auto-tested | real-env |
| --- | --- | --- | --- |
| yes | no (needs Fcitx5) | no | no |
