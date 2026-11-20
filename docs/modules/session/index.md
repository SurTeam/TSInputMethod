# Module: session

**Status:** implemented / compiled / auto-tested.

## Goals

- Own per-client IME session state and isolation.
- Route key events to the engine and return composition / candidates / commit.
- Support Chinese editing keys: letter compose, backspace, Esc cancel, space commit the highlighted candidate, Enter/Return commit raw preedit and clear candidates, digit select, `-`/`=` and Up/Down candidate paging with a fixed `page_size=9`, Left/Right highlight move within the current page.
- Candidate list display labels are numbered per page as `1.`…`9.` while commit/select still use the underlying candidate text. `CandidateList.selected_index` carries the highlighted page-local index for adapters.
- With no preedit, digits commit as themselves and punctuation follows rime-ice `half_shape` (mixed Chinese/ASCII, paired quotes). During preedit, punctuation is consumed and ignored. Shift+letter opens a literal uppercase preedit with itself as the only candidate; Shift+symbol uses the same half-shape map. Control/Option/Command chords and English mode always return unconsumed passthrough.
- Selecting a candidate that only covers a prefix of the preedit soft-confirms that text into the composition display (`已选中文` + remaining raw keys) without hard-committing yet; candidates stay bare Chinese words. Backspace first undoes the latest soft confirmation (restoring its keys with the composition cursor at the end of the full preedit), then deletes raw preedit characters from that end. Completing the remaining keys hard-commits the whole confirmed phrase.
- Maintain learning context as `CommittedContext` (hard-committed text) plus `ConfirmedComposition` (current soft-confirmed segments). `EffectiveContext` is the Unicode-code-point tail of their concatenation with default window 48 and hard max 64. Soft-confirm immediately updates EffectiveContext and feeds it into remaining-code decode/ranking via `engine.ContextDecoder`; hard commit appends the actual committed text once and clears confirmed segments so the same phrase is not double-counted. Escape/backspace roll composition context back; `ClearLearningContext` / `InvalidateLearningContext`, `Manager.Close`, and `Manager.ClearAllLearningContext` drop committed learning context. Sessions do not share context. Semantic boundary resets also clear committed learning context: Chinese↔English `SetMode`, and adapter-observed IME deactivate / app focus loss via IPC `InvalidateLearningContext` (macOS `deactivateServer`, Windows `OnSetFocus(FALSE)`). Platform adapters should also call invalidate when they observe external cursor/edits; the daemon cannot assume app buffer sync. Accurate tracking of post-commit in-app Backspace or other external edits is a follow-up once reliable platform edit signals exist—do not assume `CommittedContext` stays synchronized with the target text field.
- Displayed candidates are stored as an immutable `candidateSnapshot` (preedit, EffectiveContext, generation, full candidate metadata, ranking/userdict version placeholders). Digit/space/mouse select and commit reuse that snapshot so indexes cannot drift from a later decode. Non-top selections with compatible ConsumedKeys/Incomplete create `PendingFeedback` versus the displayed top; top-only or single-candidate skips pairwise. Soft-confirm backspace drops the last pending event; Escape clears all pending. Final hard commit accepts pending once into `FeedbackEvent`s exposed via `TakeAcceptedFeedback`. Stale `SelectCandidate` generations are rejected.
- Soft-confirmed segments retain candidate Syllables. On final hard commit of confirmed text, Session notifies an optional `PhraseObserver` with the concatenated phrase and recoverable syllables so UserDictionary auto-learn can promote repeated reliable phrases. Raw preedit-only commits without syllables are not observed.
- Classify printable `characters` first after modifier passthrough. Edit `key_code` values follow macOS CG codes from Karabiner-Elements: Delete=51, Escape=53, Space=49, ForwardDelete=117, Return=36, keypad Enter=76, Left=123, Right=124, Down=125, Up=126 ([ADR 0002](../../decisions/0002-semantic-edit-keycodes.md)). Wire modifiers use AppKit bit layout (`shift=1<<17`, `control=1<<18`, `option=1<<19`, `command=1<<20`). Legacy ASCII 8/27/32 are not edit keys.

## Boundaries

- **In scope:** session table, preedit buffer, generation, mode switch (commits raw preedit before switching), edit-key behavior, EffectiveContext lifecycle for ranking.
- **Out of scope:** protobuf framing (ipc), dictionary algorithms (engine/pinyin), OS adapter callbacks, SGD personalization weights.

## Source locations

| Path | Role |
| --- | --- |
| `cmd/tsimed` | Wires session manager into the daemon |
| `internal/session` | Session manager and session objects |
| `internal/session/context.go` | Unicode context window helpers |
| `internal/session/feedback.go` | Immutable candidate snapshots, PendingFeedback, accepted FeedbackEvent |
| `internal/session/session.go` | Soft-confirm syllable capture + PhraseObserver notify on hard commit |
| `internal/engine` | Called by session for conversion; `DecodeWithContext` receives EffectiveContext |

## Related

- [lifecycle.md](../../architecture/lifecycle.md)
- [dataFlow.md](../../architecture/dataFlow.md)
- [../pinyin/index.md](../pinyin/index.md)
