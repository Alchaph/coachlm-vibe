# AGENTS.md

This file is the contract for any AI agent working on this codebase.
Read it fully before touching any code.

---

## Project overview

A Go desktop app (Wails v2) for runners that:
- Syncs activities automatically via Strava API
- Parses and stores activity metrics in SQLite
- Maintains a structured athlete context (profile, training load, insights)
- Routes chat to Claude, ChatGPT, or a local LLM
- Saves chat insights back into the context for future sessions

---

## Non-negotiable workflow

Every task follows this exact sequence. No exceptions.

```
1. READ   → Load the story file for the feature you are working on
2. UPDATE → Set story status to `in-progress`
3. BUILD  → Implement the feature exactly as specified
4. TEST   → Run the tests; all must pass
5. UPDATE → Set story status to `done` (or `failed` with notes)
```

If a story file does not exist for what you are about to build, stop and create one first.

---

## Repository structure

```
/
├── AGENTS.md               ← you are here
├── stories/                ← one .md file per feature
│   ├── _template.md        ← copy this when creating new stories
│   ├── S01-strava-oauth.md
│   ├── S02-activity-sync.md
│   └── ...
├── frontend/               ← Wails frontend (Svelte)
├── internal/
│   ├── strava/             ← Strava API client + webhook
│   ├── storage/            ← SQLite layer
│   ├── context/            ← Context engine + prompt assembler
│   ├── llm/                ← LLM router (Claude / OpenAI / local)
│   └── fit/                ← FIT file parser (optional import)
├── app.go                  ← Wails app bindings
├── main.go
└── go.mod
```

---

## Story file format

Every story lives in `/stories/SXX-short-name.md`.
Use the template at `/stories/_template.md`.

### Status values

| Status | Meaning |
|---|---|
| `draft` | Written, not started |
| `in-progress` | Agent is actively working on it |
| `done` | Implemented, tested, passing |
| `failed` | Blocked or tests not passing — add notes |
| `skipped` | Deliberately deferred |

### Updating status

At the top of the story file there is a `status:` field in the frontmatter.
Update it by editing that line. Do not change anything else in the frontmatter
unless the story itself has changed scope.

---

## Testing rules

- Every story must have corresponding tests before it is marked `done`
- Unit tests live next to the package they test (`_test.go`)
- Integration tests live in `/tests/integration/`
- Run tests with `go test ./...` from the repo root
- Do not mark a story `done` if any test fails
- If a test is skipped intentionally, leave a comment explaining why

---

## Context engine — special rules

The context engine (`/internal/context/`) is the most sensitive part of the codebase.
Changes here affect every LLM interaction.

- Never modify the prompt template without updating the relevant story
- The assembled context must always fit within the configured token budget
- Older training summaries must be compressed before recent ones
- Pinned insights from chat are never compressed or dropped

---

## LLM router interface

All three LLM backends implement this interface. Do not break it.

```go
type LLM interface {
    Chat(ctx context.Context, messages []Message) (string, error)
    Name() string
}
```

Adding a new backend = implement the interface + add a story.

---

## Strava sync rules

- OAuth tokens are stored encrypted in SQLite, never in plaintext
- Webhook handler must respond within 2 seconds (Strava requirement)
- Activity stream fetch (HR, pace, cadence) happens async after webhook receipt
- Deduplication: check activity ID before inserting

---

## What agents must NOT do

- Do not commit secrets, API keys, or tokens
- Do not skip writing tests to save time
- Do not mark a story `done` without running `go test ./...`
- Do not change the LLM interface without updating all three implementations
- Do not modify another story's status unless you are working on it

---

## When you are blocked

Update the story status to `failed`, add a `## Blocker` section at the bottom
of the story file describing the issue, and stop. Do not guess or work around it silently.
