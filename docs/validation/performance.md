# Performance validation

**Status:** bench entry points exist; recorded timings and dictionary residency samples are host-local, not SLA claims.

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
| Daemon RSS after dict load | closer to on-disk `.tsdict` than 10× Go expansion | pre-opt live `tsimed` RSS ≈516 MiB; post-opt LoadFile HeapAlloc Δ ≈34.7 MiB (see below) |
| Dict LoadFile HeapAlloc (full rime-ice) | ≪ prior ~418 MiB eager expansion | after compact mmap runtime: ≈34.7 MiB |

## Dictionary residency (rime-ice.tsdict)

Same compiled index: ≈43 MiB on disk, `entry_count=1910826`.

| Stage | Method | Result |
| --- | --- | --- |
| Before (eager `[]Entry` + `map[string][]int`) | `LoadFile` then `runtime.GC` + `MemStats` | HeapAlloc Δ ≈ **418.0 MiB**; live `tsimed` RSS ≈ **516 MiB** |
| After (mmap bytes + `entryOff` / sorted `keyBlob` / `postings`) | same probe on compact runtime | HeapAlloc Δ ≈ **34.7 MiB**; `compact=true`; approximate index bytes ≈7.3 MiB offsets + ≈12.9 MiB keys + ≈7.3 MiB postings |

Notes:

- mmap file pages are outside Go `HeapAlloc`; compare HeapAlloc for Go-owned structures, and process RSS after shipping a rebuilt daemon for end-to-end residency.
- Load still builds a temporary grouping map that is discarded before return; steady-state no longer retains per-entry Go strings.
- Design mirrors SQLite-style sorted key + offset/postings + on-demand decode over a memory-mapped table.

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
- Dictionary residency: temporary `LoadFile` + `runtime.MemStats` probe against full `rime-ice.tsdict` (package test cwd may resolve the installed app Resources copy).
- Re-evaluate FlatBuffers / Cap’n Proto only with comparative benches (see ADR 0001).
