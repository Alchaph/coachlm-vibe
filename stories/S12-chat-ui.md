---
id: S12
title: Chat UI
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S12 — Chat UI

## User story

As a **runner**,
I want to **chat with an AI coach through a conversational interface**
so that **I can get personalized training advice**.

## Acceptance criteria

- [ ] Text input field for sending messages
- [ ] Messages displayed in chronological order
- [ ] AI responses rendered as markdown
- [ ] Loading indicator while waiting for LLM response
- [ ] Enter key sends message
- [ ] Empty messages prevented (validation)
- [ ] Error state shown when backend call fails
- [ ] Empty state shown when no messages exist

## Technical notes

Lives in `frontend/`. Plain Svelte (not SvelteKit — Wails webview). 
Wails binding in `app.go`: sends a message string and returns the LLM response or an error. 
Depends on S09 (LLM router) for backend.

## Tests required

- Unit: message validation, markdown rendering
- Integration: send → receive via Wails binding
- Edge cases: very long message, rapid sends, markdown edge cases like code blocks/tables, first-load empty state

## Out of scope

Voice input, file attachments, message editing/deletion, typing indicators, themes

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |
| 2026-03-16 | in-progress | Implementation started |
| 2026-03-16 | done | Chat UI implemented with stub SendMessage binding |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
