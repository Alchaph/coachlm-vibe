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
1. READ    → Load the story file for the feature you are working on
2. UPDATE  → Set story status to `in-progress`
3. BUILD   → Implement the feature exactly as specified
4. TEST    → Run `go test ./...` (all must pass) + run e2e Playwright tests
5. UPDATE  → Set story status to `done` (or `failed` with notes)
6. COMMIT  → git commit with message format: `feat|fix|docs(SXX): short description`
7. RELEASE → Create and push a semver tag (e.g. `v1.8.2`) — this triggers the release
             pipeline which builds binaries and publishes a documented GitHub Release
             with auto-generated notes
```

If a story file does not exist for what you are about to build, stop and create one first.

### Release versioning

Tags follow `vMAJOR.MINOR.PATCH` semver:

- **PATCH** bump — bug fix story (e.g. `fix:`)
- **MINOR** bump — new feature story (e.g. `feat:`)
- **MAJOR** bump — breaking change (rare; discuss first)

Always check the latest tag before creating a new one:

```bash
git tag --sort=-v:refname | head -5
```

Use the **next semver** after the highest `vX.Y.Z` tag. Do **not** create
`v0.XX` style tags for individual stories — every release tag must be a proper
`vMAJOR.MINOR.PATCH` that advances the sequence.

### E2e tests in CI

The Playwright e2e suite lives in `frontend/e2e/`. Tests run against the Vite
dev server and mock the Wails backend via `frontend/e2e/mocks/wails.ts`.

To run locally (requires Playwright browsers installed):

```bash
cd frontend && npx playwright test
```

CI does **not** currently run e2e tests (no browser installed on the runner).
Before pushing a release tag, verify locally that:
1. `go test ./...` passes
2. The Playwright suite passes (or document why a specific test is skipped)

### Documented release notes

The release workflow (`release.yml`) uses `generate_release_notes: true` which
auto-populates the GitHub Release body from commit messages since the last tag.
Write commit messages that are user-readable — they become the public changelog.

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
