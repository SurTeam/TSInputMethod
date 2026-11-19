#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST="$ROOT/dist/macos"
STAGE="$DIST/TSInputMethod-macos"
APP_SRC="$ROOT/platforms/macos/build/TSInputMethod.app"
APP_STAGE="$STAGE/TSInputMethod.app"
INSTALLER_STAGE="$STAGE/Install TSInputMethod.app"
DMG_STAGE="$DIST/dmg-root"
DMG_PATH="$DIST/TSInputMethod-macos.dmg"
SIGN_IDENTITY="${TSIME_SIGN_IDENTITY:-}"

export GOSUMDB="${GOSUMDB:-off}"

resolve_sign_identity() {
  if [[ -n "$SIGN_IDENTITY" ]]; then
    return 0
  fi
  local line
  line="$(security find-identity -v -p codesigning | awk '/SurTeam/ {print; exit}')"
  if [[ -z "$line" ]]; then
    echo "error: no SurTeam codesigning identity found in keychain" >&2
    exit 1
  fi
  SIGN_IDENTITY="$(printf '%s\n' "$line" | sed -E 's/.*"([^"]+)".*/\1/')"
  if [[ -z "$SIGN_IDENTITY" ]]; then
    echo "error: failed to parse SurTeam identity from: $line" >&2
    exit 1
  fi
}

sign_path() {
  local target="$1"
  codesign --force --options runtime --timestamp=none --sign "$SIGN_IDENTITY" "$target"
}

sign_app_bundle() {
  local app="$1"
  local entitlements="$ROOT/platforms/macos/TSInputMethod.entitlements"
  if [[ -x "$app/Contents/MacOS/imed" ]]; then
    sign_path "$app/Contents/MacOS/imed"
  fi
  if [[ -x "$app/Contents/MacOS/TSInputMethod" ]]; then
    if [[ -f "$entitlements" ]]; then
      codesign --force --options runtime --timestamp=none --entitlements "$entitlements" --sign "$SIGN_IDENTITY" "$app/Contents/MacOS/TSInputMethod"
    else
      sign_path "$app/Contents/MacOS/TSInputMethod"
    fi
  fi
  if [[ -x "$app/Contents/MacOS/install" ]]; then
    sign_path "$app/Contents/MacOS/install"
  fi
  if [[ -f "$entitlements" && -x "$app/Contents/MacOS/TSInputMethod" ]]; then
    codesign --force --deep --options runtime --timestamp=none --entitlements "$entitlements" --sign "$SIGN_IDENTITY" "$app"
  else
    codesign --force --deep --options runtime --timestamp=none --sign "$SIGN_IDENTITY" "$app"
  fi
}

rm -rf "$STAGE" "$DMG_STAGE"
mkdir -p "$STAGE" "$DMG_STAGE"

echo "==> build imed"
mkdir -p "$STAGE/bin"
(cd "$ROOT" && go build -o "$STAGE/bin/imed" ./cmd/imed)

echo "==> build macOS IMKit app"
(cd "$ROOT/platforms/macos" && make app)
rm -rf "$APP_STAGE"
cp -R "$APP_SRC" "$APP_STAGE"

echo "==> embed imed + dictionary into app bundle"
mkdir -p "$APP_STAGE/Contents/MacOS" \
  "$APP_STAGE/Contents/Resources/data/dicts" \
  "$APP_STAGE/Contents/Resources/LaunchAgents"
cp "$STAGE/bin/imed" "$APP_STAGE/Contents/MacOS/imed"
chmod +x "$APP_STAGE/Contents/MacOS/imed"
cp "$ROOT/data/dicts/rime-ice.tsdict" "$APP_STAGE/Contents/Resources/data/dicts/"
cp "$ROOT/data/dicts/LICENSE.rime-ice.json" "$APP_STAGE/Contents/Resources/data/dicts/"
if [[ -f "$ROOT/data/dicts/rime-ice.tsdict.meta.json" ]]; then
  cp "$ROOT/data/dicts/rime-ice.tsdict.meta.json" "$APP_STAGE/Contents/Resources/data/dicts/"
fi

cat > "$APP_STAGE/Contents/Resources/LaunchAgents/com.surteam.tsime.imed.plist" <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.surteam.tsime.imed</string>
  <key>ProgramArguments</key>
  <array>
    <string>__IMED__</string>
  </array>
  <key>EnvironmentVariables</key>
  <dict>
    <key>TSIME_SCHEME</key>
    <string>xiaohe</string>
    <key>TSIME_DICT</key>
    <string>__DICT__</string>
  </dict>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>__STDOUT__</string>
  <key>StandardErrorPath</key>
  <string>__STDERR__</string>
</dict>
</plist>
EOF

echo "==> create Install TSInputMethod.app"
rm -rf "$INSTALLER_STAGE"
mkdir -p "$INSTALLER_STAGE/Contents/MacOS" "$INSTALLER_STAGE/Contents/Resources"
cp -R "$APP_STAGE" "$INSTALLER_STAGE/Contents/Resources/TSInputMethod.app"

cp "$APP_STAGE/Contents/Resources/AppIcon.icns" "$INSTALLER_STAGE/Contents/Resources/AppIcon.icns"

cat > "$INSTALLER_STAGE/Contents/Info.plist" <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>install</string>
  <key>CFBundleIconFile</key>
  <string>AppIcon.icns</string>
  <key>CFBundleIdentifier</key>
  <string>com.surteam.TSInputMethod.Installer</string>
  <key>CFBundleName</key>
  <string>Install TSInputMethod</string>
  <key>CFBundleDisplayName</key>
  <string>Install TSInputMethod</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleVersion</key>
  <string>0.1.1</string>
  <key>LSUIElement</key>
  <false/>
</dict>
</plist>
EOF

cat > "$INSTALLER_STAGE/Contents/MacOS/install" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

SELF_DIR="$(cd "$(dirname "$0")" && pwd)"
SRC_APP="$SELF_DIR/../Resources/TSInputMethod.app"
DEST_DIR="/Library/Input Methods"
DEST_APP="$DEST_DIR/TSInputMethod.app"
USER_DEST="$HOME/Library/Input Methods/TSInputMethod.app"
AGENT_LABEL="com.surteam.tsime.imed"
AGENT_DIR="$HOME/Library/LaunchAgents"
AGENT_PLIST="$AGENT_DIR/$AGENT_LABEL.plist"
LOG_DIR="$HOME/Library/Logs"
TEMPLATE="$SRC_APP/Contents/Resources/LaunchAgents/$AGENT_LABEL.plist"

if [[ ! -d "$SRC_APP" ]]; then
  osascript -e 'display alert "TSInputMethod installer" message "Embedded IME app is missing from the installer bundle." as critical' || true
  exit 1
fi

mkdir -p "$AGENT_DIR" "$LOG_DIR" "$HOME/Library/Input Methods"

INSTALL_TMP="$(mktemp -d)/TSInputMethod.app"
cp -R "$SRC_APP" "$INSTALL_TMP"
osascript <<COPY
do shell script "mkdir -p '/Library/Input Methods' && rm -rf '/Library/Input Methods/TSInputMethod.app' && cp -R '$INSTALL_TMP' '/Library/Input Methods/TSInputMethod.app' && xattr -cr '/Library/Input Methods/TSInputMethod.app' && chown -R root:wheel '/Library/Input Methods/TSInputMethod.app' && rm -rf '$USER_DEST'" with administrator privileges with prompt "TSInputMethod needs administrator access to install into /Library/Input Methods. Touch ID or password is OK."
COPY

IMED="$DEST_APP/Contents/MacOS/imed"
IME_BIN="$DEST_APP/Contents/MacOS/TSInputMethod"
DICT="$DEST_APP/Contents/Resources/data/dicts/rime-ice.tsdict"
STDOUT="$LOG_DIR/tsime-imed.log"
STDERR="$LOG_DIR/tsime-imed.err.log"

if [[ ! -x "$IMED" ]]; then
  osascript -e 'display alert "TSInputMethod installer" message "Embedded imed binary is missing." as critical' || true
  exit 1
fi
if [[ ! -f "$DICT" ]]; then
  osascript -e 'display alert "TSInputMethod installer" message "Embedded dictionary is missing." as critical' || true
  exit 1
fi

if [[ -x "$IME_BIN" ]]; then
  "$IME_BIN" --register-input-source >/tmp/tsime-register.log 2>&1 || true
fi

UID_NUM="$(id -u)"
OLD_AGENT_LABEL="com.surteam.tsime.imed"
if launchctl print "gui/$UID_NUM/$OLD_AGENT_LABEL" >/dev/null 2>&1; then
  launchctl bootout "gui/$UID_NUM/$OLD_AGENT_LABEL" >/dev/null 2>&1 || true
fi
rm -f "$AGENT_DIR/$OLD_AGENT_LABEL.plist"
if launchctl print "gui/$UID_NUM/$AGENT_LABEL" >/dev/null 2>&1; then
  launchctl bootout "gui/$UID_NUM/$AGENT_LABEL" >/dev/null 2>&1 || true
fi

sed \
  -e "s|__IMED__|$IMED|g" \
  -e "s|__DICT__|$DICT|g" \
  -e "s|__STDOUT__|$STDOUT|g" \
  -e "s|__STDERR__|$STDERR|g" \
  "$TEMPLATE" > "$AGENT_PLIST"

launchctl bootstrap "gui/$UID_NUM" "$AGENT_PLIST"
launchctl enable "gui/$UID_NUM/$AGENT_LABEL" >/dev/null 2>&1 || true
launchctl kickstart -k "gui/$UID_NUM/$AGENT_LABEL" >/dev/null 2>&1 || true

osascript <<OSA
display dialog "TSInputMethod installed to:\n$DEST_APP\n\nDaemon LaunchAgent loaded (Xiaohe default).\nHans input mode registered and enabled.\n\nIf the menu bar does not show TSInputMethod yet, reopen Keyboard settings or log out/in." buttons {"OK"} default button 1 with title "TSInputMethod"
OSA
EOF
chmod +x "$INSTALLER_STAGE/Contents/MacOS/install"

cat > "$STAGE/README.md" <<'EOF'
# TSInputMethod macOS trial

Primary install: open the DMG and double-click **Install TSInputMethod**.

That installer:

1. Copies `TSInputMethod.app` into `/Library/Input Methods` (admin prompt)
2. Registers and enables the Hans input mode (`TISRegisterInputSource` + enable parent/Hans)
3. Installs a LaunchAgent that auto-starts embedded `imed` from `/Library`
4. Defaults `TSIME_SCHEME=xiaohe`

Then enable the input source in System Settings → Keyboard → Input Sources → 简体中文.

Developer fallback (not required for trial):

```bash
bash scripts/package_macos.sh
```
EOF

echo "==> codesign with SurTeam identity"
resolve_sign_identity
echo "using identity: $SIGN_IDENTITY"
sign_app_bundle "$APP_STAGE"
rm -rf "$INSTALLER_STAGE/Contents/Resources/TSInputMethod.app"
cp -R "$APP_STAGE" "$INSTALLER_STAGE/Contents/Resources/TSInputMethod.app"
sign_app_bundle "$INSTALLER_STAGE"

echo "==> create DMG"
cp -R "$INSTALLER_STAGE" "$DMG_STAGE/Install TSInputMethod.app"
cat > "$DMG_STAGE/README.txt" <<'EOF'
Double-click "Install TSInputMethod" to install the IME into /Library/Input Methods, register it, and start the daemon.
Then add TSInputMethod under System Settings → Keyboard → Input Sources → 简体中文.
Default scheme: Xiaohe (flypy).
Signed with SurTeam identity from the build machine keychain.
EOF

rm -f "$DMG_PATH"
hdiutil create \
  -volname "TSInputMethod" \
  -srcfolder "$DMG_STAGE" \
  -ov \
  -format UDZO \
  "$DMG_PATH" >/dev/null

codesign --force --sign "$SIGN_IDENTITY" "$DMG_PATH" || true

echo "==> verify signatures"
codesign --verify --verbose=2 "$APP_STAGE"
codesign --verify --verbose=2 "$APP_STAGE/Contents/MacOS/imed"
codesign --verify --verbose=2 "$INSTALLER_STAGE"
codesign --display --verbose=2 "$APP_STAGE" 2>&1 | sed -n '1,20p'
codesign --display --verbose=2 "$APP_STAGE/Contents/MacOS/imed" 2>&1 | sed -n '1,20p'

echo "==> optional zip fallback (secondary)"
ARCHIVE="$DIST/TSInputMethod-macos.zip"
rm -f "$ARCHIVE"
(cd "$DIST" && zip -qr "TSInputMethod-macos.zip" "TSInputMethod-macos")

echo "primary package: $DMG_PATH"
echo "stage dir: $STAGE"
echo "secondary zip: $ARCHIVE"
echo "sign identity: $SIGN_IDENTITY"
echo "note: SurTeamCode is a self-signed codesign identity (issuer=SurTeamCode), not Apple Developer ID Application; Gatekeeper may still warn for downloaded artifacts; notarization is not claimed."

