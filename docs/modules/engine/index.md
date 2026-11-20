# Module: engine

**Status:** implemented / compiled / auto-tested via session + pinyin packages.

## Goals

- Perform pinyin composition, conversion, and candidate ranking.
- Load and query dictionaries prepared by `tools/dictBuilder`.
- Expose a pure Go API consumed by the session module—no OS IME types.
- Accept optional EffectiveContext via `ContextDecoder.DecodeWithContext` and forward it on `ranking.Request.Context` so remaining-code ranking after soft-confirm sees updated context.
- Attach an optional UserDictionary overlay via `PinyinEngine.SetUserDictionary`; the decoder merges user hits with the system dictionary before prune so new words can enter candidates under real Xiaohe input.
- Honor PrivacyMode by no-oping `EnqueueFeedback` and `RecordSelectedUsage`; expose clear/reset helpers (`ClearPersonalWeights`, `ClearUserDictionary`, `ClearPersonalData`) and `Close` that stop the update queue and persist dirty personal data.

## Boundaries

- **In scope:** dictionary-backed decode, optional user overlay wiring, candidate lists for a single session call, context plumbing into ranking requests.
- **Out of scope:** IPC, process management, platform UI, shipping neural model weights, SGD weight updates.

## Source locations

| Path | Role |
| --- | --- |
| `internal/engine` | `Engine` / `ContextDecoder`, `StubEngine`, `PinyinEngine` |
| `internal/pinyin` | Continuous segmentation and lattice decode |
| `internal/ranking` | FastRanker wired by `PinyinEngine` |
| `internal/inference` | Optional Backend injection point |
| `internal/dictionary` | Compiled index load/lookup |
| `internal/emoji` | OpenCC-style Chinese→emoji map |
| `tools/dictBuilder` | Offline dictionary build |
| `cmd/tsimed` | Loads `.tsdict` + optional `emoji.txt` (or stub) at daemon start |

## Related

- [../dictionary/index.md](../dictionary/index.md)
- [../pinyin/index.md](../pinyin/index.md)
- [../ranking/index.md](../ranking/index.md)
- [../inference/index.md](../inference/index.md)
- [../../architecture/dataFlow.md](../../architecture/dataFlow.md)
