# Platform: Linux

**Status:** source implemented (Fcitx5 addon + install/uninstall). **compiled/real-env: no** on this macOS host (Fcitx5 headers missing).

## Adapter

- Fcitx5 `InputMethodEngineV2` under `platforms/linux`.
- UDS IPC via `tsime_c_api`.
- English mode does not consume keys; Chinese mode updates preedit + CandidateList from daemon results.

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
