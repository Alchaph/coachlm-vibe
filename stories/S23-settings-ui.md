---
id: S23
title: Settings UI with LLM configuration
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S23 — Settings UI with LLM configuration

## User story

As a **runner**,
I want to **open a Settings tab to configure my LLM backend and API keys**
so that **I can switch between Claude, ChatGPT, and Ollama without editing files**.

## Acceptance criteria

- [ ] New "Settings" tab in the tab bar (alongside Chat and Dashboard)
- [ ] Settings.svelte component with sections for LLM configuration
- [ ] Dropdown to select active LLM backend (Claude / OpenAI / Local)
- [ ] Input fields for Claude API key and OpenAI API key (masked with show/hide toggle)
- [ ] Input field for Ollama endpoint URL (shown only when Local is selected)
- [ ] Save button that persists settings via Wails binding
- [ ] Load existing settings on mount and populate form
- [ ] Success/error feedback after save
- [ ] Wails bindings in app.go: `GetSettings()` and `SaveSettings()` exposed to frontend
- [ ] Saving settings hot-swaps the active LLM client (no app restart needed)
- [ ] API keys encrypted before storage (uses existing storage.SaveSettings)
- [ ] Matches existing dark theme and component style

## Technical notes

Backend storage layer already exists (`internal/storage/settings.go`). This story adds:
1. Wails bindings in `app.go` to expose `GetSettings` and `SaveSettings` to the frontend
2. `frontend/src/Settings.svelte` component
3. Tab entry in `App.svelte`

The `SaveSettings` binding must:
- Encrypt API keys using the existing crypto module before saving
- Recreate the LLM client with new settings (hot-swap)
- Return error on invalid config

The `GetSettings` binding must:
- Return decrypted API keys (masked in frontend, not in backend)
- Return nil/empty if no settings exist yet

Encryption key derivation: use a deterministic key from the data directory path (same approach as existing token encryption).

## Tests required

- Unit: GetSettings returns nil when no settings, SaveSettings round-trip
- Unit: Settings.svelte renders all fields, save triggers binding
- Integration: save settings → reload → settings persisted
- Edge cases: empty API key, switch LLM mid-chat, invalid endpoint URL

## Out of scope

Strava connection UI (S24), onboarding wizard (S25), API key validation against provider

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
