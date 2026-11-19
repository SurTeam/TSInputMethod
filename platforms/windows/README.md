# platforms/windows

TSF Text Service adapter for TSInputMethod.

## Build (MSVC / Windows)

```bat
cmake -S platforms/windows -B platforms/windows/build -A x64
cmake --build platforms/windows/build --config Release
```

Produces `tsime_tsf.dll` linked against `libtsime_ipc` and Named Pipe IPC.

## Install / uninstall

```powershell
./scripts/install.ps1 -DllPath .\build\Release\tsime_tsf.dll
./scripts/uninstall.ps1 -DllPath .\build\Release\tsime_tsf.dll
```

Registers the COM InprocServer under HKCU. Add/remove the language profile in Windows Settings as needed.

## Runtime

- Run `imed.exe` (Named Pipe listener).
- Pipe: `TSIME_PIPE` or `\\.\pipe\tsime-<username>`.
- English mode leaves keys uneaten (native passthrough). Chinese mode forwards editable keys to the daemon and commits returned text.

## Verification

This macOS development host does **not** MSVC-compile or real-env test the DLL. Status is source-implemented + scripted install/uninstall only until built on Windows.
