---
id: S40
title: Chat sessions cloud sync
status: draft
created: 2026-03-17
updated: 2026-03-17
---

# S40 — Chat sessions cloud sync

## User story

As a **runner using CoachLM on multiple devices**,
I want **my chat history to sync automatically across devices**
so that **I never lose coaching conversations and can continue discussions from anywhere**.

## Problem

Today, chat sessions are stored locally in SQLite only. If a user reinstalls the app, switches devices, or uses CoachLM on multiple computers, their chat history is lost or fragmented. The context sync (S21) backs up profile, activities, and insights, but not the actual conversation history that led to those insights.

## Acceptance criteria

- [ ] Chat sessions are included in cloud sync alongside context data
- [ ] On app launch with cloud sync enabled, the app checks for newer remote chat history
- [ ] Chat history sync triggers after every new message (debounced 30s) or after closing a session
- [ ] Conflict resolution: remote timestamp wins by default; user can override to keep local
- [ ] Chat export format includes: session ID, timestamp, messages (role, content, timestamp), and associated session metadata
- [ ] Chat import merges with local history; duplicate sessions detected by session ID and timestamp
- [ ] When pulling remote chat history, the app shows a summary of what changed (e.g., "3 new sessions, 12 updated sessions")
- [ ] All sync operations run in background; UI never blocks

## Technical notes

Extend `internal/cloudsync/` from S21 to sync chat sessions:
- Remote file key: `coachlm/chat_sessions.coachctx`
- Format: JSON envelope similar to S20, with `sessions` array

Chat session schema in export:
```json
{
  "version": "1",
  "exported_at": "2026-03-17T...",
  "sessions": [
    {
      "id": "session-123",
      "created_at": "2026-03-15T...",
      "messages": [
        {"role": "user", "content": "...", "timestamp": "..."},
        {"role": "assistant", "content": "...", "timestamp": "..."}
      ]
    }
  ]
}
```

Wails bindings (extend `app.go`):
- `ExportChatSessions() ([]byte, error)` — returns JSON export
- `ImportChatSessions(data []byte, replaceAll bool) error` — merge or replace sessions

Conflict detection logic:
- Compare session timestamps; remote wins if newer
- If timestamps equal (rare), keep local (local gets last write preference for active session)

Sync state tracking: add `last_chat_sync_at` to `cloud_sync_state` table

## Tests required

- Unit: chat session export format validates against schema
- Unit: import merge logic correctly handles duplicates, timestamps
- Unit: conflict resolution respects timestamp comparison
- Integration: round-trip through cloud provider preserves all data
- Edge cases:
  - Empty chat history (no sessions) — export produces valid JSON with empty sessions array
  - Very large chat history (> 1000 messages) — sync completes, no data loss
  - Corrupted remote file — sync logs error, preserves local data
  - Concurrent message send + sync — no race conditions, data integrity maintained

## Out of scope

- Real-time sync during active chat (async only)
- Chat search or indexing
- Chat export to PDF or markdown formats
- Selective sync (e.g., sync only starred sessions)

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-17 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
