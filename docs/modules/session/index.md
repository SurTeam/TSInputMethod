# Module: session

**Status:** implemented / compiled / auto-tested.

## Goals

- Own per-client IME session state and isolation.
- Route key events to the engine and return composition / candidates / commit.
- Support Chinese editing keys: letter compose, backspace, Esc cancel, space commit top candidate, digit select.

## Boundaries

- **In scope:** session table, preedit buffer, generation, mode switch, edit-key behavior.
- **Out of scope:** protobuf framing (ipc), dictionary algorithms (engine/pinyin), OS adapter callbacks.

## Source locations

| Path | Role |
| --- | --- |
| `cmd/imed` | Wires session manager into the daemon |
| `internal/session` | Session manager and session objects |
| `internal/engine` | Called by session for conversion |

## Related

- [lifecycle.md](../../architecture/lifecycle.md)
- [dataFlow.md](../../architecture/dataFlow.md)
- [../pinyin/index.md](../pinyin/index.md)
