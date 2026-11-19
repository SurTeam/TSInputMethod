# Build

**Status:** Go daemon and dictionary builder are buildable.

## Daemon

```bash
export GOSUMDB=off   # only if local sumdb is broken
go build -o bin/tsimed ./cmd/tsimed
TSIME_DICT=data/dicts/rime-ice.tsdict ./bin/tsimed
```

Environment:

- `TSIME_DICT` — path to `.tsdict` (defaults to `data/dicts/rime-ice.tsdict` or fixture)
- `TSIME_SOCKET` — UDS path override

## Dictionary tooling

```bash
python3 tools/dictBuilder/build_dict.py \
  --source-dir third_party/rime-ice \
  --inputs 8105.dict.yaml 41448.dict.yaml base.dict.yaml ext.dict.yaml tencent.dict.yaml others.dict.yaml \
  --output data/dicts/rime-ice.tsdict \
  --license-out data/dicts/LICENSE.rime-ice.json
cp third_party/rime-ice/opencc/emoji.txt data/dicts/emoji.txt
```

Environment:

- `TSIME_EMOJI` — optional path to OpenCC-style `emoji.txt` (defaults next to the `.tsdict`)

See `tools/dictBuilder/README.md`.

## Planned adapters

| Platform | Expected build surface |
| --- | --- |
| macOS | Xcode / `xcodebuild` or documented script under `platforms/macos` |
| Windows | MSBuild / documented script under `platforms/windows` |
| Linux | CMake or meson as chosen by the Fcitx5 addon layout under `platforms/linux` |

## Verification

Record successful commands in [../validation/index.md](../validation/index.md). Do not mark **compiled** without a recorded build.
