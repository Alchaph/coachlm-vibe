---
id: S16
title: Settings screen
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S16 — Settings screen

## User story

As a **runner**,
I want to **configure API keys, choose my LLM, and manage my Strava connection**
so that **I can customize the app to my needs**.

## Acceptance criteria

- [ ] Input fields for Claude and OpenAI API keys
- [ ] API keys stored encrypted in SQLite (same encryption as OAuth tokens)
- [ ] Dropdown for active LLM backend (Claude / OpenAI / Local)
- [ ] Ollama endpoint URL configuration field
- [ ] Button to re-authorize Strava
- [ ] Current Strava connection status displayed
- [ ] Error state if settings save fails

## Technical notes

Lives in `frontend/` for UI, `internal/storage/` for encrypted storage. 
Same encryption mechanism as S01 OAuth tokens. 
Wails bindings in `app.go`: save settings, retrieve settings, and re-authorize Strava — each returning an error on failure. 
Relies on the same encryption used in S01 and the LLM backend names from S09. 
Table: `settings`.

## Tests required

- Unit: settings CRUD, validation, key masking
- Integration: save → reload, LLM switch
- Edge cases: invalid API key, empty fields, switch LLM mid-chat, re-auth failure

## Out of scope

Themes, notifications, data export, account deletion, usage stats

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
