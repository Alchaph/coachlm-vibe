---
id: S15
title: Activity dashboard
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S15 — Activity dashboard

## User story

As a **runner**,
I want to **see my recent activities and basic metrics at a glance**
so that **I can track my training**.

## Acceptance criteria

- [ ] Display list of recent activities (last 20)
- [ ] Show per-activity metrics: date, distance, duration, avg pace, avg HR
- [ ] Activities sorted by date descending
- [ ] Empty state when no activities exist
- [ ] Activity data sourced from SQLite `activities` table
- [ ] Error state if data fetch fails

## Technical notes

Lives in `frontend/`. 
Wails binding in `app.go`: fetches recent activities with a configurable limit, returning a list of activity records or an error. 
Reads from `activities` table (same schema as S03). 
No charting — just a list with summary metrics. 
Depends on S03 (activity data).

## Tests required

- Unit: list rendering, metric formatting
- Integration: fetch → display
- Edge cases: zero activities, missing metrics, 1000+ activities, non-running types

## Out of scope

Detail view, charting/graphs, comparison, training load, route maps

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
