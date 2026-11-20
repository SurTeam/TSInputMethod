# ADR 0006: Online SGD personalization with local user dictionary

- **Status:** Accepted
- **Date:** 2026-11-20

## Context

TSInputMethod already decodes double-pinyin and ranks with FastRanker frequency order. Users need context-sensitive correction learning and a persistent user-dictionary overlay without shipping an ML runtime, network services, or scheme changes. Settings must default on without environment variables. Personal data must stay local with clear/reset and an explicit privacy mode.

## Decision

1. **EffectiveContext** is `Tail(CommittedContext + ConfirmedComposition, window 48 / max 64 Unicode)`. Soft-confirm updates ranking immediately; backspace/escape roll back; final commit appends committed text once.
2. **Immutable candidate snapshots** bind display/select/commit. Non-top compatible selections create `PendingFeedback`; SGD accepts only after final commit; undo/cancel/stale generation invalidate; pairwise samples stay inside compatible Coverage/Incomplete/Composed groups.
3. **UserDictionary** is a syllable-based overlay (exact/prefix/double-pinyin) merged before prune, with usage counts, auto-learn from reliable repeated phrase commits, atomic JSON at `UserConfigDir/TSInputMethod/user_dictionary.json` (0600), and CLI `cmd/tsime-userdict`.
4. **PersonalModel** is pure-Go Feature-Hashing (FNV-1a, dim `2^18`) sparse pairwise logistic SGD with L2/grad-clip/weight-cap, short-term preference cache (capacity 256, TTL 5m, LRU eviction), and bounded async `UpdateQueue` (non-blocking drop). Weights persist separately at `UserConfigDir/TSInputMethod/personal_model.json` without raw context history. Capacity: sparse map trimmed to `MaxWeights` (default 50000) by smallest absolute weight.
5. **FusedRanker** score: `log1p(base) + UserFrequencyWeight*log1p(usage) + SGDWeight*PersonalScore`, only within compatible groups. Disabling personalization or enabling PrivacyMode restores FastRanker-only order and skips enqueue.
6. **Config** lives in `internal/personal.Settings` defaults-on (personalization, user dictionary, auto-learn); no env vars for these toggles.
7. **Privacy:** explicit PrivacyMode; clear/reset for weights, user dict, and short learning context; local-only files with user perms; no raw context training logs. Secure-input auto-detection is **not** claimed—adapters do not currently expose a reliable secure-field signal.
8. **Eval:** `cmd/tsime-personal-eval` replays deterministic fixtures predict→observe→train and reports Top-1/5, MRR, rank deltas, false promotions, memory, latency across Fast / UserFreq / SGD combinations.

## Consequences

- Personalization can improve same-context corrections without rewriting scheme or Coverage rules.
- Cross-context false promotions remain possible; fixtures and docs must treat them as measured risk, not eliminated.
- Operators clear data via engine APIs / `cmd/tsime-personal` / `cmd/tsime-userdict`; PrivacyMode is the honest sensitive-field control until platform secure-input detection exists.

## Alternatives considered

| Alternative | Why not |
| --- | --- |
| Neural / ONNX reranker | Violates pure-Go, offline, low-latency constraints |
| Mutating built-in `.tsdict` | Breaks read-only system dict and packaging |
| Env-var feature flags | Conflicts with defaults-on settings requirement |
| Persisting raw context logs | Privacy violation |

## References

- [../modules/ranking/index.md](../modules/ranking/index.md)
- [../modules/dictionary/index.md](../modules/dictionary/index.md)
- [../modules/session/index.md](../modules/session/index.md)
- [../principles/security.md](../principles/security.md)
