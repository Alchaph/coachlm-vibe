---
id: S29
title: Context tab with editable profile and insights
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S29 — Context tab with editable profile and insights

## User story

As a **runner**,
I want a dedicated Context tab showing my athlete profile, pinned insights, and training summaries
so that I can see and edit the data my AI coach uses, making the app less of a black box.

## Acceptance criteria

- [ ] New Context tab accessible from sidebar navigation
- [ ] Athlete profile section: editable form for age, max HR, threshold pace, weekly mileage target, race goals, injury history
- [ ] Save button persists profile changes via `SaveProfileData` binding
- [ ] Pinned insights section: list all saved insights with delete button
- [ ] Deleting an insight calls `DeletePinnedInsight` binding and removes it from the list
- [ ] Training summary section: read-only display of recent activities (last 10)
- [ ] Empty states shown when no profile, no insights, or no activities exist
- [ ] Success/error feedback on save and delete actions

## Technical notes

- Create `frontend/src/Context.svelte` component
- New Wails bindings needed in `app.go`:
  - `GetProfileData() (*ProfileData, error)` — wraps `db.GetProfile()`
  - `SaveProfileData(data ProfileData) error` — wraps `db.SaveProfile()`
  - `GetPinnedInsights() ([]InsightData, error)` — wraps `db.GetInsights()`
  - `DeletePinnedInsight(id int64) error` — wraps `db.DeleteInsight()`
- New types in `app.go`:
  - `ProfileData` struct with json tags matching frontend expectations
  - `InsightData` struct with id, content, sourceSessionId, createdAt
- Update Wails TS bindings: `App.d.ts`, `App.js`, `models.ts`
- CRITICAL: Pinned insights are NEVER compressed or dropped (AGENTS.md constraint)

## Tests required

- Unit: `GetProfileData` returns nil when no profile exists
- Unit: `SaveProfileData` validates and persists profile
- Unit: `GetPinnedInsights` returns all insights
- Unit: `DeletePinnedInsight` removes insight by ID
- Unit: `DeletePinnedInsight` returns error for non-existent ID

## Out of scope

- Editing insights inline (only delete is supported)
- Training summary editing
- Context token budget visualization

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
