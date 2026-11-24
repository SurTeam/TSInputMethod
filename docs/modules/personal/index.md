# Module: personal

**Status:** implemented / compiled / auto-tested (settings defaults).

## Goals

- Hold defaults-on personalization toggles without environment variables.
- Expose PrivacyMode and weight knobs consumed by `PinyinEngine.ApplySettings`.

## Boundaries

- **In scope:** `Settings` snapshot and setters for personalization, user dictionary, auto-learn, PrivacyMode, SGD/UserFrequency weights, learn min-repeats.
- **Out of scope:** SGD math (ranking), user-dict storage (dictionary), session feedback lifecycle.

## Source locations

| Path | Role |
| --- | --- |
| `internal/personal/settings.go` | Defaults-on settings |
| `internal/personal/settings_test.go` | Default enable / disable checks |

## Operator notes

- Defaults: personalization, user dictionary, and auto-learn **on**; PrivacyMode **off**.
- PrivacyMode disables fused personalization scoring, feedback/usage recording, and phrase auto-learn when applied via the engine; it also discards pending auto-learn counts.
- `AutoLearnEnabled=false` (or `UserDictionaryEnabled=false`) stops `PinyinEngine.ObserveCommittedPhrase` from promoting phrases; `tsimed` routes the session `PhraseObserver` through the engine so these toggles apply.
- Tuning: `SGDWeight` (default 40), `UserFrequencyWeight` (default 8), `LearnMinRepeats` (default 2, applied to the engine learner by `ApplySettings`) via `Settings` setters—not env vars.
- Clear weights: `go run ./cmd/tsime-personal clear-weights`
- Clear user dict: `go run ./cmd/tsime-userdict clear learned|all`
- Clear short context: `Manager.ClearAllLearningContext` / `Session.ClearLearningContext`
- Storage: `UserConfigDir/TSInputMethod/personal_model.json` and `user_dictionary.json` (`0700`/`0600`)
- Eval: `go run ./cmd/tsime-personal-eval` (deterministic fixtures only)

## Related

- [../ranking/index.md](../ranking/index.md)
- [../dictionary/index.md](../dictionary/index.md)
- [../../decisions/0006-online-sgd-personalization.md](../../decisions/0006-online-sgd-personalization.md)
- [../../principles/security.md](../../principles/security.md)
