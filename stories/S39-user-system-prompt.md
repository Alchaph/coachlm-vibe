---
id: S39
title: User-defined system prompt additions
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S39 — User-defined system prompt additions

## User story

As a **power user who knows what they're doing**,
I want to **add my own instructions to the system prompt**
so that **I can customize CoachLM's behavior without waiting for feature requests** (e.g., "always give me pace in min/km", "respond in German", "never suggest supplements", "focus on injury prevention").

## Problem

The system prompt is hardcoded in the Go backend (`assembler.go`). Users can't customize behavior even for simple preferences. Someone who wants responses in a specific language, unit system, or communication style has no way to express that. Power users are blocked from tailoring the coach to their needs.

## Acceptance criteria

- [ ] A new settings field `CustomSystemPrompt` (TEXT) stores the user's additional instructions
- [ ] Settings UI includes a textarea for custom prompt additions, with a placeholder like "Add your own instructions, e.g., 'Always respond in German' or 'Never suggest supplements'"
- [ ] Custom prompt is appended to the system preamble in the context assembler — AFTER the built-in framework but BEFORE profile/insights/training blocks
- [ ] Custom prompt is included in the token budget calculation (if it pushes over budget, it gets truncated like profile/training, but after the core framework)
- [ ] Custom prompt survives app restart (persisted in SQLite)
- [ ] If the custom prompt is empty, no extra text is added (no change to existing behavior)
- [ ] The custom prompt is shown as a separate section in the assembled prompt, clearly labeled so the LLM knows it's user input

## Technical notes

- `internal/storage/settings.go`: Add `CustomSystemPrompt string` column to the settings table (or use a new table if settings schema is fixed-arity)
- `app.go`: Add `CustomSystemPrompt` to `SettingsData` struct; wire through `SaveSettings` and `LoadSettings`
- `internal/context/assembler.go`: Add `CustomPrompt` field to `PromptInput`; modify `AssemblePrompt` to append it after the core framework:
  ```go
  customBlock := ""
  if input.CustomPrompt != "" {
      customBlock = "## Custom Instructions\n" + input.CustomPrompt
  }
  // Then include in budget and assembly
  ```
- Update `app.go` where it calls `AssemblePrompt` to pass the custom prompt from settings
- Frontend:
  - `Settings.svelte`: Add textarea for custom prompt (3-4 rows, monospace font optional)
  - Persist when user clicks "Save Settings"
  - No validation — any string is allowed (user responsibility)
- Label suggestion in the assembled prompt: `## Your Custom Instructions` — this helps the LLM understand it's user-provided

## Tests required

- Unit: `assembler_test.go` — custom prompt appears in assembled output when provided
- Unit: custom prompt is truncated if it exceeds token budget (after core framework but before profile)
- Unit: empty custom prompt produces no extra section
- Integration: save custom prompt in Settings, restart app, verify it loads and appears in prompt
- Edge case: very long custom prompt (>1000 chars) — should truncate gracefully
- Edge case: custom prompt contains prompt injection attempts — the LLM receives it as-is; backend doesn't sanitize (trust the LLM's inherent safety training; if user wants to jailbreak themselves, that's on them)

## Out of scope

- Multiple custom prompt slots (e.g., "training style", "response format") — one field is sufficient for v1
- Custom prompt templates / presets
- Import/export custom prompts
- Syntax highlighting or validation in the textarea

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
