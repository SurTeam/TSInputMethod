# Module: ranking

**Status:** implemented / compiled / auto-tested (unit + personal benches). Real neural backends not shipped. Online SGD + FusedRanker (Base+UserFrequency+SGD within compatible groups) unit-tested; defaults-on settings live in `internal/personal`. Eval CLI: `cmd/tsime-personal-eval`.

## Goals

- Provide a stable frequency FastRanker for the hot path.
- Expose a Ranking Coordinator that can inject an inference backend.
- Support asynchronous rerank updates gated by candidate `generation` so stale results are discarded.
- Keep no-model / backend-failure paths returning base candidates.
- Provide pure-Go Feature-Hashing pairwise SGD personalization with short-term bias cache and a bounded async update queue.

## Boundaries

- **In scope:** ranker interface, coordinator, sync/async orchestration, generation discard helper, personal feature hashing, sparse SGD, short-term cache, bounded feedback queue, separate personal-model persistence.
- **Out of scope:** ONNX/Core ML runtimes, platform candidate UI refresh, full settings UI.

## Source locations

| Path | Role |
| --- | --- |
| `internal/ranking/types.go` | Shared request/candidate types |
| `internal/ranking/fast.go` | Frequency FastRanker |
| `internal/ranking/coordinator.go` | Sync rank + optional async backend |
| `internal/ranking/features.go` | Shared FNV-1a feature extractor |
| `internal/ranking/sgd.go` | Sparse pairwise logistic SGD PersonalModel |
| `internal/ranking/cache.go` | Bounded time-decay short-term preference cache |
| `internal/ranking/queue.go` | Single-worker bounded update queue |
| `internal/ranking/persistence.go` | Separate personal_model.json load/save/reset |
| `internal/ranking/coordinator_test.go` | Inject / degrade / stale-update tests |
| `internal/ranking/fusion.go` | FusedRanker: Base+UserFreq+SGD within Coverage/Incomplete groups |
| `internal/ranking/personal_test.go` | Hash stability, pairwise lift, caps, cache, queue, persist |
| `internal/ranking/lifecycle_test.go` | Queue drain on close, trim headroom, concurrent score/update/save |
| `internal/personal/settings.go` | Defaults-on personalization toggles (no env vars) |
| `internal/engine` | `PinyinEngine` wires fused coordinator + feedback queue; ClearPersonalWeights / ClearPersonalData / Close |
| `cmd/tsime-personal` | Personal model path / status / clear-weights CLI |

## Design notes

1. Decode always produces base candidates via FastRanker (`Rank` / `RankSync`) without waiting on models. FastRanker keeps coverage (syllable-prefix length) ahead of raw score so progressive shorter prefixes stay visible, and keeps same-coverage exact candidates contiguous so FusedRanker / context personalization can reorder inside that group.
2. When a backend and update handler are set, `Rank` may start async inference; callers apply updates only through `ApplyUpdate(currentGeneration, update)`.
3. `RankWithContext` is the synchronous injection path used by tests and future blocking backends; failures fall back to the base list.
4. `PersonalModel` uses the same `ExtractFeatures` for Score and PairwiseUpdate (FNV-1a, dim 2^18). Updates apply LR/L2/grad-clip/weight-cap. Short-term cache boosts recent corrections with TTL decay. `UpdateQueue.Enqueue` never blocks; full queues drop. Persistence stores weights only (no raw context history) at `UserConfigDir/TSInputMethod/personal_model.json`; incompatible dimension/schema/hash loads are rejected. `ResetPersonalization` clears weights and cache.
5. `FusedRanker` runs FastRanker first, then reorders only inside same Coverage/Incomplete/Composed groups using `log1p(base) + userFreqWeight*log1p(usage) + sgdWeight*PersonalScore`. Disabling personalization restores FastRanker-only order. Session accepted feedback drains into `UpdateQueue` via `FeedbackSink`.
6. **Capacity / eviction:** when weights exceed `MaxWeights` (default 50000), the smallest-magnitude ones are dropped down to 90% of the cap (sort-based, so the cost is amortized instead of paid on every update); short-term cache capacity 256 with TTL decay and LRU eviction (expired entries are evicted on write, not under the read lock); update queue capacity 128 with non-blocking drop. `UpdateQueue.Close` rejects new feedback, drains what is already queued, and waits for the worker. `Save` clears the model's unsaved-changes flag only if no update arrived during the write. Persistence never stores raw context strings.
7. **Code key:** `PinyinEngine.EnqueueFeedback` normalizes the raw double-pinyin code with the active scheme (`pinyin.NormalizeInput`) so training features use the same `code_candidate` key that `FusedRanker` scores with (`analyzed.Joined`).
8. **Eval:** `go run ./cmd/tsime-personal-eval` replays deterministic fixtures predict→observe→train for Fast / UserFreq / SGD combinations and reports Top-1/5, MRR, rank deltas, false promotions, `est_model_bytes` (weights×16 + cache×64 + usage map), `heap_alloc_b` (`runtime.MemStats`), and latency (fixture metrics only).
9. **Clear:** `ResetPersonalization` + `cmd/tsime-personal clear-weights`; PrivacyMode disables scoring/enqueue (see security principles).

## Related

- [../inference/index.md](../inference/index.md)
- [../engine/index.md](../engine/index.md)
- [../pinyin/index.md](../pinyin/index.md)
- [../session/index.md](../session/index.md)
