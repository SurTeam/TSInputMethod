# Module: pinyin

**Status:** implemented / auto-tested (quanpin + Xiaohe/flypy shuangpin).

## Goals

- Continuous full-pinyin syllable segmentation with ambiguity (e.g. `xian` vs `xi'+'an`).
- Lattice / composition search over dictionary entries.
- Frequency-oriented candidate ordering for the engine API.
- Xiaohe (flypy) double-pinyin expands by fixed 2-key pairs (rime-ice / flypy algebra), then decodes on those syllables without re-segmenting the joined quanpin string.
- Optional UserDictionary overlay merges into exact/prefix/partial/lattice lookups via `Decoder.SetUserDictionary` before the finish prune.

## Boundaries

- **In scope:** syllable validity, segmentation, decode candidates, scheme normalization, user-overlay merge points.
- **Out of scope:** IPC, session editing keys (owned by `internal/session`), neural ranking, user-dictionary persistence.

## Schemes

| Scheme | Env `TSIME_SCHEME` | Behavior |
| --- | --- | --- |
| xiaohe / flypy (default) | `xiaohe`, `flypy`, `double`, `shuangpin`, or unset | Fixed 2-key pairs via flypy map (`hc`→`hao`, `uijp`→`shi'+'jie`); decode keeps that split |
| quanpin | `quanpin` | Preedit is full pinyin with continuous segmentation |

Flypy mapping follows rime-ice `double_pinyin_flypy.schema.yaml` algebra / preedit rules (and the common Xiaohe table: `c`=ao, `h`=ang, `v/i/u`=zh/ch/sh). Joining syllables into `shijie` and re-running quanpin segmentation is wrong for flypy because `jie` can also be read as `ji'+'e`.

Examples: `nihc` → `[ni, hao]` → first page `你好` then single-char `你/尼/泥…` (weak same-length phrases like `拟好` stay off the first page); `nih` → incomplete trailing `h` surfaces short matches like `你好` ahead of long phrase spam; single-key `h` prefers short `h*` syllables (e.g. `哈`) over encyclopedia-length phrases; `wo`/`yo` stay `wo`/`yo` (not `wuo`/`yuo`); ambiguous flypy keys expand every valid reading and interleave them into candidates (`lo`→`luo`/`lo` → 落…/咯, `lt`→`lve`/`lue` → 略, `dv`→`dui`/`du` → 对/度); `uijp` → `[shi, jie]` → 世界 with raw comment `uijp` (covered keys, not quanpin `shi'jie`); `vsgorfmb` → `[zhong, guo, ren, min]` → 中国人民 plus progressive prefix words 中国人/中国/中.

Flypy decode lists dictionary words by decreasing complete-syllable coverage (N…1), including odd trailing keys. Incomplete/prefix hits prefer shorter syllable coverage and shorter text so single/half keys feel like common shuangpin IMEs. Lattice concatenations such as 尼好/你好是接 are not used on the flypy path. Session paging keeps a fixed `page_size=9`. Engine comments remain the covered raw preedit keys for adapters that need them; macOS candidate panel shows bare Chinese while composition carries soft-confirmed Chinese + remaining keys.

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
