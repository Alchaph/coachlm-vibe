---
id: S33
title: Enhanced athlete context and pace input
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S33 — Enhanced athlete context and pace input

## User story

As a **runner setting up my coaching profile**,
I want **more profile fields to describe my training background, and the threshold pace input in min:sec format everywhere**
so that **the LLM coach has richer context about me and I don't have to convert my pace into raw seconds**.

## Problem

1. **Pace input inconsistency**: The onboarding wizard (step 4) already uses a friendly min:sec threshold pace input, but the Context tab still shows a raw "Threshold Pace (sec/km)" number field. A user seeing `300` has no idea that means `5:00/km`.

2. **Missing context fields**: The current profile captures age, max HR, threshold pace, weekly mileage, race goals, and injury history. Important coaching signals are missing — the LLM doesn't know how experienced the runner is, how many days per week they train, what surfaces they run on, or their resting heart rate. Adding these gives the coach significantly better context for personalized advice.

## Acceptance criteria

- [ ] Context tab (`Context.svelte`) threshold pace input changed from a single "sec/km" number field to the same min:sec two-field format used in the onboarding wizard
- [ ] New profile fields added to backend and both UI surfaces (Context tab + onboarding wizard step 4):
  - **Experience level** — select: beginner / intermediate / advanced / elite
  - **Training days per week** — number (1–7)
  - **Resting heart rate** — number (30–120 bpm)
  - **Preferred terrain** — select: road / trail / track / mixed
- [ ] Backend `AthleteProfile` struct, SQLite schema (migration), and `ProfileData` DTO updated with the four new fields
- [ ] New fields are included in the context engine prompt assembly so the LLM receives them
- [ ] All new fields are optional — profile validation must not reject a profile missing them
- [ ] Existing profiles (from before this migration) continue to load without errors; new fields default to zero-values
- [ ] Onboarding wizard step 4 asks all profile fields (existing + new)

## Technical notes

- `internal/storage/profile.go`: Add `ExperienceLevel string`, `TrainingDaysPerWeek int`, `RestingHR int`, `PreferredTerrain string` to `AthleteProfile`
- `internal/storage/migrations.go`: Add ALTER TABLE migration to add the four new columns with sensible defaults (empty string / 0)
- `app.go`: Extend `ProfileData` struct with the new JSON fields; update `GetProfileData` and `SaveProfileData` mappings
- `frontend/wailsjs/`: Regenerate or manually update Wails bindings if needed
- `Context.svelte`: Replace `thresholdPaceSecs` number input with two inputs (min + sec) and a `:` separator, matching the onboarding wizard pattern; add inputs for the four new fields
- `Onboarding.svelte`: Add inputs for the four new fields in step 4
- `internal/context/`: Update prompt assembler to include new fields when non-empty/non-zero
- Experience and terrain are `<select>` dropdowns with predefined options, not free text

## Tests required

- Unit: `internal/storage/profile_test.go` — save and retrieve profile with new fields; verify defaults for missing fields
- Unit: `app_test.go` — round-trip `SaveProfileData` / `GetProfileData` with new fields
- Unit: context engine — assembled prompt includes new fields when set, omits them when empty
- Edge case: existing profile (pre-migration) loads without error, new fields return zero-values
- Edge case: profile with only some new fields set — only populated fields appear in context

## Out of scope

- Heart rate zone calculation from resting + max HR (separate story)
- Auto-detecting experience level from activity history
- Per-activity terrain tagging

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
