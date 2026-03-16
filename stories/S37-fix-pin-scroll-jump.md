---
id: S37
title: Fix chat scroll jumping on pin
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S37 — Fix chat scroll jumping on pin

## User story

As a **user reading chat history**,
I want **the chat to stay where I am when I pin a message**
so that **I don't lose my place in the conversation when saving an insight**.

## Problem

The chat component has an `afterUpdate` lifecycle hook that unconditionally scrolls to the bottom:

```typescript
afterUpdate(() => {
  if (chatContainer) {
    chatContainer.scrollTop = chatContainer.scrollHeight
  }
})
```

This runs on **every** component update. When a user pins a message (clicks the pin button), `pinnedIndices` is updated, triggering a reactivity update, which runs `afterUpdate`, which scrolls to the bottom — even if the user had scrolled up to read older messages.

## Acceptance criteria

- [ ] Pinning a message does NOT scroll the chat to the bottom
- [ ] The chat still auto-scrolls to bottom when:
  - A new message is sent by the user
  - A new response is received from the assistant
  - (This is the only time auto-scroll should happen)
- [ ] Manual scroll position is preserved when unrelated updates occur (e.g., pin, unpin, feedback toast appears)

## Technical notes

- `frontend/src/App.svelte`: Remove or refactor the `afterUpdate` hook
- Instead of unconditional scroll-on-update, only scroll when:
  1. User sends a message (`send()` function)
  2. New assistant message is appended (in the `finally` block after `await SendMessage`)
- Implementation options:
  - Option A (simple): Call `chatContainer.scrollTop = chatContainer.scrollHeight` at the end of `send()` and after appending assistant response
  - Option B (more precise): Track whether the user was at the bottom before update, restore after — but this adds complexity
  - Recommended: Option A — simpler and matches the intended behavior
- The `afterUpdate` hook can be removed entirely once explicit scroll calls are added

## Tests required

- Manual: Pin a message while scrolled up → verify scroll position stays unchanged
- Manual: Send a message → verify chat scrolls to bottom
- Manual: Receive an assistant response → verify chat scrolls to bottom
- Edge case: Pin button shows feedback toast ("Insight saved!") → verify no scroll jump

## Out of scope

- Scrolling to a specific pinned message
- "Jump to latest" button
- Touch / mobile scroll behavior (assume same fix works)

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
