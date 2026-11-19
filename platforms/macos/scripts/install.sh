#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_SRC="$ROOT/build/TSInputMethod.app"
DEST="/Library/Input Methods/TSInputMethod.app"
USER_DEST="$HOME/Library/Input Methods/TSInputMethod.app"
AGENT_LABEL="com.surteam.tsimed"
if [[ ! -d "$APP_SRC" ]]; then
  echo "missing built app: $APP_SRC" >&2
  exit 1
fi
INSTALL_TMP="$(mktemp -d)/TSInputMethod.app"
cp -R "$APP_SRC" "$INSTALL_TMP"
osascript <<OSA
do shell script "mkdir -p '/Library/Input Methods' && rm -rf '/Library/Input Methods/TSInputMethod.app' && cp -R '$INSTALL_TMP' '/Library/Input Methods/TSInputMethod.app' && xattr -cr '/Library/Input Methods/TSInputMethod.app' && chown -R root:wheel '/Library/Input Methods/TSInputMethod.app' && rm -rf '$USER_DEST'" with administrator privileges
OSA
killall TSInputMethod >/dev/null 2>&1 || true
if [[ -x "$DEST/Contents/MacOS/TSInputMethod" ]]; then
  "$DEST/Contents/MacOS/TSInputMethod" --register-input-source || true
fi
UID_NUM="$(id -u)"
if launchctl print "gui/$UID_NUM/$AGENT_LABEL" >/dev/null 2>&1; then
  launchctl kickstart -k "gui/$UID_NUM/$AGENT_LABEL" >/dev/null 2>&1 || true
fi
killall TSInputMethod >/dev/null 2>&1 || true
echo "installed $DEST"
echo "Hans mode registered/enabled via TIS; IMKit process refreshed."
echo "If the menu bar is still stale, use the IME menu Restart item or reopen Keyboard settings; log out/in only as a last resort."
echo "Ensure tsimed is running and reachable via TSIME_SOCKET or /tmp/tsime-\$UID.sock."
