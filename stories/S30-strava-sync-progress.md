---
id: S30
title: Manual Strava sync with progress indicator
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S30 — Manual Strava sync with progress indicator

## User story

As a **runner**,
I want to manually trigger a Strava activity sync and see progress in real time
so that I know what's happening during the import and don't wonder if the app is stuck.

## Acceptance criteria

- [ ] "Sync Activities" button on Dashboard (only visible when Strava is connected)
- [ ] Clicking the button calls `SyncStravaActivities` binding
- [ ] Progress bar or status text updates in real time via Wails events
- [ ] Events emitted: `strava:sync:start`, `strava:sync:progress` (current/total), `strava:sync:complete`, `strava:sync:error`
- [ ] Sync pages through `GET /api/v3/athlete/activities?page=N&per_page=30`
- [ ] Each activity is deduplication-checked via `GetActivityByStravaID` before insert
- [ ] Token refresh handled automatically when tokens are expired
- [ ] Activity list refreshes after sync completes
- [ ] Error states shown to user (no Strava tokens, API errors, network failures)
- [ ] Button disabled during active sync

## Technical notes

- Create `internal/strava/sync.go` with `FetchAthleteActivities` function
- New exported `StravaActivitySummary` struct for list endpoint response (list endpoint returns different fields than detail endpoint)
- New Wails binding in `app.go`: `SyncStravaActivities() error`
  - Get encrypted tokens from DB, decrypt
  - Check expiry, refresh if needed
  - Page through activities list
  - Emit Wails events for progress
  - Save new activities to DB
- Frontend: listen to Wails events via `EventsOn` from `../wailsjs/runtime/runtime.js`
- Strava list endpoint returns: id, name, type, start_date, distance, moving_time, average_speed, average_heartrate, max_heartrate, average_cadence
- Token decrypt: `encKey := sha256.Sum256([]byte("coachlm-encryption-key"))` then `oauthClient.DecryptToken(encryptedToken)`

## Tests required

- Unit: `FetchAthleteActivities` returns activities from mock API
- Unit: `FetchAthleteActivities` paginates until empty page
- Unit: `FetchAthleteActivities` handles API errors gracefully
- Unit: `FetchAthleteActivities` sends correct Authorization header

## Out of scope

- Webhook-based sync (already exists in S02)
- Syncing activity streams (HR, pace, cadence data points)
- Selective sync (syncing only specific date ranges)

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
