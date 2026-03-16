---
id: S25
title: New user onboarding wizard
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S25 — New user onboarding wizard

## User story

As a **new user**,
I want to **be guided through initial setup when I first launch the app**
so that **I can configure my LLM and optionally connect Strava before using the coach**.

## Acceptance criteria

- [ ] Wizard shown automatically on first launch (no settings saved yet)
- [ ] Step 1: Welcome screen with app description and "Get Started" button
- [ ] Step 2: LLM configuration — select backend (Claude/OpenAI/Local), enter API key or Ollama endpoint
- [ ] Step 3: Strava connection (optional) — enter Client ID/Secret, connect button, or "Skip" option
- [ ] Step 4: Completion screen with "Start Chatting" button
- [ ] Progress indicator showing current step
- [ ] Settings saved at completion (reuses S23 SaveSettings binding)
- [ ] Strava OAuth triggered inline if user chooses to connect (reuses S24 StartStravaAuth binding)
- [ ] Wizard does not show again after completion (settings exist = wizard skipped)
- [ ] Wails binding: `IsFirstRun()` returns true if no settings exist
- [ ] Back/Next navigation between steps
- [ ] Skip button available on optional steps (Strava)
- [ ] Matches existing dark theme

## Technical notes

This is primarily a frontend feature. The wizard is an `Onboarding.svelte` component that:
1. Shows as a full-screen overlay when `IsFirstRun()` returns true
2. Walks through setup steps using local component state
3. At completion, calls `SaveSettings()` (from S23) to persist LLM config
4. Optionally triggers `StartStravaAuth()` (from S24) for Strava connection
5. Once settings are saved, the wizard never appears again

`IsFirstRun` binding in `app.go`: simply checks if `db.GetSettings()` returns nil.

The wizard component renders in `App.svelte` before the main tab UI, conditionally based on `IsFirstRun()`.

## Tests required

- Unit: IsFirstRun returns true when no settings, false after save
- Unit: Onboarding.svelte renders all steps, navigation works
- Integration: complete wizard → settings saved → wizard hidden on reload
- Edge cases: close app mid-wizard, skip all optional steps

## Out of scope

Profile setup (name, age, goals) — that's S04. Import existing data. Account creation.

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
