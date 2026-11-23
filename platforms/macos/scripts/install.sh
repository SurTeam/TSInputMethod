#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_SRC="$ROOT/build/TSInputMethod.app"
DEST="/Library/Input Methods/TSInputMethod.app"
USER_DEST="$HOME/Library/Input Methods/TSInputMethod.app"
AGENT_LABEL="com.surteam.tsimed"
ENTITLEMENTS="$ROOT/TSInputMethod.entitlements"
SIGN_IDENTITY="${TSIME_SIGN_IDENTITY:-}"
if [[ ! -d "$APP_SRC" ]]; then
  echo "missing built app: $APP_SRC" >&2
  exit 1
fi
if [[ -z "$SIGN_IDENTITY" ]]; then
  SIGN_IDENTITY="$(/usr/bin/security find-identity -v -p codesigning | /usr/bin/awk '/SurTeam/ {print; exit}' | /usr/bin/sed -E 's/.*"([^"]+)".*/\1/')"
fi
if [[ -z "$SIGN_IDENTITY" ]]; then
  echo "error: no SurTeam codesigning identity found; set TSIME_SIGN_IDENTITY" >&2
  exit 1
fi
INSTALL_TMP="$(mktemp -d)/TSInputMethod.app"
cp -R "$APP_SRC" "$INSTALL_TMP"
if [[ -x "$INSTALL_TMP/Contents/MacOS/tsimed" ]]; then
  /usr/bin/codesign --force --options runtime --timestamp=none --sign "$SIGN_IDENTITY" "$INSTALL_TMP/Contents/MacOS/tsimed"
fi
if [[ -f "$ENTITLEMENTS" ]]; then
  /usr/bin/codesign --force --options runtime --timestamp=none --entitlements "$ENTITLEMENTS" --sign "$SIGN_IDENTITY" "$INSTALL_TMP/Contents/MacOS/TSInputMethod"
  /usr/bin/codesign --force --deep --options runtime --timestamp=none --entitlements "$ENTITLEMENTS" --sign "$SIGN_IDENTITY" "$INSTALL_TMP"
else
  /usr/bin/codesign --force --deep --options runtime --timestamp=none --sign "$SIGN_IDENTITY" "$INSTALL_TMP"
fi
/usr/bin/codesign --verify --verbose=2 "$INSTALL_TMP"
if [[ "$(/usr/bin/id -u)" -eq 0 ]]; then
  /bin/mkdir -p "/Library/Input Methods"
  /bin/rm -rf "$DEST"
  /bin/cp -R "$INSTALL_TMP" "$DEST"
  /usr/bin/xattr -dr com.apple.quarantine "$DEST" || true
  /usr/bin/xattr -cr "$DEST" || true
  /usr/sbin/chown -R root:wheel "$DEST"
  CONSOLE_USER="$(/usr/bin/stat -f %Su /dev/console 2>/dev/null || true)"
  if [[ -z "$CONSOLE_USER" || "$CONSOLE_USER" == "root" || "$CONSOLE_USER" == "_mbsetupuser" ]]; then
    CONSOLE_USER="${SUDO_USER:-}"
  fi
  if [[ -n "$CONSOLE_USER" && "$CONSOLE_USER" != "root" ]]; then
    /bin/rm -rf "/Users/$CONSOLE_USER/Library/Input Methods/TSInputMethod.app" || true
  fi
else
  osascript <<OSA
do shell script "mkdir -p '/Library/Input Methods' && rm -rf '/Library/Input Methods/TSInputMethod.app' && cp -R '$INSTALL_TMP' '/Library/Input Methods/TSInputMethod.app' && /usr/bin/xattr -dr com.apple.quarantine '/Library/Input Methods/TSInputMethod.app' && /usr/bin/xattr -cr '/Library/Input Methods/TSInputMethod.app' && /usr/sbin/chown -R root:wheel '/Library/Input Methods/TSInputMethod.app' && rm -rf '$USER_DEST'" with administrator privileges
OSA
  CONSOLE_USER="$(/usr/bin/id -un)"
fi
/usr/bin/killall TSInputMethod >/dev/null 2>&1 || true
IME_BIN="$DEST/Contents/MacOS/TSInputMethod"
run_as_console_user() {
  local cmd="$1"
  if [[ "$(/usr/bin/id -u)" -eq 0 && -n "${CONSOLE_USER:-}" && "$CONSOLE_USER" != "root" ]]; then
    /usr/bin/su -m "$CONSOLE_USER" -c "$cmd" || true
  else
    bash -lc "$cmd" || true
  fi
}
if [[ -x "$IME_BIN" ]]; then
  run_as_console_user "\"$IME_BIN\" --register-input-source"
  run_as_console_user "\"$IME_BIN\" --enable-input-source"
fi
if [[ "$(/usr/bin/id -u)" -eq 0 && -n "${CONSOLE_USER:-}" && "$CONSOLE_USER" != "root" ]]; then
  UID_NUM="$(/usr/bin/id -u "$CONSOLE_USER")"
  AGENT_PLIST="/Users/$CONSOLE_USER/Library/LaunchAgents/$AGENT_LABEL.plist"
else
  UID_NUM="$(/usr/bin/id -u)"
  AGENT_PLIST="$HOME/Library/LaunchAgents/$AGENT_LABEL.plist"
fi
if [[ -f "$AGENT_PLIST" ]]; then
  if ! /bin/launchctl print "gui/$UID_NUM/$AGENT_LABEL" >/dev/null 2>&1; then
    /bin/launchctl bootstrap "gui/$UID_NUM" "$AGENT_PLIST" >/dev/null 2>&1 || true
    /bin/launchctl enable "gui/$UID_NUM/$AGENT_LABEL" >/dev/null 2>&1 || true
  fi
  /bin/launchctl kickstart -k "gui/$UID_NUM/$AGENT_LABEL" >/dev/null 2>&1 || true
fi
/usr/bin/killall TextInputMenuAgent TextInputSwitcher >/dev/null 2>&1 || true
/usr/bin/killall TSInputMethod >/dev/null 2>&1 || true
echo "installed $DEST"
echo "Hans mode registered/enabled via TIS as user ${CONSOLE_USER:-current}; IMKit/TextInput agents refreshed."
echo "If the menu bar is still stale, use the IME menu Restart item or reopen Keyboard settings; log out/in only as a last resort."
echo "Ensure tsimed is running and reachable via TSIME_SOCKET or /tmp/tsime-\$UID.sock."
