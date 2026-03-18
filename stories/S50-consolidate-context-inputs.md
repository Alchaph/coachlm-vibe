---
id: S50
title: Consolidate user context inputs into Context tab
status: done
created: 2026-03-18
updated: 2026-03-18
---

# S50 — Consolidate user context inputs into Context tab

## User story

As a runner using CoachLM, I want all my athlete profile and context-related inputs in a single dedicated Context tab, so that I have a clear, unified place to manage my personal information and avoid confusion from duplicated inputs across the application.

## Acceptance criteria

- [ ] Remove duplicate Athlete Profile form from Onboarding wizard
- [ ] Onboarding should guide new users to complete their profile in the Context tab instead
- [ ] Context tab remains the canonical location for Athlete Profile fields (age, maxHR, threshold pace, weekly mileage, race goals, injury history, experience level, training days, resting HR, preferred terrain)
- [ ] Settings tab retains LLM model selection (useLocalModel, ollamaEndpoint, ollamaModel) and Strava connection actions
- [ ] Custom System Prompt remains in Settings tab (or moves to Context tab - decide based on UX evaluation)
- [ ] Context Export/Import buttons remain in Settings tab (or move to Context tab - decide based on UX evaluation)
- [ ] Pinned Insights remain displayed and managed in Context tab (already correctly placed)
- [ ] Navigation works correctly: onboarding completion → Context tab
- [ ] All tests pass: `go test ./...` and Playwright e2e tests
- [ ] No data loss: existing profiles saved via onboarding continue to work

## Technical notes

### Current state

**Athlete Profile fields** (currently in two places):
- `frontend/src/Context.svelte` — canonical profile form
- `frontend/src/Onboarding.svelte` — duplicate profile step (variables: profileAge, profileMaxHR, profileThresholdMins, profileThresholdSecs, profileWeeklyMileage, profileRaceGoals, profileInjuryHistory, profileExperienceLevel, profileTrainingDaysPerWeek, profileRestingHR, profilePreferredTerrain)

**Backend bindings**:
- `app.go`: `GetProfileData()`, `SaveProfileData(ProfileData)`
- `internal/storage/profile.go`: `AthleteProfile` struct, `SaveProfile`, `GetProfile`
- Validation in `ValidateProfile`: Age 1-120, MaxHR 100-220, ThresholdPaceSecs > 0

**Settings/Preferences** (in Settings.svelte):
- LLM model toggle (useLocalModel checkbox)
- Ollama endpoint (ollamaEndpoint input)
- Ollama model (ollamaModel input)
- Custom System Prompt (customSystemPrompt textarea)
- Context Export/Import buttons (`ExportContext`, `ImportContext`)

### Files to change

**Frontend**:
- `frontend/src/Onboarding.svelte`
  - Remove the profile step with all profile input bindings
  - Replace with a simple instruction: "Complete your athlete profile in the Context tab"
  - Add a prominent button/link to open Context tab (e.g., "Go to Context")
  - On completion, navigate user to `activeTab = 'context'`
  - Remove `saveProfile()` call that invokes `SaveProfileData` RPC during onboarding
  - Keep `checkContextReadiness()` for determining when onboarding should show (may need to adjust logic since profile won't be saved during onboarding)

- `frontend/src/Context.svelte`
  - No changes needed if keeping as-is; already the canonical profile form
  - If moving Custom System Prompt here: add textarea for customSystemPrompt, wire to `SaveSettingsData()`
  - If moving Export/Import here: add buttons, wire to `ExportContext`/`ImportContext` RPCs

- `frontend/src/Settings.svelte`
  - Keep LLM model selection (useLocalModel, ollamaEndpoint, ollamaModel)
  - Keep Strava connect/disconnect actions
  - If moving Custom System Prompt to Context: remove from Settings
  - If moving Export/Import to Context: remove from Settings

- `frontend/src/App.svelte`
  - Ensure navigation to Context tab works from onboarding completion
  - Verify `activeTab` state transitions correctly

**Backend**:
- `app.go`: No changes needed if RPCs stay the same
- `internal/storage/profile.go`: No changes needed
- `internal/storage/settings.go`: No changes needed

### Validation and UX considerations

**Onboarding behavior change**:
- Before: Onboarding required profile completion before proceeding
- After: Onboarding shows a "go to Context" call-to-action; users can skip and complete profile later

**Optional vs required fields**:
- Onboarding currently saves profile even with empty fields (hasProfile boolean guards readiness)
- Context.svelte currently saves all fields; backend validation may reject invalid data
- Decide: Should Context tab allow empty profile? If yes, adjust backend validation or UX to make certain fields optional

**Custom System Prompt placement**:
- Option A: Keep in Settings tab (LLM configuration lives in Settings)
- Option B: Move to Context tab (it's part of coaching context, more discoverable)
- Evaluate UX and decide in implementation

**Export/Import placement**:
- Option A: Keep in Settings tab (context backup feels like a utility/setting)
- Option B: Move to Context tab (context management feels more natural here)
- Evaluate UX and decide in implementation

### Testing requirements

**Unit tests**:
- Verify `SaveProfileData` still works after UI changes
- Verify `GetProfileData` returns correct data
- Test profile validation logic (edge cases: empty threshold, invalid age/HR)

**Integration tests**:
- Verify onboarding completion without profile save works
- Verify navigation to Context tab after onboarding
- Verify profile save from Context tab persists to DB
- Verify LLM settings save still works from Settings tab

**E2e tests (Playwright)**:
- Test new onboarding flow (no profile form, link to Context)
- Test Context tab profile form saves correctly
- Test Settings tab LLM configuration saves correctly
- Test that existing profiles still load

## Out of scope

- Backend storage schema changes (no changes to DB tables)
- New RPC endpoints or modifications to existing ones
- Moving Strava connection UI (stays in Settings)
- Moving chat UI (stays separate)
- Moving dashboard/activity views (stays separate)

---

## Status history

| Date | Status | Notes |
|------|--------|-------|
| 2026-03-18 | draft | Created story based on codebase analysis |

<!-- Agent: add a Blocker section here if status is set to failed -->
