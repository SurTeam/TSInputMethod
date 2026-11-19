#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST="$ROOT/dist/macos"
STAGE="$DIST/TSInputMethod-macos"
APP_SRC="$ROOT/platforms/macos/build/TSInputMethod.app"

export GOSUMDB="${GOSUMDB:-off}"

echo "==> build imed"
mkdir -p "$STAGE/bin" "$STAGE/data/dicts" "$STAGE/scripts"
(cd "$ROOT" && go build -o "$STAGE/bin/imed" ./cmd/imed)

echo "==> copy dictionary"
cp "$ROOT/data/dicts/rime-ice.tsdict" "$STAGE/data/dicts/"
cp "$ROOT/data/dicts/LICENSE.rime-ice.json" "$STAGE/data/dicts/"
if [[ -f "$ROOT/data/dicts/rime-ice.tsdict.meta.json" ]]; then
  cp "$ROOT/data/dicts/rime-ice.tsdict.meta.json" "$STAGE/data/dicts/"
fi

echo "==> build macOS IMKit app"
(cd "$ROOT/platforms/macos" && make app)
rm -rf "$STAGE/TSInputMethod.app"
cp -R "$APP_SRC" "$STAGE/TSInputMethod.app"

cat > "$STAGE/scripts/install.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$HOME/Library/Input Methods/TSInputMethod.app"
mkdir -p "$HOME/Library/Input Methods"
rm -rf "$DEST"
cp -R "$ROOT/TSInputMethod.app" "$DEST"
echo "installed $DEST"
echo "Next:"
echo "  1) Start daemon: $ROOT/scripts/run_imed.sh"
echo "  2) System Settings → Keyboard → Input Sources → add TSInputMethod"
echo "  3) If Gatekeeper blocks: xattr -dr com.apple.quarantine \"$DEST\""
echo "Default scheme is Xiaohe (flypy). Override with TSIME_SCHEME=quanpin."
EOF
chmod +x "$STAGE/scripts/install.sh"

cat > "$STAGE/scripts/uninstall.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
DEST="$HOME/Library/Input Methods/TSInputMethod.app"
rm -rf "$DEST"
echo "removed $DEST"
EOF
chmod +x "$STAGE/scripts/uninstall.sh"

cat > "$STAGE/scripts/run_imed.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export TSIME_SCHEME="${TSIME_SCHEME:-xiaohe}"
export TSIME_DICT="${TSIME_DICT:-$ROOT/data/dicts/rime-ice.tsdict}"
export TSIME_SOCKET="${TSIME_SOCKET:-/tmp/tsime-$UID.sock}"
exec "$ROOT/bin/imed"
EOF
chmod +x "$STAGE/scripts/run_imed.sh"

cat > "$STAGE/README.md" <<'EOF'
# TSInputMethod macOS trial package

Scheme default: **Xiaohe / flypy** (`TSIME_SCHEME=xiaohe`).

## Install

```bash
bash scripts/install.sh
```

## Run daemon (required while typing)

```bash
bash scripts/run_imed.sh
```

Socket: `/tmp/tsime-$UID.sock` (override with `TSIME_SOCKET`).

## Enable input source

System Settings → Keyboard → Input Sources → add **TSInputMethod**.

Unsigned local build: if blocked, run:

```bash
xattr -dr com.apple.quarantine "$HOME/Library/Input Methods/TSInputMethod.app"
```

## Uninstall

```bash
bash scripts/uninstall.sh
```

## Notes

- Chinese mode uses daemon decode; English menu mode does not forward keys.
- Candidate UI is minimal (marked text / commit); host polish is ongoing.
- real-env host verification is performed by you after install.
EOF

ARCHIVE="$DIST/TSInputMethod-macos.zip"
rm -f "$ARCHIVE"
(cd "$DIST" && zip -qr "TSInputMethod-macos.zip" "TSInputMethod-macos")
echo "package ready: $ARCHIVE"
echo "stage dir: $STAGE"
