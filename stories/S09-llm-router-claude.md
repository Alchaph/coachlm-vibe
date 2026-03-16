---
id: S09
title: LLM router interface and Claude implementation
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S09 — LLM router interface and Claude implementation

## User story

As a **system**,
I want to **route chat messages to Claude API**
so that **runners get AI coaching responses**.

## Acceptance criteria

- [ ] Define the `LLM` interface with `Chat(ctx context.Context, messages []Message) (string, error)` and `Name() string`
- [ ] Define the `Message` struct with `Role` (system/user/assistant) and `Content` fields
- [ ] Implement Claude backend using Anthropic Messages API
- [ ] API key stored securely in SQLite (not plaintext)
- [ ] Handle API errors gracefully (rate limits, network failures, malformed responses)
- [ ] Support configurable model selection via configuration
- [ ] `Name()` returns "claude"

## Technical notes

Lives in `internal/llm/`. Interface defined in AGENTS.md lines 110-113: 
`type LLM interface { Chat(ctx context.Context, messages []Message) (string, error); Name() string }`.
`Message` type requires `Role string` and `Content string`. 
S09 defines the core contract that S10 and S11 must follow. 
Claude implementation uses the Anthropic Messages API endpoint.

## Tests required

- Unit: Message formatting, error wrapping, `Name()` returns "claude"
- Integration: Round-trip with mock Claude API
- Edge cases: Empty messages, missing API key, rate limiting, malformed response, context length exceeded

## Out of scope

- Streaming responses
- Tool or function calling
- Multi-modal input
- Claude-specific prompt optimization

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |
| 2026-03-16 | in-progress | Agent started implementation |
| 2026-03-16 | done | Interface + Claude impl + 14 tests passing |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
