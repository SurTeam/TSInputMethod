# Module: session

**Status:** implemented / compiled / auto-tested.

## Goals

- Own per-client IME session state and isolation.
- Route key events to the engine and return composition / candidates / commit.
- Support Chinese editing keys: letter compose, backspace, Esc cancel, space commit top candidate, Enter/Return commit raw preedit and clear candidates, digit select, `-`/`=` candidate paging with a fixed `page_size=9` regardless of input length or total candidate count.
- Candidate list display labels are numbered per page as `1.`…`9.` while commit/select still use the underlying candidate text.
- Selecting a candidate that only covers a prefix of the preedit commits that text and leaves the remaining code in composition for continued phrase input (including after odd-length flypy keys).
- Classify printable `characters` first. Edit `key_code` values follow macOS CG codes from Karabiner-Elements: Delete=51, Escape=53, Space=49, ForwardDelete=117, Return=36, keypad Enter=76 ([ADR 0002](../../decisions/0002-semantic-edit-keycodes.md)). Legacy ASCII 8/27/32 are not edit keys.

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
