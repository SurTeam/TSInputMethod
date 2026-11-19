# Module: session

**Status:** implemented / compiled / auto-tested.

## Goals

- Own per-client IME session state and isolation.
- Route key events to the engine and return composition / candidates / commit.
- Support Chinese editing keys: letter compose, backspace, Esc cancel, space commit the highlighted candidate, Enter/Return commit raw preedit and clear candidates, digit select, `-`/`=` and Up/Down candidate paging with a fixed `page_size=9`, Left/Right highlight move within the current page.
- Candidate list display labels are numbered per page as `1.`…`9.` while commit/select still use the underlying candidate text. `CandidateList.selected_index` carries the highlighted page-local index for adapters.
- With no preedit, digits commit as themselves and punctuation follows rime-ice `half_shape` (mixed Chinese/ASCII, paired quotes). During preedit, punctuation is consumed and ignored. Shift+letter opens a literal uppercase preedit with itself as the only candidate; Shift+symbol uses the same half-shape map. Control/Option/Command chords and English mode always return unconsumed passthrough.
- Selecting a candidate that only covers a prefix of the preedit soft-confirms that text into the composition display (`已选中文` + remaining raw keys) without hard-committing yet; candidates stay bare Chinese words. Backspace first undoes the latest soft confirmation (restoring its keys with the composition cursor at the end of the full preedit), then deletes raw preedit characters from that end. Completing the remaining keys hard-commits the whole confirmed phrase.
- Classify printable `characters` first after modifier passthrough. Edit `key_code` values follow macOS CG codes from Karabiner-Elements: Delete=51, Escape=53, Space=49, ForwardDelete=117, Return=36, keypad Enter=76, Left=123, Right=124, Down=125, Up=126 ([ADR 0002](../../decisions/0002-semantic-edit-keycodes.md)). Wire modifiers use AppKit bit layout (`shift=1<<17`, `control=1<<18`, `option=1<<19`, `command=1<<20`). Legacy ASCII 8/27/32 are not edit keys.

## Boundaries

- **In scope:** session table, preedit buffer, generation, mode switch (commits raw preedit before switching), edit-key behavior.
- **Out of scope:** protobuf framing (ipc), dictionary algorithms (engine/pinyin), OS adapter callbacks.

## Source locations

| Path | Role |
| --- | --- |
| `cmd/tsimed` | Wires session manager into the daemon |
| `internal/session` | Session manager and session objects |
| `internal/engine` | Called by session for conversion |

## Related

- [lifecycle.md](../../architecture/lifecycle.md)
- [dataFlow.md](../../architecture/dataFlow.md)
- [../pinyin/index.md](../pinyin/index.md)
