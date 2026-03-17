---
id: S34
title: Free cloud LLM backend for zero-setup onboarding
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S34 — Free cloud LLM backend for zero-setup onboarding

## User story

As a **new user who just downloaded CoachLM**,
I want **a free cloud LLM that works immediately without API keys or local model installation**
so that **I can start chatting with my running coach in seconds instead of wrestling with Ollama setup or API billing**.

## Problem

Today every LLM backend requires setup: Claude and OpenAI need paid API keys; Ollama needs a local install + model pull. A first-time user hitting the onboarding wizard has no zero-friction option — they either pay upfront or install Ollama (which requires downloading multi-GB models). This creates a steep barrier for casual runners who just want to try the app.

## Acceptance criteria

- [ ] A new LLM backend `free` is available that calls a free-tier cloud API (Google Gemini free tier via the Gemini API, using `gemini-2.0-flash`) — no API key required from the user; the app ships with a built-in key or uses the free unauthenticated endpoint
- [ ] The `free` backend implements the existing `llm.LLM` interface (`Chat` + `Name`)
- [ ] The `free` backend is the **default** selection in the onboarding wizard (step 2) — user can switch to Claude/OpenAI/Ollama if they prefer
- [ ] Settings UI (`Settings.svelte`) includes the `free` option in the backend selector; when selected, no configuration fields are shown (no API key, no endpoint, no model picker)
- [ ] Onboarding wizard step 2 shows the `free` option first with a subtitle like "No setup required" to signal it's the easiest path
- [ ] If the free API is unreachable or rate-limited, the error message suggests switching to another backend
- [ ] The built-in API key (if any) is stored in the binary, not in plaintext source — use a build-time variable or embed
- [ ] Chat quality is acceptable for coaching: the model must handle multi-turn conversation with the system prompt context

## Technical notes

- **Recommended provider**: Google Gemini free tier — generous rate limits (15 RPM, 1M TPM for flash), no billing required, supports system instructions
- `internal/llm/free.go`: New file implementing `LLM` interface. Calls `https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent` with the built-in API key
- Convert `llm.Message` roles to Gemini format: `system` → `system_instruction`, `user` → `user`, `assistant` → `model`
- `internal/llm/llm.go`: No changes to the interface
- `app.go`: Add `"free"` case to the LLM router switch in `initLLM` or equivalent
- `internal/storage/settings.go`: `active_llm` column already stores a string — `"free"` is a valid new value, no schema change needed
- **API key strategy**: Use `go:embed` or `-ldflags -X` to inject the key at build time. The key should NOT appear in source control — add it via CI secret or `.env` build file. For development, fall back to `GEMINI_API_KEY` env var.
- Frontend: Add `<option value="free">Free (Gemini Flash)</option>` to both `Settings.svelte` and `Onboarding.svelte` backend selectors. When `activeLlm === 'free'`, show no config fields.

## Tests required

- Unit: `internal/llm/free_test.go` — mock HTTP response, verify request format (system instruction, message mapping, API key header)
- Unit: verify `Name()` returns `"free"` (or `"Gemini Flash"`)
- Unit: error handling — API returns 429 rate limit → meaningful error message
- Unit: error handling — API unreachable → timeout and clear error
- Integration: settings round-trip — save `active_llm = "free"`, reload, verify backend initializes
- Edge case: empty conversation (no user messages) — should not crash

## Out of scope

- Streaming responses (separate story if wanted)
- Usage tracking / rate limit UI
- Multiple free providers (only Gemini for now)
- Fallback to another backend on failure (user must switch manually)
- Gemini safety filter configuration

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
