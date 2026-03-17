---
id: S35
title: Strava sync overview with activity stats
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S35 — Strava sync overview with activity stats

## User story

As a **runner who syncs activities from Strava**,
I want **a clear overview of how many activities are stored and what the coaching context contains**
so that **I know my data actually imported and the coach has enough training history to work with**.

## Problem

After syncing, the Dashboard shows a transient "Synced N new activities (M total)" message that disappears after 5 seconds. The Context tab shows the 10 most recent activities in a table but no summary stats. There's no persistent, at-a-glance indicator of how much training data the app holds. A user who synced yesterday has no way to quickly confirm "yes, my 200 runs are in here" without counting table rows.

## Acceptance criteria

- [ ] The Dashboard shows a persistent stats bar (above the activity table) displaying: **total activities stored**, **date range** (earliest → latest activity), and **total distance** (sum of all activities, in km)
- [ ] The Context tab "Training Summary" section header shows the **total activity count** next to it (e.g. "Training Summary (142 activities)")
- [ ] A new backend method `GetActivityStats()` returns `{ totalCount int, totalDistanceKm float64, earliestDate string, latestDate string }` — computed via a single SQL query
- [ ] Stats update automatically after a Strava sync completes (listen to the existing `strava:sync:complete` event)
- [ ] Stats load on mount for both Dashboard and Context tab
- [ ] When no activities exist, the stats bar is hidden (don't show "0 activities" — the existing empty state message is sufficient)

## Technical notes

- `internal/storage/activities.go`: Add `GetActivityStats()` method:
  ```sql
  SELECT COUNT(*), COALESCE(SUM(distance), 0),
         MIN(start_date), MAX(start_date)
  FROM activities
  ```
- `app.go`: Add `GetActivityStats()` Wails binding that calls storage and returns a `StatsData` struct
- `frontend/wailsjs/`: Add binding for `GetActivityStats`
- `Dashboard.svelte`: Add a stats row above the table showing total count, date range, total km. Reload stats on `strava:sync:complete`. Style consistently with existing sync bar.
- `Context.svelte`: Update the "Training Summary" `<h2>` to include the count. Call `GetActivityStats()` on mount and on `strava:sync:context-ready`.
- Stats bar design suggestion: a subtle row with 3 metrics separated by dividers, using muted text color (`#94a3b8`), matching existing dashboard style

## Tests required

- Unit: `internal/storage/activities_test.go` — `GetActivityStats` with 0, 1, and multiple activities; verify count, distance sum, date range
- Unit: `app_test.go` — `GetActivityStats` Wails binding returns correct struct shape
- Edge case: no activities → stats method returns zeroes, frontend hides stats bar
- Edge case: single activity → earliest and latest date are the same
- Edge case: activities with 0 distance (e.g. treadmill with no GPS) — should not break sum

## Out of scope

- Per-week / per-month activity breakdown charts
- Activity type distribution (e.g. "120 runs, 22 rides")
- Detailed sync history log
- Export activity data

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
