# Module: inference

**Status:** implemented interface + noop/unavailable stubs / auto-tested. No real model weights.

## Goals

- Define a Backend interface for future ONNX / Core ML / other local runtimes.
- Keep Go responsible for scheduling and fusion, not kernel math.
- Guarantee unavailable backends fail explicitly so ranking can degrade.

## Boundaries

- **In scope:** Backend contract, NoopBackend, UnavailableBackend, FuncBackend helper.
- **Out of scope:** shipping neural graphs, GPU/NPU drivers, remote model servers.

## Source locations

| Path | Role |
| --- | --- |
| `internal/inference/backend.go` | Backend API and stubs |
| `internal/inference/backend_test.go` | Noop / unavailable coverage |

## Integration

Ranking Coordinator accepts any `inference.Backend`. Engine exposes `PinyinEngine.SetInferenceBackend` for injection without changing session or platform adapters.

## Related

- [../ranking/index.md](../ranking/index.md)
- [../engine/index.md](../engine/index.md)
