# Module: ranking

**Status:** implemented / compiled / auto-tested (unit). Real neural backends not shipped.

## Goals

- Provide a stable frequency FastRanker for the hot path.
- Expose a Ranking Coordinator that can inject an inference backend.
- Support asynchronous rerank updates gated by candidate `generation` so stale results are discarded.
- Keep no-model / backend-failure paths returning base candidates.

## Boundaries

- **In scope:** ranker interface, coordinator, sync/async orchestration, generation discard helper.
- **Out of scope:** training, ONNX/Core ML runtimes, platform candidate UI refresh.

## Source locations

| Path | Role |
| --- | --- |
| `internal/ranking/types.go` | Shared request/candidate types |
| `internal/ranking/fast.go` | Frequency FastRanker |
| `internal/ranking/coordinator.go` | Sync rank + optional async backend |
| `internal/ranking/coordinator_test.go` | Inject / degrade / stale-update tests |
| `internal/engine` | `PinyinEngine` wires coordinator on decode |

## Design notes

1. Decode always produces base candidates via FastRanker (`Rank` / `RankSync`) without waiting on models.
2. When a backend and update handler are set, `Rank` may start async inference; callers apply updates only through `ApplyUpdate(currentGeneration, update)`.
3. `RankWithContext` is the synchronous injection path used by tests and future blocking backends; failures fall back to the base list.

## Related

- [../inference/index.md](../inference/index.md)
- [../engine/index.md](../engine/index.md)
- [../pinyin/index.md](../pinyin/index.md)
