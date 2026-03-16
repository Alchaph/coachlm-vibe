---
id: S02
title: Strava webhook receiver
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S02 — Strava webhook receiver

## User story

As a **system**,
I want to **receive Strava webhook notifications**
so that **new activities are synced automatically**.

## Acceptance criteria

- [ ] Webhook endpoint responds within 2 seconds
- [ ] Webhook subscription validation via GET challenge-response
- [ ] Async activity stream fetch after webhook acknowledgment
- [ ] Deduplication by activity ID before processing
- [ ] Handle webhook event types for create, update, and delete
- [ ] Verify webhook signature for security
- [ ] References `activities` table for dedup check

## Technical notes

Lives in `internal/strava/`.
Handler must respond immediately then spawn goroutine for processing.
Deduplication check against `activities` table.
Depends on S01 for valid OAuth tokens.
Strava requirement: respond within 2 seconds.

## Tests required

- Unit: signature validation, dedup logic, challenge response
- Integration: webhook → async fetch with mock Strava
- Edge cases: duplicate webhooks, malformed payload, rapid successive events, 429 rate limits

## Out of scope

Activity data parsing (S03), webhook subscription creation, UI notifications

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
