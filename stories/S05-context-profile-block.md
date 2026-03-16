---
id: S05
title: Context engine — profile block assembly
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S05 — Context engine — profile block assembly

## User story

As a **system**,
I want to **assemble the athlete's profile into a structured text block**
so that **LLM conversations have accurate runner context**.

## Acceptance criteria

- [ ] Read profile from `athlete_profile` table
- [ ] Format into structured, human-readable text block
- [ ] Output is deterministic (same input → same output)
- [ ] Include all profile fields with labels
- [ ] Handle missing/optional fields gracefully (omit rather than show empty)

## Technical notes

Lives in `internal/context/`. One block of the assembled context — S08 combines all blocks. Depends on S04 for profile data. Output is the "profile block."

## Tests required

- Unit: full profile formatting, partial profile formatting
- Integration: storage → block
- Edge cases: empty profile, special characters, missing optional fields

## Out of scope

Training summary (S06), pinned insights (S07), full assembly (S08), token counting

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |
| 2026-03-16 | done | Implemented FormatProfileBlock + tests |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
