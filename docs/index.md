# TSInputMethod documentation index

Entry point for technical docs. Code is truth; keep these pages aligned with the repository.

## Verification status legend

| Status | Meaning |
| --- | --- |
| **implemented** | Source or docs exist in-tree for the described piece |
| **compiled** | Build succeeded for that piece on a documented target |
| **auto-tested** | Automated tests ran and passed with recorded command/result |
| **real-env** | Exercised on a real host IME stack (IMKit / TSF / Fcitx5), not only unit/integration stubs |

Do not promote a status without evidence. This foundation turn marks documentation pages as **implemented** only; runtime, adapters, and tests remain unimplemented until other work lands.

## Architecture

- [overview.md](architecture/overview.md) — Go daemon, binary IPC, platform adapters
- [processModel.md](architecture/processModel.md) — processes and IPC transport
- [dataFlow.md](architecture/dataFlow.md) — key input path
- [lifecycle.md](architecture/lifecycle.md) — daemon / session / adapter lifecycle
- [compatibility.md](architecture/compatibility.md) — platform compatibility notes

## Modules

- [ipc](modules/ipc/index.md) — framing, transport, protobuf contracts
- [session](modules/session/index.md) — per-client IME sessions
- [engine](modules/engine/index.md) — conversion / candidate engine
- [dictionary](modules/dictionary/index.md) — compiled rime-ice index load/lookup
- [pinyin](modules/pinyin/index.md) — continuous syllable segmentation and lattice decode
- [ranking](modules/ranking/index.md) — FastRanker + injectable coordinator + online SGD fusion
- [personal](modules/personal/index.md) — Defaults-on personalization settings / PrivacyMode
- [inference](modules/inference/index.md) — Backend interface and no-model stubs

## Platforms

- [macOS](platforms/macos/index.md) — InputMethodKit adapter; SurTeam-signed DMG OOBE; TIS Hans/`zh-Hans` register+enable+select verified (Settings Add browser / typing pending)
- [Windows](platforms/windows/index.md) — TSF adapter (source + scripts; compile on MSVC)
- [Linux](platforms/linux/index.md) — Fcitx5 adapter (source + scripts; needs Fcitx5)

## Principles

- [engineering.md](principles/engineering.md)
- [coding.md](principles/coding.md)
- [performance.md](principles/performance.md)
- [reliability.md](principles/reliability.md)
- [security.md](principles/security.md)
- [documentation.md](principles/documentation.md)
- [crossPlatform.md](principles/crossPlatform.md)

## Development

- [gettingStarted.md](development/gettingStarted.md)
- [build.md](development/build.md)
- [test.md](development/test.md)
- [workflow.md](development/workflow.md)
- [documentationReview.md](development/documentationReview.md)

## Decisions

- [decisions/index.md](decisions/index.md)
- [0001-ipc-protobuf-framed.md](decisions/0001-ipc-protobuf-framed.md)
- [0002-semantic-edit-keycodes.md](decisions/0002-semantic-edit-keycodes.md)
- [0003-custom-candidate-panel.md](decisions/0003-custom-candidate-panel.md)
- [0004-macos-settings-xdg-config.md](decisions/0004-macos-settings-xdg-config.md)
- [0005-macos-daemon-login-lifecycle.md](decisions/0005-macos-daemon-login-lifecycle.md)
- [0006-online-sgd-personalization.md](decisions/0006-online-sgd-personalization.md)

## Validation

- [validation/index.md](validation/index.md)
- [compatibilityMatrix.md](validation/compatibilityMatrix.md)
- [performance.md](validation/performance.md)
- [knownIssues.md](validation/knownIssues.md)

## Source directories (planned / emerging)

| Path | Role |
| --- | --- |
| `cmd/tsimed` | Daemon entrypoint |
| `cmd/tsime-userdict` | User dictionary management CLI |
| `cmd/tsime-personal` | Personal model path/status/clear CLI |
| `cmd/tsime-personal-eval` | Deterministic personalization replay metrics |
| `internal/personal` | Defaults-on personalization settings (no env vars) |
| `internal/ipc` | IPC server/client, framing |
| `internal/session` | Session manager |
| `internal/engine` | Input engine core |
| `internal/dictionary` | Compiled dictionary index |
| `internal/pinyin` | Continuous pinyin decoder |
| `internal/ranking` | FastRanker + Ranking Coordinator |
| `internal/inference` | Inference Backend interface / stubs |
| `data/dicts` | Built `.tsdict`, emoji map, license records |
| `third_party/rime-ice` | Cached upstream YAML + OpenCC emoji sources |
| `platforms/macos` | macOS IMKit adapter |
| `platforms/windows` | Windows TSF adapter |
| `platforms/linux` | Linux Fcitx5 adapter |
| `tools/dictBuilder` | Dictionary build tooling |
| `tests/integration` | Cross-module integration tests |
| `tests/compatibility` | Platform compatibility suites |
| `tests/performance` | Latency / throughput benches |

## Build and test entry

- Build: see [development/build.md](development/build.md) (Go toolchain once `cmd/tsimed` exists).
- Test: see [development/test.md](development/test.md).
- Agent protocol: [../Agent.md](../Agent.md).
