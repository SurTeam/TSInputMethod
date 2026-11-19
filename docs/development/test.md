# Test

**Status:** core Go unit tests exist for ipc, session, dictionary, pinyin, ranking, inference, and engine ranking wiring.

## Layers

| Layer | Location | Intent |
| --- | --- | --- |
| Unit | `internal/...` | ipc framing, session edit keys, dictionary lookup, continuous pinyin, ranking/inference |
| Integration | `tests/integration` | daemon + client over real UDS/pipe (not yet filled) |
| Compatibility | `tests/compatibility` | cross-OS contract checks (not yet filled) |
| Performance | `tests/performance` + `*_bench_test.go` | latency bench entry points |

## Commands

```bash
export GOSUMDB=off
go test ./...
go test -bench=. -benchmem ./internal/ipc ./internal/pinyin ./tests/performance
```

Observed pass packages for s0/s1: `internal/dictionary`, `internal/pinyin`, `internal/session`, `internal/ipc`.

Expected s2 packages: `internal/ranking`, `internal/inference`, `internal/engine` (plus prior).

## Honesty rule

Never record pass status without the command output or CI evidence.
