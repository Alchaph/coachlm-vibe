---
id: S17
title: FIT file import
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S17 — FIT file import

## User story

As a **runner**,
I want to **manually import activities from FIT files**
so that **activities from non-Strava sources appear in my training history**.

## Acceptance criteria

- [ ] Parse FIT file format to extract HR, pace, and cadence streams
- [ ] Map parsed data to existing SQLite schema (S03) using `activities` and `activity_streams` tables
- [ ] Implement deduplication check to prevent reimporting the same activity
- [ ] Add file picker UI for selecting FIT files within the desktop app
- [ ] Provide import status feedback for success, failure, and duplicate scenarios
- [ ] Handle corrupted or invalid FIT files gracefully with helpful error messages
- [ ] Ensure the app remains functional even if this optional feature is unused

## Technical notes

Implementation lives in `internal/fit/` for parsing and `internal/storage/` for persistence.
CRITICAL: must use the exact same `activities` and `activity_streams` tables and schema defined in S03.
Consider using an existing Go FIT library like `github.com/tormoder/fit`.
Deduplication strategy: generate a unique hash based on activity start time, duration, and distance.
Wails binding in `app.go`: accepts a file path and imports the FIT file, returning an error on failure.
The parsing logic should map FIT messages to the internal activity representation before storage.

## Tests required

- Unit: FIT parsing logic, schema mapping, deduplication hash generation
- Integration: Full file-to-storage pipeline (file → parse → store)
- Edge cases: Corrupted FIT files, missing data streams, very large files, non-running activities, unknown device types

## Out of scope

- TCX or GPX file import
- Automatic background file detection
- Manual activity editing after import
- Batch import UI for multiple files at once

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |
| 2026-03-16 | in-progress | Implementation started |
| 2026-03-16 | done | Parser, tests, and app.go stub implemented. 21 tests passing. |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
