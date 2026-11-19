# ADR 0002: Wire key_code uses macOS CG virtual keycodes

- **Status:** Accepted (supersedes earlier ASCII 8/27/32 edit-key contract)
- **Date:** 2026-03-21

## Context

An earlier contract treated `KeyEvent.key_code` as semantic ASCII controls (Backspace=8, Escape=27, Space=32). Adapters remapped host keys into those values. That collided with real macOS CG/ANSI codes documented by Karabiner-Elements, where letter `c` is already `0x08`, hyphen is `0x1b`, and letter `u` is `0x20`.

Karabiner-Elements `pqrs::osx::cg_event::key_code` is the authoritative macOS virtual-key table used here:

- `keyboard_delete_or_backspace = 0x33` (51)
- `keyboard_escape = 0x35` (53)
- `keyboard_spacebar = 0x31` (49)
- `keyboard_return_or_enter = 0x24` (36)
- `keyboard_keypad_enter = 0x4c` (76)
- `keyboard_delete_forward = 0x75` (117)
- letters/punctuation keep their CG codes (`c=0x08`, `hyphen=0x1b`, `u=0x20`, …)

## Decision

1. `KeyEvent.key_code` on the wire is the macOS CG virtual keycode space.
2. macOS adapter forwards `NSEvent.keyCode` unchanged.
3. Linux/Windows adapters translate host edit keys into the same CG codes (BackSpace/VK_BACK→51, Escape→53, Space→49, Return/Enter→36, keypad Enter→76, Delete→117).
4. Printable input is still classified primarily from `characters` (compose, digit select, `-`/`=` paging, space commit, `\r`/`\n` raw-preedit commit).
5. Legacy ASCII edit codes 8/27/32/127 are not edit keys on the wire.

## Consequences

- Letter `c` / hyphen / letter `u` no longer collide with fabricated ASCII edit codes.
- Session edit-key constants match Karabiner CG values directly.
- Non-macOS adapters own a small host→CG translation for edit keys only.

## Alternatives considered

| Alternative | Why not |
| --- | --- |
| Keep ASCII 8/27/32 semantic codes | Invented second numbering; collisions with macOS CG |
| HID usage IDs on the wire | Larger change; CG table already unique for needed keys |
| Character-only events | Loses empty-character Delete/Escape |

## References

- Karabiner-Elements `vendor/vendor/include/pqrs/osx/cg_event/key_code.hpp`
- [../principles/crossPlatform.md](../principles/crossPlatform.md)
- [../platforms/macos/index.md](../platforms/macos/index.md)
- [../modules/session/index.md](../modules/session/index.md)
