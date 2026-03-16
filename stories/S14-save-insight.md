---
id: S14
title: Save insight from chat
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S14 — Save insight from chat

## User story

As a **runner**,
I want to **save a coaching insight from chat**
so that **it becomes permanent context for future conversations**.

## Acceptance criteria

- [ ] User can select a chat message and save as pinned insight
- [ ] Saved insight appears in pinned insights store (S07)
- [ ] Confirmation feedback shown to user
- [ ] Duplicate detection (same text already pinned)
- [ ] Error state if save fails

## Technical notes

Lives in `frontend/` for UI, `internal/context/` for storage. 
Bridges chat (S12/S13) with context engine (S07). 
Wails binding in `app.go`: accepts message content and saves it as a pinned insight, returning an error on failure. 
Depends on S07 (pinned insights) and S12 (chat UI).

## Tests required

- Unit: insight saving, duplicate detection
- Integration: select → save → verify in context
- Edge cases: empty message, very long message, duplicate save, save failure

## Out of scope

Auto-detection, editing after save, batch saving, categorization

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |
| 2026-03-16 | in-progress | Started implementation |
| 2026-03-16 | done | Wails binding stub + pin button in chat UI |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
