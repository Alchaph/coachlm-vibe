---
id: S03
title: Activity stream ingestion
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S03 — Activity stream ingestion

## User story

As a **runner**,
I want **my detailed activity data (heart rate, pace, cadence) stored locally**
so that **the coaching AI can analyze my training patterns**.

## Acceptance criteria

- [ ] Fetch HR, pace, and cadence streams from Strava API
- [ ] Store per-second data points in SQLite `activity_streams` table
- [ ] Handle activities with missing streams gracefully
- [ ] Map to consistent schema usable by Strava sync and FIT import (S17)
- [ ] Store activity summary metadata in `activities` table
- [ ] Support non-running activity types without crashing
- [ ] Table `activities` includes `activity_id` and `strava_id` columns

## Technical notes

Lives in `internal/strava/` for fetch and `internal/storage/` for persistence.
Tables: `activities` (summary) and `activity_streams` (per-second).
Schema maps to same structure as S17 (FIT import).
Depends on S01 (tokens) and S02 (webhook triggers).

## Tests required

- Unit: stream parsing, schema mapping, missing field handling
- Integration: fetch → store with mock API
- Edge cases: no HR data, 10+ hour activity, zero-length activity, non-running types

## Out of scope

Activity analysis/statistics, dashboard display (S15), FIT parsing (S17)

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
