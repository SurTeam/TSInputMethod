# ADR 0003: Custom macOS candidate panel

## Status

Accepted

## Context

Horizontal `IMKCandidates` (`kIMKSingleRowSteppingCandidatePanel`) could show candidates and accept selection, but left/right highlight stayed stuck on the first item even when session `selected_index` and commit target were correct. Attribute and label-prefix workarounds still depended on IMK selection APIs and remained unreliable under `SendServerKeyEventFirst`.

## Decision

Own the macOS candidate UI in the adapter:

- Draw a borderless floating `NSPanel` with a custom horizontal row of up to 9 equal-width columns.
- Render darker numeric labels `1`–`9` beside bare Chinese candidate text; highlight follows session `selected_index`.
- Keep column count and locked equal column widths stable across paging/expansion; mouse wheel pages via session Up/Down.
- Keep composition/preedit in the client through `setMarkedText` attributes owned by `TSInputController`, not through `IMKCandidates`.

Session still owns paging, `selected_index`, digit select, and soft-confirm semantics over IPC.

## Consequences

- macOS adapter no longer uses `IMKCandidates` for display or highlight.
- Caret-relative panel placement may need host-client coordinate fallbacks in some apps.
- Docs and KI-005 now describe the custom panel path; real-env TextEdit typing remains pending after reinstall.
