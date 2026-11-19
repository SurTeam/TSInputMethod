# Module: pinyin

**Status:** implemented / auto-tested (quanpin + Xiaohe/flypy shuangpin).

## Goals

- Continuous full-pinyin syllable segmentation with ambiguity (e.g. `xian` vs `xi'+'an`).
- Lattice / composition search over dictionary entries.
- Frequency-oriented candidate ordering for the engine API.
- Xiaohe (flypy) double-pinyin input expanded to quanpin before decode.

## Boundaries

- **In scope:** syllable validity, segmentation, decode candidates, scheme normalization.
- **Out of scope:** IPC, session editing keys (owned by `internal/session`), neural ranking.

## Schemes

| Scheme | Env `TSIME_SCHEME` | Behavior |
| --- | --- | --- |
| quanpin (default) | `quanpin` / unset | Preedit is full pinyin |
| xiaohe / flypy | `xiaohe`, `flypy`, `double`, `shuangpin` | Pairs expand via flypy map then reuse Decoder |

Example: Xiaohe `nihc` → `nihao` → dictionary candidates (你好).

## Source locations

| Path | Role |
| --- | --- |
| `internal/pinyin/syllable.go` | Valid syllable set |
| `internal/pinyin/segment.go` | Continuous segmentation |
| `internal/pinyin/decoder.go` | Exact / lattice / prefix decode |
| `internal/pinyin/xiaohe.go` | Flypy/Xiaohe pair map (`FlypyToFull`) |
| `internal/pinyin/scheme.go` | Scheme parse / env / normalize |
| `internal/engine` | `PinyinEngine` applies scheme before decode |

## Related

- [../dictionary/index.md](../dictionary/index.md)
- [../session/index.md](../session/index.md)
- [../engine/index.md](../engine/index.md)
