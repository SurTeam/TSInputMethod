# Module: engine

**Status:** implemented / compiled / auto-tested via session + pinyin packages.

## Goals

- Perform pinyin composition, conversion, and candidate ranking.
- Load and query dictionaries prepared by `tools/dictBuilder`.
- Expose a pure Go API consumed by the session module—no OS IME types.

## Boundaries

- **In scope:** dictionary-backed decode, candidate lists for a single session call.
- **Out of scope:** IPC, process management, platform UI, shipping neural model weights.

## Source locations

| Path | Role |
| --- | --- |
| `internal/engine` | `Engine` interface, `StubEngine`, `PinyinEngine` |
| `internal/pinyin` | Continuous segmentation and lattice decode |
| `internal/ranking` | FastRanker wired by `PinyinEngine` |
| `internal/inference` | Optional Backend injection point |
| `internal/dictionary` | Compiled index load/lookup |
| `tools/dictBuilder` | Offline dictionary build |
| `cmd/tsimed` | Loads `.tsdict` (or stub) at daemon start |

## Related

- [../dictionary/index.md](../dictionary/index.md)
- [../pinyin/index.md](../pinyin/index.md)
- [../ranking/index.md](../ranking/index.md)
- [../inference/index.md](../inference/index.md)
- [../../architecture/dataFlow.md](../../architecture/dataFlow.md)
