# tools/dictBuilder

Offline builder that converts rime-ice `*.dict.yaml` tables into a compiled `.tsdict` index for `internal/dictionary`.

## Usage

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

## Status

Implemented for 8105 + base. Optional `ext` can be passed via `--inputs`.
