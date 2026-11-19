#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BUILD="$ROOT/build"
PREFIX="${PREFIX:-/usr}"
ADDON_DIR="${FCITX5_ADDON_DIR:-$PREFIX/lib/fcitx5}"
IM_DIR="${FCITX5_IM_DIR:-$PREFIX/share/fcitx5/inputmethod}"
ADDON_CONF_DIR="${FCITX5_ADDON_CONF_DIR:-$PREFIX/share/fcitx5/addon}"
if [[ ! -f "$BUILD/tsime.so" && ! -f "$BUILD/tsime" ]]; then
  echo "missing built addon under $BUILD — configure/build with Fcitx5 present" >&2
  exit 1
fi
MODULE="$BUILD/tsime.so"
if [[ ! -f "$MODULE" ]]; then
  MODULE="$BUILD/tsime"
fi
sudo install -d "$ADDON_DIR" "$IM_DIR" "$ADDON_CONF_DIR"
sudo install -m 755 "$MODULE" "$ADDON_DIR/tsime.so"
sudo install -m 644 "$ROOT/tsime.conf" "$IM_DIR/tsime.conf"
sudo install -m 644 "$ROOT/tsime-addon.conf" "$ADDON_CONF_DIR/tsime.conf"
if command -v fcitx5-remote >/dev/null 2>&1; then
  fcitx5-remote -r || true
fi
echo "installed Fcitx5 addon tsime; enable TSInputMethod in Fcitx5 config"
echo "ensure tsimed listens on TSIME_SOCKET or \$XDG_RUNTIME_DIR/tsime.sock"
