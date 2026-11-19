#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_SRC="$ROOT/build/TSInputMethod.app"
DEST="$HOME/Library/Input Methods/TSInputMethod.app"
if [[ ! -d "$APP_SRC" ]]; then
  echo "missing built app: $APP_SRC" >&2
  exit 1
fi
mkdir -p "$HOME/Library/Input Methods"
rm -rf "$DEST"
cp -R "$APP_SRC" "$DEST"
echo "installed $DEST"
echo "Enable TSInputMethod in System Settings → Keyboard → Input Sources."
echo "Ensure imed is running and reachable via TSIME_SOCKET or /tmp/tsime-\$UID.sock."
