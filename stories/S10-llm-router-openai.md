---
id: S10
title: LLM router — OpenAI implementation
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S10 — LLM router — OpenAI implementation

## User story

As a **system**,
I want to **route chat messages to OpenAI API**
so that **runners have an alternative coaching backend**.

## Acceptance criteria

- [ ] Implements the `LLM` interface defined in S09: `Chat(ctx context.Context, messages []Message) (string, error)` and `Name() string`
- [ ] Maps common `Message` struct to OpenAI chat completion format
- [ ] API key stored securely (not plaintext)
- [ ] Handle API errors gracefully (rate limits, context window, authentication)
- [ ] Support model selection (GPT-4, etc.) via configuration
- [ ] `Name()` returns "openai"

## Technical notes

Lives in `internal/llm/`. 
Must implement the exact interface from S09. 
OpenAI chat completion API uses a compatible message format (role and content). 
Depends on S09 for interface and `Message` type definition.

## Tests required

- Unit: Message mapping, error wrapping, `Name()` returns "openai"
- Integration: Round-trip with mock OpenAI API
- Edge cases: Empty messages, missing API key, rate limiting, token limit exceeded

## Out of scope

- Streaming responses
- Function or tool calling
- Embeddings
- Image generation or fine-tuning

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
