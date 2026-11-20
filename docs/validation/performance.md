# Performance validation

**Status:** bench entry points exist; recorded timings are host-local samples, not SLA claims.

## Budgets (targets, not guarantees)

| Metric | Target (aspirational) | Measured |
| --- | --- | --- |
| Key → candidates IPC+engine (p95) | < 10ms | not measured end-to-end |
| Frame encode/decode overhead | TBD | sample on M1 Pro: ~65 ns/op (`BenchmarkWriteReadFrame`) |
| Pinyin decode (fixture) | TBD | sample: `nihao` ~2082 ns/op; `shijie` ~1599 ns/op |
| Ranking fast path | TBD | sample: ~1557 ns/op (`BenchmarkRankingFastPath`) |
| PersonalModel.Score (trained) | TBD | sample M1 Pro: ~827.5 ns/op, 112 B/op (`BenchmarkPersonalScore`) |
| FusedRanker (3 cands) | TBD | sample M1 Pro: ~5421 ns/op, 1296 B/op (`BenchmarkFusedRanker`) |
| Personal eval est_model_bytes (fixture) | TBD | Fast 0; UserFreq 48; SGD 464 (21 weights); UserFreq+SGD 320 |
| Personal eval heap_alloc_b (fixture) | TBD | ~224–239 KiB after GC (`runtime.MemStats.HeapAlloc`) |
| Daemon RSS after dict load | TBD | not measured |

## Capacity / eviction (personalization)

| Resource | Limit | Policy |
| --- | --- | --- |
| SGD sparse weights | `MaxWeights` default 50000 | drop smallest absolute weights |
| Short-term preference cache | capacity 256, TTL 5m | LRU + TTL decay |
| Feedback update queue | capacity 128 | non-blocking drop when full |
| EffectiveContext | window 48 / max 64 Unicode | tail clamp |

## Bench entry commands

```bash
export GOSUMDB=off
go test -bench=BenchmarkWriteReadFrame -benchmem ./internal/ipc
go test -bench=BenchmarkDecode -benchmem ./internal/pinyin
go test -bench=BenchmarkRankingFastPath -benchmem ./tests/performance
go test -bench='BenchmarkPersonalScore|BenchmarkFusedRanker' -benchmem ./internal/ranking
go run ./cmd/tsime-personal-eval
```

Record machine, OS, commit SHA, and full command output before promoting any number into a claim. Fixture eval metrics are not production accuracy.

## Method

- Package benches live next to modules (`*_bench_test.go`) plus `tests/performance`.
- Re-evaluate FlatBuffers / Cap’n Proto only with comparative benches (see ADR 0001).
