---
id: S11
title: LLM router — local LLM implementation
status: draft
created: 2026-03-16
updated: 2026-03-16
---

# S11 — LLM router — local LLM implementation

## User story

As a **runner**,
I want to **use a local LLM for privacy or offline coaching**
so that **my data never leaves my machine**.

## Acceptance criteria

- [ ] Implements the `LLM` interface defined in S09: `Chat(ctx context.Context, messages []Message) (string, error)` and `Name() string`
- [ ] Connects to Ollama-compatible HTTP endpoint
- [ ] Configurable endpoint URL (default localhost:11434)
- [ ] Handle connection failures gracefully (Ollama not running or port blocked)
- [ ] Support model selection through application configuration
- [ ] `Name()` returns "local"
- [ ] No API key required for local operation

## Technical notes

Lives in `internal/llm/`. 
Ollama exposes its chat API at `/api/chat`. 
Requires no API key. 
Endpoint URL is configurable to allow remote Ollama servers. 
Must handle "Ollama not running" as a specific error case. 
Depends on S09 for interface definition.

## Tests required

- Unit: Message formatting, endpoint URL construction, `Name()` returns "local"
- Integration: Round-trip with mock Ollama server
- Edge cases: Ollama not running, model not downloaded, slow response, very large response

## Out of scope

- Ollama installation or system setup
- Model management (pull or delete)
- GPU configuration
- Embedding API integration

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
