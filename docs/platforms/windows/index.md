# Platform: Windows

**Status:** source implemented (TSF COM text service + install/uninstall scripts). **compiled/real-env: no** on this macOS host.

## Adapter

- TSF `ITfTextInputProcessorEx` + `ITfKeyEventSink` under `platforms/windows`.
- Named Pipe IPC via shared `tsime_c_api` (`TSIME_PIPE` or `\\.\pipe\tsime-<user>`).
- English mode: keys not eaten (native passthrough).
- Chinese mode: editable keys forwarded to daemon; commit text inserted through TSF.
- `OnSetFocus(FALSE)` calls `InvalidateLearningContext` so committed learning context is not reused after app focus loss; accurate post-commit Backspace/external-edit tracking remains a follow-up.
- Host virtual keys translate into macOS CG codes on the wire (VK_BACK→51, VK_ESCAPE→53, VK_SPACE→49, VK_DELETE→117); letters/digits use `characters` ([ADR 0002](../../decisions/0002-semantic-edit-keycodes.md)).

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
