# tools/dictBuilder

Offline builder that converts rime-ice `*.dict.yaml` tables into a compiled `.tsdict` index for `internal/dictionary`.

## Usage

Default full Chinese tables (雾凇 `rime_ice.dict.yaml` import set including optional 41448):

```bash
python3 tools/dictBuilder/build_dict.py \
  --source-dir third_party/rime-ice \
  --inputs 8105.dict.yaml 41448.dict.yaml base.dict.yaml ext.dict.yaml tencent.dict.yaml others.dict.yaml \
  --output data/dicts/rime-ice.tsdict \
  --license-out data/dicts/LICENSE.rime-ice.json
```

Put `41448` immediately after `8105` so common characters keep 8105 weights and rare glyphs trail.

Smaller 8105+base only build (legacy / faster):

```bash
python3 tools/dictBuilder/build_dict.py \
  --source-dir third_party/rime-ice \
  --inputs 8105.dict.yaml base.dict.yaml \
  --output data/dicts/rime-ice.tsdict \
  --license-out data/dicts/LICENSE.rime-ice.json
```

Fixture path for tests:

```bash
python3 tools/dictBuilder/build_dict.py \
  --source-dir testdata/dicts \
  --inputs fixture.dict.yaml \
  --output testdata/dicts/fixture.tsdict \
  --license-out testdata/dicts/LICENSE.fixture.json
```

## Outputs

- `.tsdict`: binary index (`TSIMDI01`)
- `.tsdict.meta.json`: build metadata
- `LICENSE.*.json`: source project, versions, GPL-3.0 redistribution note, no Rime Runtime dependency

## Emoji

Emoji is not compiled into `.tsdict`. Ship fogong OpenCC map `data/dicts/emoji.txt` (from `third_party/rime-ice/opencc/emoji.txt`) next to the index; `cmd/tsimed` loads it via `TSIME_EMOJI` or the same directory lookup and expands matching Chinese candidates.

## Status

Implemented for 8105 + 41448 + base + ext + tencent + others.
