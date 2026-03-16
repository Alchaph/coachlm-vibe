---
id: S04
title: Athlete profile setup
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S04 — Athlete profile setup

## User story

As a **runner**,
I want to **enter my profile data (age, max HR, threshold pace, goals, injury history)**
so that **the coaching AI has accurate context about me**.

## Acceptance criteria

- [ ] Store profile fields in SQLite: age, max HR, threshold pace, weekly mileage target, race goals, injury history
- [ ] Validate input ranges (e.g., age 1-120, HR 100-220)
- [ ] Update existing profile (not just create)
- [ ] Profile accessible to context engine for assembly
- [ ] Fields are typed: age (int), max HR (int), threshold pace (duration), goals (text), injuries (text with dates)

## Technical notes

Lives in `internal/storage/`. Table: `athlete_profile`. Structured record, not free-text. No Strava dependency — manual input only.

## Tests required

- Unit: field validation, CRUD
- Integration: save → retrieve round-trip
- Edge cases: empty profile, partial update, negative age, missing optional fields

## Out of scope

Auto-detection from activities, profile UI form (S16), LLM-suggested changes

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
