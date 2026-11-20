# Module: dictionary

**Status:** implemented / auto-tested (fixture + compiled rime-ice index load path; user dictionary CRUD/persist/concurrency).

## Goals

- Load compiled Chinese dictionary indexes produced by `tools/dictBuilder`.
- Support exact syllable-sequence lookup and prefix lookup for continuous pinyin.
- Keep runtime free of Rime Runtime.
- Provide a local UserDictionary overlay (manual + auto-learned) that merges with the system dictionary before candidate prune.

## Boundaries

- **In scope:** binary index load, syllable set, entry lookup by pinyin key, weight-ordered results, user overlay CRUD/lookup, atomic JSON persistence, phrase auto-learn helpers.
- **Out of scope:** YAML parsing at runtime, ranking models / SGD, platform UI, OpenCC emoji expansion (see `internal/emoji` / engine).
- Built-in `.tsdict` remains read-only; user data lives under the OS user config directory.

## User dictionary

- Syllables are scheme-independent quanpin sequences validated via `SetValidSyllableFunc` (wired to `pinyin.IsValidSyllable` by the engine/CLI).
- APIs: Add/Remove/Update/Lookup/List/ClearLearned/ClearAll, exact + prefix lookup, MergeEntries for system dedupe/frequency fusion.
- Persistence: versioned JSON at `UserConfigDir/TSInputMethod/user_dictionary.json` (0600, temp+rename); CLI uses flock coordination + reload.
- Auto-learn: `PhraseLearner` promotes reliable repeated Han phrases with recoverable syllables after a configurable min-repeat threshold.
- Management CLI: `cmd/tsime-userdict` (`add|remove|update|lookup|list|clear|reload|path`).

## Source locations

| Path | Role |
| --- | --- |
| `internal/dictionary` | Runtime loader and lookup |
| `internal/dictionary/user_dictionary.go` | User overlay CRUD + merge |
| `internal/dictionary/user_storage.go` | Atomic JSON persistence |
| `internal/dictionary/user_learning.go` | Phrase auto-learn |
| `cmd/tsime-userdict` | Manual management CLI |
| `tools/dictBuilder/build_dict.py` | Offline YAML → `.tsdict` |
| `data/dicts/rime-ice.tsdict` | Built 8105+41448+base+ext+tencent+others index |
| `data/dicts/emoji.txt` | Shipped rime-ice OpenCC emoji map |
| `data/dicts/LICENSE.rime-ice.json` | License / provenance record |
| `third_party/rime-ice/` | Cached upstream YAML + LICENSE + opencc |

## Related

- [../pinyin/index.md](../pinyin/index.md)
- [../engine/index.md](../engine/index.md)
- [../session/index.md](../session/index.md)
