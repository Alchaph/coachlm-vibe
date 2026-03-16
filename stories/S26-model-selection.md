---
id: S26
title: Model selection per LLM backend
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S26 — Model selection per LLM backend

## User story

As a **runner**,
I want to **choose which specific model each AI backend uses** (e.g. claude-sonnet-4-20250514, gpt-4o-mini, llama3.1)
so that **I can balance response quality, speed, and cost for my coaching sessions**.

## Acceptance criteria

- [ ] Settings store `claude_model`, `openai_model`, `ollama_model` fields in SQLite
- [ ] When a model field is empty, the backend falls back to its hardcoded default
- [ ] `createLLMClient` passes the stored model to each backend constructor
- [ ] Settings UI shows a text input for the active backend's model with placeholder showing the default
- [ ] Onboarding wizard includes model input on the LLM step
- [ ] Saving settings with a new model reloads the LLM client with that model
- [ ] All existing tests continue to pass
- [ ] New tests cover round-trip save/load of model fields

## Technical notes

Each backend already accepts a `Model` field in its config struct and defaults when empty:
- Claude: `claude-sonnet-4-20250514` (in `ClaudeConfig.Model`)
- OpenAI: `gpt-4o` (in `OpenAIConfig.Model`)
- Local: `llama3` (in `LocalConfig.Model`)

Changes required:
1. `internal/storage/migrations.go` — three new `ALTER TABLE` migrations
2. `internal/storage/settings.go` — add fields to `Settings`, update `SaveSettings` / `GetSettings`
3. `app.go` — add fields to `SettingsData`, wire through `createLLMClient`
4. `main.go` — update `createLLMClient` to pass model fields
5. `frontend/src/Settings.svelte` — model text input per backend
6. `frontend/src/Onboarding.svelte` — model text input on step 2

Free-text input (not a dropdown) because model names change frequently and users may use custom/fine-tuned models.

## Tests required

- Unit: Save settings with model fields, read them back, verify defaults when empty
- Integration: `createLLMClient` picks up stored model value
- Edge cases: Empty model string falls back to default, whitespace-only model treated as empty

## Out of scope

- Model validation against provider API (user responsibility)
- Auto-fetching available models from provider APIs
- Per-conversation model switching

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |
| 2026-03-16 | in-progress | Implementation started |
| 2026-03-16 | done | All fields wired end-to-end, tests passing |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
