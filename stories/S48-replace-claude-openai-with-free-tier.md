---
id: S48
title: Replace Claude and ChatGPT backends with Gemini 2.0 Flash as the default model
status: done
created: 2026-03-18
updated: 2026-03-18
---

# S48 — Replace Claude and ChatGPT backends with Gemini 2.0 Flash as the default model

## User story

As a **runner using CoachLM**,
I want to **chat with the AI coach without needing to pay for or configure Claude or OpenAI API keys**
so that **the app works out of the box for everyone, powered by Gemini 2.0 Flash**.

## Acceptance criteria

- [ ] The Claude and OpenAI LLM backends are removed from the app entirely (code, settings UI, storage schema, validation)
- [ ] The Gemini backend (`internal/llm/free.go`) is renamed internally from `"free"` to `"gemini"` — the `Name()` method returns `"gemini"`, the storage value is `"gemini"`, and the ldflags variable is renamed to `builtinGeminiApiKey`
- [ ] Everywhere the model is shown to the user — settings page, onboarding, any tooltip or label — it is displayed as **"Gemini 2.0 Flash"**, not "Free", "free", or "Gemini"
- [ ] The local/Ollama backend is retained as an optional power-user override, labelled **"Local Model (Ollama)"** in the UI
- [ ] Settings UI no longer shows: backend selector dropdown, Claude API key field, Claude model field, OpenAI API key field, OpenAI model field
- [ ] Settings UI retains: Ollama endpoint + model fields under a collapsible "Advanced: Local Model (Ollama)" section
- [ ] The Gemini API key is injected at build time via `-ldflags "-X coachlm/internal/llm.builtinGeminiApiKey=..."` — users never enter a key
- [ ] If no built-in key is present (local dev builds), the app reads `GEMINI_API_KEY` from the environment as a fallback
- [ ] `active_llm` in the SQLite settings table uses `"gemini"` (not `"free"`) for new installs; existing rows with `"claude"`, `"openai"`, or `"free"` are migrated to `"gemini"` on first startup
- [ ] `validLLMs` in `internal/storage/settings.go` is updated to only allow `"gemini"` and `"local"`
- [ ] `SettingsData` struct in `app.go` drops `ClaudeAPIKey`, `OpenAIAPIKey`, `ClaudeModel`, `OpenAIModel`, `ActiveLLM` fields
- [ ] `SaveSettingsData` and `GetSettingsData` in `app.go` no longer handle Claude/OpenAI fields
- [ ] Onboarding wizard step 2 ("Choose Your AI Backend") is removed; the wizard goes directly from Welcome → Connect Strava → Athlete Profile → All Set
- [ ] All existing `go test ./...` tests pass
- [ ] E2e Playwright specs updated: remove backend-selector tests, add tests confirming "Gemini 2.0 Flash" label is present and no Claude/OpenAI fields exist
- [ ] The wails mock (`frontend/e2e/mocks/wails.ts`) is updated to drop Claude/OpenAI fields from `DEFAULT_SETTINGS` and use `activeLlm: 'gemini'`

## Technical notes

### Files to change

| File | Change |
|---|---|
| `internal/llm/free.go` | Rename file to `gemini.go`; rename struct `Free` → `Gemini`, config `FreeConfig` → `GeminiConfig`, constructor `NewFree` → `NewGemini`; rename `builtinFreeApiKey` → `builtinGeminiApiKey`; `Name()` returns `"gemini"` |
| `internal/llm/free_test.go` | Rename to `gemini_test.go`; update all references |
| `internal/storage/settings.go` | Replace `"free"` and `"claude"` and `"openai"` in `validLLMs` with `"gemini"` and `"local"`; remove `ClaudeAPIKey`, `OpenAIAPIKey`, `ClaudeModel`, `OpenAIModel` from `Settings` struct and SQL columns |
| `internal/llm/claude.go` + `claude_test.go` | Delete both files |
| `internal/llm/openai.go` + `openai_test.go` | Delete both files |
| `main.go` (`createLLMClient`) | Replace `"free"` case with `"gemini"`; remove `"claude"` and `"openai"` cases; default to `gemini`; keep `"local"` case |
| `app.go` (`SettingsData`) | Remove `ClaudeAPIKey`, `OpenAIAPIKey`, `ClaudeModel`, `OpenAIModel`, `ActiveLLM` |
| `app.go` (`GetSettingsData`, `SaveSettingsData`) | Stop reading/writing removed fields |
| `frontend/src/Settings.svelte` | Remove backend selector, Claude fields, OpenAI fields; display a static label "Powered by Gemini 2.0 Flash"; keep Ollama section under collapsible "Advanced: Local Model (Ollama)" |
| `frontend/src/Onboarding.svelte` | Remove step 2 (backend choice); renumber remaining steps; update any references to model name to say "Gemini 2.0 Flash" |
| `frontend/e2e/mocks/wails.ts` | Drop Claude/OpenAI fields; set `activeLlm: 'gemini'` in `DEFAULT_SETTINGS` |
| `frontend/e2e/settings.spec.ts` | Update to match new settings layout; assert "Gemini 2.0 Flash" label is visible |
| `frontend/e2e/s46-free-ai.spec.ts` | Delete; replace with `s48-gemini-default.spec.ts` |
| `frontend/e2e/onboarding.spec.ts` | Update step count to 4 and step 2 title to "Connect Strava" |

### Migration strategy

In `storage.DB.initSchema()`, run after table creation:

```sql
UPDATE settings SET active_llm = 'gemini' WHERE active_llm IN ('claude', 'openai', 'free');
```

This is idempotent — safe to run on every startup.

### Build-time key injection

Add a `GEMINI_API_KEY` GitHub Actions secret and update the release workflow:

```yaml
- name: Build
  run: wails build ${{ matrix.build-flags }} -ldflags "-X coachlm/internal/llm.builtinGeminiApiKey=${{ secrets.GEMINI_API_KEY }}"
```

This is out of scope for the story itself (requires a real key in CI secrets) but the code must be structured to support it.

### Model display name

The constant `defaultFreeModel = "gemini-2.0-flash"` in `gemini.go` drives the actual API call. The UI always shows the human-readable label **"Gemini 2.0 Flash"** — never the raw identifier.

### Keep `internal/llm/local.go`

Ollama/local support stays unchanged. Advanced users who want full privacy or a specific model can point the app at a local Ollama instance.

## Tests required

- Unit: delete `claude_test.go` and `openai_test.go`; rename `free_test.go` → `gemini_test.go` and update all references
- Unit: `TestSettings_Validate` — update to only allow `"gemini"` and `"local"`
- Unit: migration test — verify rows with `active_llm` of `'claude'`, `'openai'`, or `'free'` are all updated to `'gemini'` on `initSchema`
- E2e: `settings.spec.ts` — no backend selector; "Gemini 2.0 Flash" label is visible; Ollama section present under "Advanced"
- E2e: `onboarding.spec.ts` — 4 steps; step 2 is "Connect Strava"
- E2e: `s48-gemini-default.spec.ts` — no Claude/OpenAI fields anywhere; "Gemini 2.0 Flash" label present in settings

## Out of scope

- Letting users enter their own Gemini API key in the UI
- Changing the Gemini model used (`gemini-2.0-flash` stays)
- Cloud sync or multi-device features

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-18 | draft | Created |
| 2026-03-18 | draft | Updated: backend renamed from "free" to "gemini"; UI label is "Gemini 2.0 Flash" throughout |
| 2026-03-18 | in-progress | Implementation started |
| 2026-03-18 | done | All backend + frontend changes complete, go test + 69/69 e2e pass |
