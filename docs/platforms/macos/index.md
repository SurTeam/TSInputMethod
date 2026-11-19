# Platform: macOS

**Status:** implemented / compiled on this host (IMKit app bundle + SurTeam-signed DMG package). **real-env: no** (system Input Source enable / TextEdit typing not verified by agent).

## Adapter

- InputMethodKit server under `platforms/macos`.
- Links `libtsime_ipc` (framed Protobuf over UDS).
- English menu mode: keys are not forwarded to the daemon (native passthrough).
- Chinese mode: `handle(_:client:)` → `tsime_client_handle_key` → preedit/commit from `KeyEventResult`.

## Trial package (primary)

```bash
bash scripts/package_macos.sh
# → dist/macos/TSInputMethod-macos.dmg  (primary)
# → dist/macos/TSInputMethod-macos.zip   (secondary fallback)
```

Open the DMG and double-click **Install TSInputMethod**.

That installer:

1. Copies embedded `TSInputMethod.app` into `~/Library/Input Methods`.
2. Installs LaunchAgent `com.surteam.tsime.imed` to auto-start embedded `imed`.
3. Defaults `TSIME_SCHEME=xiaohe` and points `TSIME_DICT` at the embedded rime-ice dictionary.

Then enable the input source in System Settings → Keyboard → Input Sources.

Developer-only local build helpers under `platforms/macos/scripts/` remain available; they are not the primary trial path.

## Signing

`scripts/package_macos.sh` resolves a SurTeam codesign identity from the keychain (default match `SurTeam*`, observed `SurTeamCode`) and signs:

- `TSInputMethod.app` and embedded `imed`
- `Install TSInputMethod.app`
- the DMG

Observed identity type: self-signed `SurTeamCode` (issuer=`SurTeamCode`), not Apple Developer ID Application. `codesign --verify` succeeds; `spctl --assess` rejects the app/DMG. Gatekeeper may still warn for downloaded artifacts. Notarization is not claimed.

Override with `TSIME_SIGN_IDENTITY` if needed.

## Runtime notes

- Packaged trial default scheme: Xiaohe (`TSIME_SCHEME=xiaohe`).
- Socket: `TSIME_SOCKET`, else `$XDG_RUNTIME_DIR/tsime.sock`, else temporary `tsime-$UID.sock`.
- Candidate UI is minimal (marked text / commit); host polish is ongoing.

## Verification

| implemented | compiled | auto-tested | real-env |
| --- | --- | --- | --- |
| yes | yes (`make app` + packaged DMG) | package/codesign verify only | no |
