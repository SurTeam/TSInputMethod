# Performance validation

**Status:** bench entry points exist; recorded timings are host-local samples, not SLA claims.

## Budgets (targets, not guarantees)

| Metric | Target (aspirational) | Measured |
| --- | --- | --- |
| Key → candidates IPC+engine (p95) | < 10ms | not measured end-to-end |
| Frame encode/decode overhead | TBD | sample on M1 Pro: ~65 ns/op (`BenchmarkWriteReadFrame`) |
| Pinyin decode (fixture) | TBD | sample: `nihao` ~2082 ns/op; `shijie` ~1599 ns/op |
| Ranking fast path | TBD | sample: ~1557 ns/op (`BenchmarkRankingFastPath`) |
| Daemon RSS after dict load | TBD | not measured |

## Bench entry commands

```bash
export GOSUMDB=off
go test -bench=BenchmarkWriteReadFrame -benchmem ./internal/ipc
go test -bench=BenchmarkDecode -benchmem ./internal/pinyin
go test -bench=BenchmarkRankingFastPath -benchmem ./tests/performance
```

Record machine, OS, commit SHA, and full command output before promoting any number into a claim.

## Method

- Package benches live next to modules (`*_bench_test.go`) plus `tests/performance`.
- Re-evaluate FlatBuffers / Cap’n Proto only with comparative benches (see ADR 0001).
