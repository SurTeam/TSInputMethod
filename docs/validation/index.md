# Validation index

Honest status only. Do not mark unrun tests as passed. Distinguish implemented / compiled / auto-tested / real-env.

## Legend

See [../index.md](../index.md) for implemented / compiled / auto-tested / real-env.

## Component snapshot

| Component | implemented | compiled | auto-tested | real-env |
| --- | --- | --- | --- | --- |
| Documentation foundation | yes | N/A | N/A | N/A |
| IPC ADR / Proto | yes | yes | N/A | N/A |
| Go daemon (`cmd/imed`) | yes | yes | smoke start | no |
| `internal/ipc` UDS | yes | yes | yes | no |
| `internal/ipc` Named Pipe | yes (windows tags) | not on this host | no | no |
| `platforms/common` libtsime_ipc | yes | yes (`cmake --build`) | frame smoke ok | no |
| `internal/session` / dictionary / pinyin | yes | yes | yes | no |
| Fogong / rime-ice dictBuilder | yes | N/A | build ran | no |
| macOS IMKit adapter | yes | yes (`make app`) | no | **no** |
| Windows TSF adapter | yes (source + scripts) | **no** (needs MSVC) | no | **no** |
| Linux Fcitx5 adapter | yes (source + scripts) | **no** (needs Fcitx5) | no | **no** |
| Ranking / Inference | yes (API + FastRanker + stubs) | yes | yes (unit) | no |
| Xiaohe / flypy shuangpin | yes | yes | yes (unit) | no |
| macOS trial package (`scripts/package_macos.sh`) | yes | yes when packaged | N/A | user installs |

## Observed this turn (s2)

- `go test ./internal/ranking ./internal/inference ./internal/engine ./internal/pinyin ./internal/session ./internal/dictionary ./internal/ipc` → ok
- Benches: `BenchmarkWriteReadFrame` ~65 ns/op; `BenchmarkDecodeNihao` ~2082 ns/op; `BenchmarkRankingFastPath` ~1557 ns/op (Apple M1 Pro samples)
- No real neural backend weights; UnavailableBackend / nil backend degrade to frequency ranking
- Platform adapters unchanged; real-env still **no**

## Pages

- [compatibilityMatrix.md](compatibilityMatrix.md)
- [performance.md](performance.md)
- [knownIssues.md](knownIssues.md)
