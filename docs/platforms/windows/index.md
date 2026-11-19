# Platform: Windows

**Status:** source implemented (TSF COM text service + install/uninstall scripts). **compiled/real-env: no** on this macOS host.

## Adapter

- TSF `ITfTextInputProcessorEx` + `ITfKeyEventSink` under `platforms/windows`.
- Named Pipe IPC via shared `tsime_c_api` (`TSIME_PIPE` or `\\.\pipe\tsime-<user>`).
- English mode: keys not eaten (native passthrough).
- Chinese mode: editable keys forwarded to daemon; commit text inserted through TSF.

## Build / install (on Windows)

```bat
cmake -S platforms/windows -B platforms/windows/build -A x64
cmake --build platforms/windows/build --config Release
powershell ./platforms/windows/scripts/install.ps1
```

## Verification

| implemented | compiled | auto-tested | real-env |
| --- | --- | --- | --- |
| yes | no (needs MSVC) | no | no |
