---
id: S06
title: Context engine — rolling training summary
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S06 — Context engine — rolling training summary

## User story

As a **system**,
I want to **maintain a rolling summary of the last 4 weeks of training**
so that **coaching advice reflects current fitness and training load**.

## Acceptance criteria

- [ ] Summarize last 4 weeks of activities from SQLite
- [ ] CRITICAL: Older training summaries must be compressed before recent ones (AGENTS.md constraint)
- [ ] Auto-update summary when new activities arrive
- [ ] Output fits within allocated portion of configurable token budget
- [ ] Compression levels defined: Week 1 (most recent) = per-run detail, Week 4 = weekly totals only
- [ ] Handle partial weeks correctly

## Technical notes

Lives in `internal/context/`. "Compression" = reducing detail level (algorithmic), NOT file/data compression. Week 1: per-run summary. Week 2: daily aggregates. Week 3: key sessions only. Week 4: weekly totals. Depends on S03 (activity data). Output is the "training summary block."

## Tests required

- Unit: compression at each level, summary generation
- Integration: activities → summary pipeline
- Edge cases: fewer than 4 weeks, no activities, 100+ activities in one week

## Out of scope

Real-time updates, history beyond 4 weeks, LLM-based summarization

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
