# Module: dictionary

**Status:** implemented / auto-tested (fixture + compiled rime-ice index load path).

## Goals

- Load compiled Chinese dictionary indexes produced by `tools/dictBuilder`.
- Support exact syllable-sequence lookup and prefix lookup for continuous pinyin.
- Keep runtime free of Rime Runtime.

## Boundaries

- **In scope:** binary index load, syllable set, entry lookup by pinyin key, weight-ordered results.
- **Out of scope:** YAML parsing at runtime, ranking models, platform UI.

## Source locations

| Path | Role |
| --- | --- |
| `internal/dictionary` | Runtime loader and lookup |
| `tools/dictBuilder/build_dict.py` | Offline YAML → `.tsdict` |
| `data/dicts/rime-ice.tsdict` | Built 8105+base index |
| `data/dicts/LICENSE.rime-ice.json` | License / provenance record |
| `third_party/rime-ice/` | Cached upstream YAML + LICENSE |

## Related

- [../pinyin/index.md](../pinyin/index.md)
- [../engine/index.md](../engine/index.md)
