#!/usr/bin/env bash
set -euo pipefail
PREFIX="${PREFIX:-/usr}"
ADDON_DIR="${FCITX5_ADDON_DIR:-$PREFIX/lib/fcitx5}"
IM_DIR="${FCITX5_IM_DIR:-$PREFIX/share/fcitx5/inputmethod}"
ADDON_CONF_DIR="${FCITX5_ADDON_CONF_DIR:-$PREFIX/share/fcitx5/addon}"
sudo rm -f "$ADDON_DIR/tsime.so" "$IM_DIR/tsime.conf" "$ADDON_CONF_DIR/tsime.conf"
if command -v fcitx5-remote >/dev/null 2>&1; then
  fcitx5-remote -r || true
fi
echo "removed Fcitx5 addon tsime"
