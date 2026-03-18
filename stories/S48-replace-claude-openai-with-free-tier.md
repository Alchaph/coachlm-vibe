---
id: S48
title: Replace Claude and ChatGPT backends with Gemini free tier
status: draft
created: 2026-03-18
updated: 2026-03-18
---

# S48 — Replace Claude and ChatGPT backends with Gemini free tier

## User story

As a **runner using CoachLM**,
I want to **chat with the AI coach without needing to pay for or configure Claude or OpenAI API keys**
so that **the app works out of the box for everyone using only the free Gemini tier**.

## Acceptance criteria

- [ ] The Claude and OpenAI LLM backends are removed from the app entirely (code, settings UI, storage schema, validation)
- [ ] The "free" Gemini backend (`internal/llm/free.go`) becomes the **only** backend — no backend selector shown in settings
- [ ] The local/Ollama backend is retained as an optional power-user override (for users who want to run their own model)
- [ ] Settings UI no longer shows: backend selector dropdown, Claude API key field, Claude model field, OpenAI API key field, OpenAI model field
- [ ] Settings UI retains: Ollama endpoint + model fields (collapsed/optional section labelled "Advanced: Local Model")
- [ ] The Gemini API key is injected at build time via `-ldflags "-X coachlm/internal/llm.builtinFreeApiKey=..."` — users never enter a key
- [ ] If no built-in key is present (local dev builds), the app reads `GEMINI_API_KEY` from the environment as a fallback — existing behaviour
- [ ] `active_llm` in the SQLite settings table defaults to `"free"` for new installs; existing rows with `claude` or `openai` are migrated to `"free"` on first startup
- [ ] `validLLMs` in `internal/storage/settings.go` is updated to only allow `"free"` and `"local"`
- [ ] `SettingsData` struct in `app.go` drops `ClaudeAPIKey`, `OpenAIAPIKey`, `ClaudeModel`, `OpenAIModel`, `ActiveLLM` fields
- [ ] `SaveSettingsData` and `GetSettingsData` in `app.go` no longer handle Claude/OpenAI fields
- [ ] Onboarding wizard step 2 ("Choose Your AI Backend") is removed; the wizard goes directly from Welcome → Connect Strava → Athlete Profile → All Set
- [ ] All existing `go test ./...` tests pass
- [ ] E2e Playwright specs updated: remove `s46-free-ai.spec.ts` backend-selector tests, add a test confirming no backend selector is present, update `settings.spec.ts` to reflect the new layout
- [ ] The wails mock (`frontend/e2e/mocks/wails.ts`) is updated to drop Claude/OpenAI fields from `DEFAULT_SETTINGS`

## Technical notes

### Files to change

| File | Change |
|---|---|
| `internal/storage/settings.go` | Remove `claude`/`openai` from `validLLMs`; remove `ClaudeAPIKey`, `OpenAIAPIKey`, `ClaudeModel`, `OpenAIModel` from the `Settings` struct and SQL; add a migration that sets `active_llm = 'free'` where it is currently `'claude'` or `'openai'` |
| `internal/llm/claude.go` + `claude_test.go` | Delete both files |
| `internal/llm/openai.go` + `openai_test.go` | Delete both files |
| `main.go` (`createLLMClient`) | Remove `claude` and `openai` switch cases; default to `free`; keep `local` case |
| `app.go` (`SettingsData`) | Remove `ClaudeAPIKey`, `OpenAIAPIKey`, `ClaudeModel`, `OpenAIModel`, `ActiveLLM` |
| `app.go` (`GetSettingsData`, `SaveSettingsData`) | Stop reading/writing removed fields |
| `frontend/src/Settings.svelte` | Remove backend selector, Claude fields, OpenAI fields; keep Ollama section under "Advanced: Local Model" collapsible |
| `frontend/src/Onboarding.svelte` | Remove step 2 (backend choice); renumber remaining steps |
| `frontend/e2e/mocks/wails.ts` | Drop Claude/OpenAI fields from `DEFAULT_SETTINGS` |
| `frontend/e2e/settings.spec.ts` | Update to match new settings layout |
| `frontend/e2e/s46-free-ai.spec.ts` | Remove backend-selector tests; replace with a test that confirms the selector is absent |
| `frontend/e2e/onboarding.spec.ts` | Update step count and step titles |

### Migration strategy

In `storage.DB.initSchema()` (or a new `storage.DB.migrate()` call), run:

```sql
UPDATE settings SET active_llm = 'free' WHERE active_llm IN ('claude', 'openai');
```

This runs on every startup but is idempotent.

### Build-time key injection

The release workflow already builds with Wails. Add a `GEMINI_API_KEY` GitHub Actions secret and pass it at build time:

```yaml
- name: Build
  run: wails build ${{ matrix.build-flags }} -ldflags "-X coachlm/internal/llm.builtinFreeApiKey=${{ secrets.GEMINI_API_KEY }}"
```

This is out of scope for the story itself (requires a real key in CI secrets) but the code must support it.

### Keep `internal/llm/local.go`

Ollama/local support stays. Advanced users who want full privacy or a specific model can point the app at a local Ollama instance via the Settings page.

## Tests required

- Unit: delete `claude_test.go` and `openai_test.go` (backends removed); `free_test.go` and `local_test.go` remain and must pass
- Unit: `TestSettings_Validate` in `internal/storage/settings_test.go` — update to only allow `"free"` and `"local"`
- Unit: migration test — verify rows with `active_llm = 'claude'` are updated to `'free'` on `initSchema`
- E2e: `settings.spec.ts` — no backend selector present; Ollama section visible under "Advanced"
- E2e: `onboarding.spec.ts` — wizard has 4 steps (not 5); step 2 is "Connect Strava"
- E2e: new `s48-free-tier-only.spec.ts` — confirm no Claude/OpenAI fields anywhere in the UI

## Out of scope

- Removing the `builtinFreeApiKey` ldflags mechanism (keep it — used in production builds)
- Adding a UI for users to enter their own Gemini key (not needed; the point is zero-setup)
- Changing the Gemini model used (`gemini-2.0-flash` stays as the default)
- Cloud sync or multi-device features

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-18 | draft | Created |
