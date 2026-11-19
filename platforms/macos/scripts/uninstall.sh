#!/usr/bin/env bash
set -euo pipefail
USER_DEST="$HOME/Library/Input Methods/TSInputMethod.app"
SYS_DEST="/Library/Input Methods/TSInputMethod.app"
rm -rf "$USER_DEST"
if [[ -d "$SYS_DEST" ]]; then
  osascript <<OSA
do shell script "rm -rf '/Library/Input Methods/TSInputMethod.app'" with administrator privileges
OSA
fi
echo "removed $USER_DEST and $SYS_DEST (if present)"
echo "Remove the input source in System Settings if it remains listed."
