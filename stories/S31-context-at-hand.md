---
id: S31
title: Context at hand — profile in onboarding + auto-rebuild after sync
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S31 — Context at hand — profile in onboarding + auto-rebuild after sync

## User story

As a **runner using CoachLM for the first time**,
I want to **enter my athlete profile during onboarding and have my context automatically built after Strava sync**
so that **my AI coach has all relevant context from the very first conversation**.

## Problem

Today the context engine only assembles data lazily when the user sends a chat message. Two critical flows leave the context empty:

1. **Onboarding wizard** — collects LLM backend and Strava credentials but never asks for athlete profile data (age, max HR, threshold pace, weekly mileage target, race goals, injury history). The profile block stays empty until the user manually visits the Context tab.
2. **Post-Strava sync** — activities are stored in SQLite but the training summary is never pre-built. The user has no indication of what the LLM will actually "see" after import.

## Acceptance criteria

### Onboarding — athlete profile step

- [ ] A new step is added to the onboarding wizard between the Strava step (3) and the finish step (4), making the wizard 5 steps total
- [ ] The new step collects: age, max HR, threshold pace (min:sec/km input), weekly mileage target (km), race goals (text), injury history (text)
- [ ] All fields are optional — the user can skip the entire step
- [ ] On "Next" or "Skip", the profile is saved via the existing `SaveProfileData` Wails binding
- [ ] The progress dots update to reflect 5 steps instead of 4
- [ ] If the user connected Strava in the previous step, a background activity sync is triggered automatically before the profile step (so the context preview on the finish screen can show training data)

### Post-Strava sync — context rebuild + preview

- [ ] After `SyncStravaActivities` completes (both from onboarding and from Dashboard), a new Wails event `strava:sync:context-ready` is emitted containing a preview of the assembled context
- [ ] The preview includes: profile summary (or "No profile configured"), training summary (4-week rolling), pinned insights count
- [ ] A new Wails binding `GetContextPreview() (string, error)` is added that returns the fully assembled system prompt the LLM would receive
- [ ] The Dashboard sync completion state shows a brief context summary (e.g., "Context updated: 12 activities across 3 weeks, profile loaded")
- [ ] The onboarding finish step (step 5) shows a context readiness indicator: which context blocks are populated (profile ✓/✗, training data ✓/✗, insights ✓/✗)

### Context freshness

- [ ] After every Strava sync, the context tab (S29) reflects the latest training summary without requiring a page reload — it listens for the `strava:sync:context-ready` event

## Technical notes

### Backend

- `GetContextPreview()` in `app.go`: calls `coachctx.AssemblePrompt` with current profile, activities (28 days), and insights. Returns the formatted string.
- After sync completes in `SyncStravaActivities`, call `GetContextPreview()` and emit the result as `strava:sync:context-ready` event payload.
- No changes to the context engine itself (`internal/context/`) — this story only wires existing functionality into the onboarding and sync flows.

### Frontend

- `Onboarding.svelte`: add step between current step 3 (Strava) and step 4 (Finish). New step uses a form identical to the profile section in `Context.svelte` (reuse or extract shared component). Update step count from 4 → 5.
- Threshold pace input: two number fields (minutes + seconds) converted to total seconds before saving.
- `Dashboard.svelte`: on `strava:sync:complete`, show a one-line context summary below the progress bar.
- `Context.svelte`: listen for `strava:sync:context-ready` and refresh displayed data.

### Existing bindings to use

- `SaveProfileData(data ProfileData)` — already exists
- `GetProfileData()` — already exists
- `SyncStravaActivities()` — already exists, emits progress events
- New: `GetContextPreview() (string, error)`

## Tests required

- Unit: `GetContextPreview` returns valid prompt with profile + activities + insights
- Unit: `GetContextPreview` returns valid prompt with no profile (graceful fallback)
- Unit: `GetContextPreview` returns valid prompt with no activities ("No recent training data")
- Integration: after `SyncStravaActivities`, `strava:sync:context-ready` event is emitted with non-empty preview
- Frontend: onboarding wizard has 5 steps, profile step saves data correctly
- Frontend: skipping profile step does not error, proceeds to finish

## Out of scope

- Modifying the context engine's compression logic or token budget (covered by S05–S08)
- Adding new profile fields beyond what `ProfileData` already supports
- Auto-syncing Strava on app startup (separate story)
- Editing the training summary directly (the Context tab already handles insights and profile)

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
