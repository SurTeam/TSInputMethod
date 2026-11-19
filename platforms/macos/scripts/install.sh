#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_SRC="$ROOT/build/TSInputMethod.app"
DEST="/Library/Input Methods/TSInputMethod.app"
USER_DEST="$HOME/Library/Input Methods/TSInputMethod.app"
if [[ ! -d "$APP_SRC" ]]; then
  echo "missing built app: $APP_SRC" >&2
  exit 1
fi
INSTALL_TMP="$(mktemp -d)/TSInputMethod.app"
cp -R "$APP_SRC" "$INSTALL_TMP"
osascript <<OSA
do shell script "mkdir -p '/Library/Input Methods' && rm -rf '/Library/Input Methods/TSInputMethod.app' && cp -R '$INSTALL_TMP' '/Library/Input Methods/TSInputMethod.app' && xattr -cr '/Library/Input Methods/TSInputMethod.app' && chown -R root:wheel '/Library/Input Methods/TSInputMethod.app' && rm -rf '$USER_DEST'" with administrator privileges
OSA
if [[ -x "$DEST/Contents/MacOS/TSInputMethod" ]]; then
  "$DEST/Contents/MacOS/TSInputMethod" --register-input-source || true
fi
echo "installed $DEST"
echo "Hans mode registered/enabled via TIS; reopen Keyboard settings or log out/in if the menu bar is stale."
echo "Ensure imed is running and reachable via TSIME_SOCKET or /tmp/tsime-\$UID.sock."
