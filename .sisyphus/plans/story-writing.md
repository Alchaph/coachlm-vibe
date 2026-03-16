# Write 17 User Stories for CoachLM

## TL;DR

> **Quick Summary**: Write 17 user story files for the CoachLM running coach desktop app, covering Strava integration, context engine, LLM routing, chat UI, and FIT import. Each story follows the `_template.md` format exactly.
> 
> **Deliverables**:
> - 17 story files in `stories/S01-*.md` through `stories/S17-*.md`
> - `_template.md` moved into `stories/` directory
> - Summary table of all stories
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES — 3 waves
> **Critical Path**: Setup → Write stories (5 parallel batches) → QA verification

---

## Context

### Original Request
Write one story file per feature for a Go desktop running coach app (Wails v2 + Svelte). 17 features covering Strava sync, athlete context engine, LLM routing (Claude/OpenAI/local), chat UI, and FIT file import. Each story must follow `_template.md` format with specific acceptance criteria, technical notes, test cases, and explicit out-of-scope boundaries.

### Interview Summary
**Key Discussions**:
- Project is brand new — only AGENTS.md and _template.md exist at root
- All 17 stories precisely defined with IDs S01-S17
- Stories must follow template exactly: frontmatter + 7 sections
- Status: `draft`, dates: `2026-03-16`

**Research Findings**:
- AGENTS.md defines LLM interface: `Chat(ctx, []Message) (string, error)` + `Name() string`
- Context engine rules: token budget, older summaries compressed first, pinned insights never dropped
- Strava rules: encrypted tokens in SQLite, <2s webhook response, async stream fetch, dedup by activity ID
- Test rules: unit tests next to package (`_test.go`), integration in `/tests/integration/`, run via `go test ./...`

### Metis Review
**Identified Gaps** (addressed):
- **Message type undefined**: S09 technical notes will specify that the `Message` struct needs definition as part of router interface work
- **Template location mismatch**: AGENTS.md expects `stories/_template.md`, file is at root — Task 1 resolves by moving it
- **Encryption specifics**: Left as implementation detail per story; acceptance criteria say "encrypted" without mandating algorithm
- **Token budget default**: S08 will specify "configurable with sensible default" — not hardcoded
- **Local LLM runtime**: User specified "Ollama-compatible endpoint" — locked in
- **LLM router dispatch**: Included in S09 (router interface + Claude impl)
- **Cross-story consistency**: Domain grouping ensures same agent writes related stories (shared DB columns, API contracts, etc.)

---

## Work Objectives

### Core Objective
Produce 17 well-structured, specific, internally-consistent story files that give an AI implementation agent everything it needs to build each feature without guessing.

### Concrete Deliverables
- `stories/` directory with `_template.md` moved into it
- 17 story files: `stories/S01-strava-oauth.md` through `stories/S17-fit-file-import.md`
- Summary table printed at the end of the final QA task

### Definition of Done
- [ ] All 17 story files exist in `stories/`
- [ ] Every file has valid YAML frontmatter with all 5 fields
- [ ] Every file has all 7 template sections (User story, Acceptance criteria, Technical notes, Tests required, Out of scope, Status history, Blocker comment)
- [ ] Every file has `status: draft` and dates `2026-03-16`
- [ ] No empty "Out of scope" sections
- [ ] Acceptance criteria are specific and testable (no subjective criteria)
- [ ] Technical notes reference file paths from AGENTS.md repository structure
- [ ] Stories sharing database tables use identical column/table names
- [ ] LLM stories (S09-S11) reference the exact interface from AGENTS.md

### Must Have
- Exact template format compliance — no added or removed sections
- Non-empty Out of Scope for every story
- Cross-story consistency within domains (Strava stories share terminology, context engine stories share architecture assumptions)
- Dependencies clearly listed where applicable
- Concrete test cases (not vague "test it works")

### Must NOT Have (Guardrails)
- No Go code in stories (except quoting the LLM interface from AGENTS.md)
- No subjective acceptance criteria ("looks good", "feels fast", "is responsive")
- No more than 8 acceptance criteria per story
- No more than 15 lines in technical notes per story
- No added template sections beyond the 7 defined ones
- No file paths that aren't in AGENTS.md's repository structure
- No stories beyond the 17 specified (no "bonus" stories)
- No UI design specifications (colors, pixel sizes, fonts) — stories specify behavior only
- No aspirational criteria ("should support future X")
- No more than 2 explicit dependencies per story

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: NO (no code yet — this is a writing task)
- **Automated tests**: None (stories are markdown files, not code)
- **Framework**: N/A

### QA Policy
Every task MUST include agent-executed QA scenarios verifying story file structure, content quality, and cross-story consistency.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **File verification**: Use Bash — check file existence, frontmatter parsing, section presence
- **Content verification**: Use Bash (grep) — verify keywords, cross-references, constraint compliance
- **Consistency verification**: Use Bash — compare terminology across domain-grouped stories

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Setup — prerequisite):
└── Task 1: Create stories/ directory, move _template.md [quick]

Wave 2 (Write stories — ALL 5 in parallel):
├── Task 2: Strava domain stories — S01, S02, S03 [writing]
├── Task 3: Athlete profile + Context engine — S04, S05, S06, S07, S08 [writing]
├── Task 4: LLM router stories — S09, S10, S11 [writing]
├── Task 5: Chat + UI stories — S12, S13, S14, S15, S16 [writing]
└── Task 6: FIT file import — S17 [writing]

Wave 3 (Verification — after all writing):
└── Task 7: QA verification + summary table [unspecified-high]

Wave FINAL (After ALL tasks — independent review, 4 parallel):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)

Critical Path: Task 1 → Tasks 2-6 (parallel) → Task 7 → F1-F4
Parallel Speedup: ~60% faster than sequential
Max Concurrent: 5 (Wave 2)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | — | 2, 3, 4, 5, 6 | 1 |
| 2 | 1 | 7 | 2 |
| 3 | 1 | 7 | 2 |
| 4 | 1 | 7 | 2 |
| 5 | 1 | 7 | 2 |
| 6 | 1 | 7 | 2 |
| 7 | 2, 3, 4, 5, 6 | F1-F4 | 3 |
| F1-F4 | 7 | — | FINAL |

### Agent Dispatch Summary

- **Wave 1**: **1 task** — T1 → `quick`
- **Wave 2**: **5 tasks** — T2-T6 → `writing`
- **Wave 3**: **1 task** — T7 → `unspecified-high`
- **FINAL**: **4 tasks** — F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. Create stories/ directory and move template

  **What to do**:
  - Create the `stories/` directory at project root
  - Move `_template.md` from project root to `stories/_template.md`
  - Verify the move was successful (file exists at new location, removed from root)

  **Must NOT do**:
  - Do not copy — move only (no duplicate template files)
  - Do not modify the template content

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Two shell commands — mkdir + mv. Trivial setup task.
  - **Skills**: []
    - No skills needed for file system operations.

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1 (sole task)
  - **Blocks**: Tasks 2, 3, 4, 5, 6
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References**:
  - `AGENTS.md:37-54` — Repository structure showing `stories/` directory and `_template.md` expected location
  - `AGENTS.md:61` — "Every story lives in `/stories/SXX-short-name.md`"

  **API/Type References**: None

  **Test References**: None

  **External References**: None

  **WHY Each Reference Matters**:
  - AGENTS.md line 61 confirms the canonical path is `stories/` — the template must be there per project contract

  **Acceptance Criteria**:

  - [ ] `stories/` directory exists at project root
  - [ ] `stories/_template.md` exists and contains the original template content (50 lines, 7 sections)
  - [ ] `_template.md` no longer exists at project root

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Directory and template in correct location
    Tool: Bash
    Preconditions: Project root contains _template.md and AGENTS.md only
    Steps:
      1. Run: test -d stories && echo "DIR_EXISTS" || echo "DIR_MISSING"
      2. Run: test -f stories/_template.md && echo "TEMPLATE_EXISTS" || echo "TEMPLATE_MISSING"
      3. Run: test -f _template.md && echo "ROOT_STILL_EXISTS" || echo "ROOT_REMOVED"
      4. Run: wc -l stories/_template.md
    Expected Result: DIR_EXISTS, TEMPLATE_EXISTS, ROOT_REMOVED, "50 stories/_template.md"
    Failure Indicators: Any of the three checks fails, or line count differs from 50
    Evidence: .sisyphus/evidence/task-1-setup-verification.txt

  Scenario: Template content integrity
    Tool: Bash
    Preconditions: Template has been moved
    Steps:
      1. Run: head -1 stories/_template.md
      2. Run: grep -c "^## " stories/_template.md
    Expected Result: First line is "---", section count is 6 (User story, Acceptance criteria, Technical notes, Tests required, Out of scope, Status history)
    Failure Indicators: First line is not "---" or section count differs
    Evidence: .sisyphus/evidence/task-1-template-integrity.txt
  ```

  **Evidence to Capture:**
  - [ ] task-1-setup-verification.txt
  - [ ] task-1-template-integrity.txt

  **Commit**: YES
  - Message: `chore(stories): create stories directory and move template`
  - Files: `stories/_template.md`
  - Pre-commit: `test -f stories/_template.md && ! test -f _template.md`

- [x] 2. Write Strava domain stories — S01, S02, S03

  **What to do**:
  Write 3 story files covering Strava integration. These stories share OAuth tokens, webhook handling, and activity data ingestion — they must use consistent terminology for tokens, activities table, and API interactions.

  **S01 — Strava OAuth2 login and token storage**:
  - User story: Runner wants to connect Strava account so activities sync automatically
  - Key acceptance criteria: OAuth2 authorization code flow, encrypted token storage in SQLite, token refresh on expiry, revocation handling, secure redirect URI
  - Technical notes: Lives in `internal/strava/`, tokens encrypted at rest in SQLite (via `internal/storage/`), never store plaintext tokens (AGENTS.md constraint)
  - Tests: Unit (encrypt/decrypt round-trip, token refresh logic, expired token detection), Integration (full OAuth flow with mock Strava server), Edge cases (revoked token, network failure during auth, concurrent refresh requests)
  - Out of scope: Strava API data fetching (S02/S03), UI for login button (S16), other OAuth providers

  **S02 — Strava webhook receiver**:
  - User story: System receives push notifications when new activities are recorded so sync happens automatically
  - Key acceptance criteria: Webhook endpoint responds within 2 seconds (AGENTS.md constraint), subscription validation (GET challenge), async activity stream fetch after acknowledgment, deduplication by activity ID before processing
  - Technical notes: Lives in `internal/strava/`, webhook handler must be fast (respond then process), activity stream fetch is a separate goroutine, dedup check against SQLite before insert (AGENTS.md constraint). Depends on S01 for valid OAuth tokens.
  - Tests: Unit (webhook signature validation, dedup logic, challenge response), Integration (webhook → async fetch pipeline with mock Strava), Edge cases (duplicate webhook events, malformed payload, rapid successive webhooks, Strava rate limiting 429 responses)
  - Out of scope: Activity data parsing (S03), webhook subscription creation (manual setup step), UI notifications

  **S03 — Activity stream ingestion (HR, pace, cadence per second → SQLite)**:
  - User story: Runner's detailed activity data is stored locally so the context engine can analyze training patterns
  - Key acceptance criteria: Fetches HR/pace/cadence streams from Strava API, stores per-second data points in SQLite, handles activities with missing streams gracefully, maps to consistent schema usable by both Strava sync and FIT import (S17)
  - Technical notes: Lives in `internal/strava/` for fetch, `internal/storage/` for persistence. Activity table and stream data table schemas must be defined clearly — S17 (FIT import) maps to the same schema. Depends on S01 (tokens) and S02 (webhook triggers fetch).
  - Tests: Unit (stream parsing, schema mapping, missing field handling), Integration (fetch → store pipeline with mock API), Edge cases (activity with no HR data, extremely long activity 10+ hours, zero-length activity, non-running activity types)
  - Out of scope: Activity analysis/statistics, dashboard display (S15), FIT file parsing (S17)

  **Cross-story consistency rules for this task**:
  - Use consistent table names: `oauth_tokens`, `activities`, `activity_streams`
  - Use consistent column terminology: `activity_id`, `strava_id`, `access_token`, `refresh_token`, `token_expires_at`
  - All three stories reference `internal/strava/` as primary package
  - Token encryption referenced in S01; S02 and S03 say "uses tokens from S01" without re-specifying encryption

  **Must NOT do**:
  - Do not write Go code in stories
  - Do not specify encryption algorithm (leave as implementation detail)
  - Do not exceed 8 acceptance criteria per story
  - Do not exceed 15 lines in technical notes per story
  - Do not add sections beyond the 7 in the template
  - Do not include subjective criteria

  **Recommended Agent Profile**:
  - **Category**: `writing`
    - Reason: Technical writing task requiring domain knowledge of Strava API, OAuth2, webhooks, and SQLite storage patterns.
  - **Skills**: []
    - No special skills needed — the agent needs writing ability and domain knowledge, both intrinsic.

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 4, 5, 6)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `stories/_template.md` — Exact template to follow for all 3 files (after Task 1 moves it)
  - `AGENTS.md:119-125` — Strava sync rules (encrypted tokens, <2s webhook, async stream fetch, dedup)
  - `AGENTS.md:37-54` — Repository structure showing `internal/strava/` and `internal/storage/`

  **API/Type References**:
  - `AGENTS.md:109-113` — LLM interface (not directly used but shows the interface pattern for consistency)

  **Test References**:
  - `AGENTS.md:82-89` — Testing rules (unit next to package, integration in /tests/integration/, go test ./...)

  **External References**:
  - Strava API docs: https://developers.strava.com/docs/reference/ — OAuth2 flow, webhook subscription, activity streams endpoints
  - Strava webhook docs: https://developers.strava.com/docs/webhooks/ — Challenge response, event payload format

  **WHY Each Reference Matters**:
  - `AGENTS.md:119-125` contains the NON-NEGOTIABLE Strava constraints that acceptance criteria must encode
  - `stories/_template.md` is the format contract — deviating means the story is invalid
  - Strava API docs provide the real endpoint names and data shapes for accurate technical notes

  **Acceptance Criteria**:

  - [ ] `stories/S01-strava-oauth.md` exists with valid frontmatter (`id: S01`, `status: draft`, `created: 2026-03-16`)
  - [ ] `stories/S02-activity-sync.md` exists with valid frontmatter (`id: S02`, `status: draft`, `created: 2026-03-16`)
  - [ ] `stories/S03-activity-streams.md` exists with valid frontmatter (`id: S03`, `status: draft`, `created: 2026-03-16`)
  - [ ] All 3 files have all 7 template sections
  - [ ] Acceptance criteria in each story are specific and testable
  - [ ] Token table referenced as `oauth_tokens` consistently across all 3 stories
  - [ ] Activity table referenced as `activities` consistently
  - [ ] S02 explicitly mentions <2s webhook response requirement
  - [ ] S02 explicitly mentions deduplication by activity ID

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: All 3 Strava story files exist with valid structure
    Tool: Bash
    Preconditions: Task 1 completed (stories/ directory exists)
    Steps:
      1. Run: for f in stories/S01-strava-oauth.md stories/S02-activity-sync.md stories/S03-activity-streams.md; do test -f "$f" && echo "EXISTS: $f" || echo "MISSING: $f"; done
      2. Run: for f in stories/S0{1,2,3}-*.md; do grep -q "status: draft" "$f" && echo "DRAFT: $f" || echo "NOT_DRAFT: $f"; done
      3. Run: for f in stories/S0{1,2,3}-*.md; do count=$(grep -c "^## " "$f"); echo "$f: $count sections"; done
    Expected Result: All 3 files exist, all have draft status, all have 6 section headers
    Failure Indicators: Any file missing, wrong status, or section count ≠ 6
    Evidence: .sisyphus/evidence/task-2-strava-stories-structure.txt

  Scenario: Cross-story consistency — shared terminology
    Tool: Bash
    Preconditions: All 3 files written
    Steps:
      1. Run: grep -l "oauth_tokens" stories/S01-strava-oauth.md
      2. Run: grep -l "activities" stories/S02-activity-sync.md stories/S03-activity-streams.md
      3. Run: grep -c "2 seconds\|<2s\|two seconds" stories/S02-activity-sync.md
    Expected Result: S01 mentions oauth_tokens, S02+S03 mention activities table, S02 mentions 2-second constraint at least once
    Failure Indicators: Any grep returns no match
    Evidence: .sisyphus/evidence/task-2-strava-consistency.txt

  Scenario: No forbidden patterns in Strava stories
    Tool: Bash
    Preconditions: All 3 files written
    Steps:
      1. Run: grep -n "```go" stories/S0{1,2,3}-*.md || echo "NO_GO_CODE"
      2. Run: grep -in "looks good\|feels fast\|is responsive\|nice\|intuitive" stories/S0{1,2,3}-*.md || echo "NO_SUBJECTIVE"
      3. Run: for f in stories/S0{1,2,3}-*.md; do ac=$(grep -c "^\- \[ \]" "$f"); if [ "$ac" -gt 8 ]; then echo "TOO_MANY: $f ($ac)"; fi; done; echo "CHECK_DONE"
    Expected Result: NO_GO_CODE, NO_SUBJECTIVE, CHECK_DONE with no TOO_MANY warnings
    Failure Indicators: Go code blocks found, subjective language found, or >8 acceptance criteria
    Evidence: .sisyphus/evidence/task-2-strava-forbidden-patterns.txt
  ```

  **Evidence to Capture:**
  - [ ] task-2-strava-stories-structure.txt
  - [ ] task-2-strava-consistency.txt
  - [ ] task-2-strava-forbidden-patterns.txt

  **Commit**: NO (groups with Wave 2 commit after all stories written)

- [x] 3. Write Athlete Profile + Context Engine stories — S04, S05, S06, S07, S08

  **What to do**:
  Write 5 story files covering the athlete profile setup and the context engine. These are the most architecturally sensitive stories — AGENTS.md explicitly calls the context engine "the most sensitive part of the codebase." All 5 must use consistent terminology for context blocks, compression, token budgets, and pinned insights.

  **S04 — Athlete profile setup (age, max HR, threshold pace, goals, injury history)**:
  - User story: Runner wants to enter their profile data so the coaching AI has accurate context
  - Key acceptance criteria: Store profile fields in SQLite (age, max HR, threshold pace, weekly mileage target, race goals, injury history), validate input ranges, update existing profile, profile accessible to context engine
  - Technical notes: Lives in `internal/context/` or `internal/storage/`. Profile is a structured record, not free-text. Fields are typed (age: int, max HR: int, threshold pace: duration, goals: text, injuries: text with dates). No dependency on Strava — this is manual input.
  - Tests: Unit (field validation, CRUD operations), Integration (save → retrieve round-trip), Edge cases (empty profile, partial update, invalid values like negative age)
  - Out of scope: Automatic profile detection from activities, UI form (S16), LLM-suggested profile changes

  **S05 — Context engine: profile block assembly**:
  - User story: System assembles the athlete's profile into a structured text block so LLM conversations have accurate runner context
  - Key acceptance criteria: Reads profile from storage, formats into structured text block, output is deterministic (same input → same output), block includes all profile fields with labels
  - Technical notes: Lives in `internal/context/`. This is one block of the assembled context (S08 combines all blocks). Output format should be human-readable but structured. Depends on S04 for profile data.
  - Tests: Unit (formatting with full profile, formatting with partial profile), Integration (storage → block assembly), Edge cases (empty profile, missing optional fields, special characters in text fields)
  - Out of scope: Training summary blocks (S06), pinned insights (S07), full context assembly (S08), token counting

  **S06 — Context engine: rolling training summary (last 4 weeks, auto-compressed)**:
  - User story: System maintains a rolling summary of recent training so coaching advice reflects current fitness
  - Key acceptance criteria: Summarizes last 4 weeks of activities, older weeks compressed more aggressively than recent ones (AGENTS.md: "older training summaries must be compressed before recent ones"), auto-updates when new activities arrive, output fits within allocated token budget portion
  - Technical notes: Lives in `internal/context/`. Compression means summarization (reducing detail level, not file compression). Week 1 (most recent) gets most detail, Week 4 gets least. Must define what "compressed" means concretely — e.g., Week 1: per-run summary, Week 4: weekly totals only. Depends on S03 (activity data in SQLite).
  - Tests: Unit (compression at each level, summary generation per week), Integration (activities → summary pipeline), Edge cases (fewer than 4 weeks of data, no activities, 100+ activities in one week)
  - Out of scope: Real-time streaming updates, historical summaries beyond 4 weeks, LLM-based summarization (compression is algorithmic)

  **S07 — Context engine: pinned insights (saved from chat, never compressed)**:
  - User story: Runner can save important coaching insights from chat so they persist across sessions and are never lost
  - Key acceptance criteria: Insights stored in SQLite with timestamp and source session, marked as "pinned" and excluded from compression, retrievable for context assembly, deletable by user
  - Technical notes: Lives in `internal/context/`. AGENTS.md constraint (critical): "Pinned insights from chat are never compressed or dropped." This is the highest-priority context block — if token budget is tight, other blocks shrink before pinned insights. Depends on S14 (save-from-chat UI action) for creation, but storage/retrieval is independent.
  - Tests: Unit (CRUD for insights, pin/unpin logic), Integration (save → retrieve → include in context), Edge cases (100+ pinned insights exceeding budget, duplicate insight text, empty insight)
  - Out of scope: Automatic insight detection from chat (S14 is manual save), insight categorization, insight search

  **S08 — Context engine: prompt template assembler + token budget manager**:
  - User story: System assembles all context blocks into a final prompt that fits within the LLM's token limit
  - Key acceptance criteria: Assembles blocks in priority order (pinned insights > profile > recent training), enforces configurable token budget, compresses/truncates lower-priority blocks when budget is tight, token counting is accurate enough to prevent overflows, output is a complete prompt ready for LLM
  - Technical notes: Lives in `internal/context/`. Assembly order: system prompt → pinned insights (S07, never cut) → profile block (S05) → training summary (S06, compressed first when budget tight). Token counting can use approximation (4 chars ≈ 1 token) or a proper tokenizer — story should specify "configurable counting method." AGENTS.md constraint: "assembled context must always fit within the configured token budget."
  - Tests: Unit (assembly order, budget enforcement, compression triggers), Integration (all blocks → final prompt within budget), Edge cases (budget smaller than pinned insights alone, zero activities, all blocks empty, very large profile)
  - Out of scope: LLM-specific tokenizers (use approximation), prompt engineering (template content), streaming context updates

  **Cross-story consistency rules for this task**:
  - Context blocks referred to consistently: "profile block" (S05), "training summary block" (S06), "pinned insights block" (S07)
  - Compression terminology: "compression" = reducing detail level (algorithmic summarization), not file/data compression
  - Priority order consistent: pinned insights > profile > training summary (S07 never compressed, S06 compressed first)
  - Token budget referenced as "configurable token budget" across S06, S07, S08
  - All five stories reference `internal/context/` as primary package
  - Table name for insights: `pinned_insights`
  - Table name for profile: `athlete_profile`

  **Must NOT do**:
  - Do not write Go code in stories
  - Do not specify a concrete token budget number (leave as "configurable with sensible default")
  - Do not exceed 8 acceptance criteria per story
  - Do not exceed 15 lines in technical notes per story
  - Do not add sections beyond the 7 in the template
  - Do not include subjective criteria
  - Do not contradict AGENTS.md context engine rules (lines 96-103)

  **Recommended Agent Profile**:
  - **Category**: `writing`
    - Reason: Technical writing requiring understanding of context engine architecture, token budgets, compression strategies, and prompt assembly patterns.
  - **Skills**: []
    - No special skills needed.

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 4, 5, 6)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `stories/_template.md` — Exact template to follow for all 5 files
  - `AGENTS.md:96-103` — Context engine special rules (most sensitive, token budget, compression order, pinned insights never dropped)
  - `AGENTS.md:37-54` — Repository structure showing `internal/context/`

  **API/Type References**:
  - `AGENTS.md:109-113` — LLM interface showing `[]Message` type that context feeds into

  **Test References**:
  - `AGENTS.md:82-89` — Testing rules

  **External References**: None needed — context engine is internal architecture

  **WHY Each Reference Matters**:
  - `AGENTS.md:96-103` is CRITICAL — contains the non-negotiable context engine rules that S05-S08 acceptance criteria must encode exactly
  - The compression order rule ("Older training summaries must be compressed before recent ones") must appear verbatim in S06
  - The pinned insights rule ("never compressed or dropped") must appear verbatim in S07

  **Acceptance Criteria**:

  - [ ] `stories/S04-athlete-profile.md` exists with valid frontmatter
  - [ ] `stories/S05-context-profile-block.md` exists with valid frontmatter
  - [ ] `stories/S06-context-training-summary.md` exists with valid frontmatter
  - [ ] `stories/S07-context-pinned-insights.md` exists with valid frontmatter
  - [ ] `stories/S08-context-prompt-assembler.md` exists with valid frontmatter
  - [ ] All 5 files have all 7 template sections with non-empty Out of Scope
  - [ ] S06 explicitly states compression order: older weeks compressed before recent
  - [ ] S07 explicitly states pinned insights are never compressed or dropped
  - [ ] S08 references priority order: pinned insights > profile > training summary

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: All 5 context engine story files exist with valid structure
    Tool: Bash
    Preconditions: Task 1 completed
    Steps:
      1. Run: for f in stories/S04-athlete-profile.md stories/S05-context-profile-block.md stories/S06-context-training-summary.md stories/S07-context-pinned-insights.md stories/S08-context-prompt-assembler.md; do test -f "$f" && echo "EXISTS: $f" || echo "MISSING: $f"; done
      2. Run: for f in stories/S0{4,5,6,7,8}-*.md; do grep -q "status: draft" "$f" && echo "DRAFT: $f" || echo "NOT_DRAFT: $f"; done
    Expected Result: All 5 files exist, all have draft status
    Failure Indicators: Any file missing or wrong status
    Evidence: .sisyphus/evidence/task-3-context-stories-structure.txt

  Scenario: AGENTS.md constraints encoded in stories
    Tool: Bash
    Preconditions: All 5 files written
    Steps:
      1. Run: grep -c "compress.*older\|older.*compress\|compressed before recent" stories/S06-context-training-summary.md
      2. Run: grep -c "never compressed\|never dropped\|never.*compress" stories/S07-context-pinned-insights.md
      3. Run: grep -c "token budget\|token limit" stories/S08-context-prompt-assembler.md
    Expected Result: Each grep returns ≥1 match
    Failure Indicators: Any grep returns 0
    Evidence: .sisyphus/evidence/task-3-context-constraints.txt

  Scenario: No forbidden patterns in context engine stories
    Tool: Bash
    Preconditions: All 5 files written
    Steps:
      1. Run: grep -n "```go" stories/S0{4,5,6,7,8}-*.md || echo "NO_GO_CODE"
      2. Run: grep -in "looks good\|feels fast\|is responsive" stories/S0{4,5,6,7,8}-*.md || echo "NO_SUBJECTIVE"
    Expected Result: NO_GO_CODE, NO_SUBJECTIVE
    Failure Indicators: Go code blocks or subjective language found
    Evidence: .sisyphus/evidence/task-3-context-forbidden-patterns.txt
  ```

  **Evidence to Capture:**
  - [ ] task-3-context-stories-structure.txt
  - [ ] task-3-context-constraints.txt
  - [ ] task-3-context-forbidden-patterns.txt

  **Commit**: NO (groups with Wave 2 commit)

- [x] 4. Write LLM Router stories — S09, S10, S11

  **What to do**:
  Write 3 story files covering the LLM router interface and all three backend implementations. All three must implement the exact same interface from AGENTS.md. Cross-story consistency is critical — the interface contract, error handling patterns, and message format must be identical.

  **S09 — LLM router interface + Claude implementation**:
  - User story: System can route chat messages to Claude API so runners get AI coaching responses
  - Key acceptance criteria: Define the `LLM` interface (`Chat` + `Name` methods per AGENTS.md), define the `Message` struct (Role + Content fields), implement Claude backend, API key stored securely (not plaintext), handle API errors gracefully, support configurable model selection
  - Technical notes: Lives in `internal/llm/`. The interface is defined in AGENTS.md lines 110-113 and must be followed exactly. The `Message` type needs to be defined here — minimum fields: `Role string` (system/user/assistant) and `Content string`. S09 defines the contract that S10 and S11 must follow. Claude API uses the Anthropic Messages API.
  - Tests: Unit (message formatting, error wrapping, Name() returns "claude"), Integration (full chat round-trip with mock Claude API), Edge cases (empty message list, API key missing, rate limiting, malformed response, context length exceeded)
  - Out of scope: Streaming responses, tool/function calling, multi-modal input, Claude-specific prompt optimization

  **S10 — LLM router: OpenAI implementation**:
  - User story: System can route chat messages to OpenAI API as an alternative coaching backend
  - Key acceptance criteria: Implements same `LLM` interface from S09, maps `Message` struct to OpenAI's chat format, API key stored securely, handles API errors, supports model selection (GPT-4, etc.)
  - Technical notes: Lives in `internal/llm/`. Must implement the exact interface defined in S09. OpenAI's chat completion API uses a compatible message format (role + content). Depends on S09 for interface definition.
  - Tests: Unit (message mapping to OpenAI format, error wrapping, Name() returns "openai"), Integration (round-trip with mock OpenAI API), Edge cases (empty messages, API key missing, rate limiting, token limit exceeded)
  - Out of scope: Streaming, function calling, embeddings, image generation, fine-tuning

  **S11 — LLM router: local LLM implementation (Ollama-compatible endpoint)**:
  - User story: Runner can use a local LLM for privacy or offline coaching without sending data to cloud APIs
  - Key acceptance criteria: Implements same `LLM` interface from S09, connects to Ollama-compatible HTTP endpoint, configurable endpoint URL (default localhost:11434), handles connection failures gracefully, supports model selection
  - Technical notes: Lives in `internal/llm/`. Ollama exposes an OpenAI-compatible API at `/api/chat`. No API key needed for local — but endpoint URL must be configurable. Should handle "Ollama not running" error case explicitly. Depends on S09 for interface definition.
  - Tests: Unit (message formatting for Ollama API, endpoint URL construction, Name() returns "local"), Integration (round-trip with mock Ollama endpoint), Edge cases (Ollama not running, model not downloaded, slow response, very large response)
  - Out of scope: Ollama installation/setup, model management (pull/delete), GPU configuration, embedding API

  **Cross-story consistency rules for this task**:
  - Interface quoted identically in all 3 stories: `Chat(ctx context.Context, messages []Message) (string, error)` and `Name() string`
  - `Message` struct defined in S09 and referenced (not redefined) in S10, S11
  - Error handling pattern consistent: wrap API-specific errors in a common error type
  - All three reference `internal/llm/` as package
  - API key storage: S09 and S10 say "stored securely" (details in S16 settings story); S11 says "no API key required"
  - Name() returns: S09→"claude", S10→"openai", S11→"local"

  **Must NOT do**:
  - Do not write Go code in stories (exception: quoting the interface from AGENTS.md is permitted)
  - Do not exceed 8 acceptance criteria per story
  - Do not exceed 15 lines in technical notes per story
  - Do not add sections beyond the 7 in the template
  - Do not redefine the LLM interface — quote AGENTS.md exactly
  - Do not specify streaming support (explicitly out of scope)

  **Recommended Agent Profile**:
  - **Category**: `writing`
    - Reason: Technical writing requiring understanding of LLM API patterns (Anthropic, OpenAI, Ollama), Go interfaces, and API client design.
  - **Skills**: []
    - No special skills needed.

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3, 5, 6)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `stories/_template.md` — Exact template to follow
  - `AGENTS.md:104-116` — LLM router interface section with exact Go interface definition
  - `AGENTS.md:37-54` — Repository structure showing `internal/llm/`

  **API/Type References**:
  - `AGENTS.md:110-113` — The exact LLM interface: `Chat(ctx context.Context, messages []Message) (string, error)` and `Name() string`

  **Test References**:
  - `AGENTS.md:82-89` — Testing rules

  **External References**:
  - Anthropic Messages API: https://docs.anthropic.com/en/api/messages — Claude's chat endpoint format
  - OpenAI Chat Completions: https://platform.openai.com/docs/api-reference/chat — OpenAI's chat format
  - Ollama API: https://github.com/ollama/ollama/blob/main/docs/api.md — Local LLM endpoint format

  **WHY Each Reference Matters**:
  - `AGENTS.md:110-113` is the SINGLE SOURCE OF TRUTH for the interface — all 3 stories must quote it identically
  - External API docs provide the real request/response shapes that technical notes should reference for each backend

  **Acceptance Criteria**:

  - [ ] `stories/S09-llm-router-claude.md` exists with valid frontmatter
  - [ ] `stories/S10-llm-router-openai.md` exists with valid frontmatter
  - [ ] `stories/S11-llm-router-local.md` exists with valid frontmatter
  - [ ] All 3 files have all 7 template sections with non-empty Out of Scope
  - [ ] All 3 stories reference the exact LLM interface from AGENTS.md
  - [ ] S09 defines the Message struct requirements (Role + Content)
  - [ ] S10 and S11 reference S09 for interface definition, not redefining it
  - [ ] Name() return values are specified: "claude", "openai", "local"

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: All 3 LLM story files exist with valid structure
    Tool: Bash
    Preconditions: Task 1 completed
    Steps:
      1. Run: for f in stories/S09-llm-router-claude.md stories/S10-llm-router-openai.md stories/S11-llm-router-local.md; do test -f "$f" && echo "EXISTS: $f" || echo "MISSING: $f"; done
      2. Run: for f in stories/S{09,10,11}-*.md; do grep -q "status: draft" "$f" && echo "DRAFT: $f" || echo "NOT_DRAFT: $f"; done
    Expected Result: All 3 files exist, all have draft status
    Failure Indicators: Any file missing or wrong status
    Evidence: .sisyphus/evidence/task-4-llm-stories-structure.txt

  Scenario: LLM interface consistency across all 3 stories
    Tool: Bash
    Preconditions: All 3 files written
    Steps:
      1. Run: grep -c "Chat(ctx\|Chat(" stories/S{09,10,11}-*.md
      2. Run: grep -c "Name()" stories/S{09,10,11}-*.md
      3. Run: grep -c "Message" stories/S09-llm-router-claude.md
    Expected Result: Each story mentions Chat and Name(), S09 mentions Message at least once
    Failure Indicators: Any story missing interface reference
    Evidence: .sisyphus/evidence/task-4-llm-interface-consistency.txt

  Scenario: No forbidden patterns in LLM stories
    Tool: Bash
    Preconditions: All 3 files written
    Steps:
      1. Run: grep -in "streaming\|stream response" stories/S{09,10,11}-*.md | grep -iv "out of scope" || echo "NO_STREAMING_IN_SCOPE"
      2. Run: grep -in "looks good\|feels fast" stories/S{09,10,11}-*.md || echo "NO_SUBJECTIVE"
    Expected Result: NO_STREAMING_IN_SCOPE (streaming only mentioned in Out of Scope), NO_SUBJECTIVE
    Failure Indicators: Streaming mentioned outside Out of Scope, or subjective criteria found
    Evidence: .sisyphus/evidence/task-4-llm-forbidden-patterns.txt
  ```

  **Evidence to Capture:**
  - [ ] task-4-llm-stories-structure.txt
  - [ ] task-4-llm-interface-consistency.txt
  - [ ] task-4-llm-forbidden-patterns.txt

  **Commit**: NO (groups with Wave 2 commit)

- [x] 5. Write Chat + UI stories — S12, S13, S14, S15, S16

  **What to do**:
  Write 5 story files covering the chat interface, history persistence, insight saving, activity dashboard, and settings screen. These are Svelte frontend stories backed by Wails Go bindings. All must specify behavior (not appearance) and define which Go functions are exposed to the frontend.

  **S12 — Chat UI (Svelte): send message, display response, markdown rendering**:
  - User story: Runner wants to chat with an AI coach through a conversational interface
  - Key acceptance criteria: Text input field for sending messages, messages display in chronological order, AI responses rendered as markdown, loading indicator during LLM response, Enter key sends message, empty messages prevented
  - Technical notes: Lives in `frontend/`. Wails v2 binds Go functions that the Svelte frontend calls. This story needs the Go binding for `SendMessage(message string) (string, error)` exposed in `app.go`. Depends on S09 (LLM router) for backend. Plain Svelte components (not SvelteKit — Wails uses webview).
  - Tests: Unit (message validation, markdown rendering), Integration (send → receive round-trip via Wails binding), Edge cases (very long message, rapid successive sends, markdown edge cases like code blocks/tables, empty state on first load)
  - Out of scope: Voice input, file attachments, message editing/deletion, real-time typing indicators, chat themes/customization

  **S13 — Chat history persistence (SQLite, per-session)**:
  - User story: Runner's chat conversations are saved so they can review previous coaching sessions
  - Key acceptance criteria: Messages persisted to SQLite per session, sessions identified by unique ID + timestamp, previous sessions loadable from history, new session created on app launch or explicit action
  - Technical notes: Lives in `internal/storage/` for persistence, binding in `app.go`. Table: `chat_sessions` (id, created_at) and `chat_messages` (id, session_id, role, content, created_at). Depends on S12 for chat UI integration.
  - Tests: Unit (message CRUD, session creation, session listing), Integration (chat → save → reload round-trip), Edge cases (session with 1000+ messages, concurrent writes, empty session, corrupted message content)
  - Out of scope: Session search/filtering, session export, message reactions, session sharing

  **S14 — Save insight from chat to pinned context**:
  - User story: Runner can save a specific coaching insight from chat so it becomes permanent context for future conversations
  - Key acceptance criteria: User can select a chat message and save it as a pinned insight, saved insight appears in pinned insights store (S07), confirmation feedback shown to user, duplicate detection (same text already pinned)
  - Technical notes: Lives in `frontend/` for UI action, `internal/context/` for storage. This bridges chat (S12/S13) with context engine (S07). The Go binding needs `SaveInsight(messageContent string) error` in `app.go`. Depends on S07 (pinned insights storage) and S12 (chat UI).
  - Tests: Unit (insight saving, duplicate detection), Integration (select message → save → verify in context), Edge cases (save empty message, save very long message, save same insight twice, save while offline)
  - Out of scope: Automatic insight detection, insight editing after save, batch insight saving, insight categorization/tagging

  **S15 — Activity dashboard (recent activities list, basic metrics display)**:
  - User story: Runner wants to see their recent activities and basic metrics at a glance
  - Key acceptance criteria: Display list of recent activities (last 20), show per-activity metrics (date, distance, duration, avg pace, avg HR), activities sorted by date descending, empty state when no activities exist, activity data sourced from SQLite
  - Technical notes: Lives in `frontend/` for display, Go binding `GetRecentActivities(limit int) ([]Activity, error)` in `app.go`. Reads from `activities` table (same schema as S03). No charting or detailed analysis — just a list with summary metrics. Depends on S03 (activity data).
  - Tests: Unit (activity list rendering, metric formatting), Integration (fetch → display pipeline), Edge cases (zero activities, activities with missing metrics, 1000+ activities pagination boundary, non-running activity types)
  - Out of scope: Activity detail view, charting/graphs, activity comparison, training load calculation, route maps

  **S16 — Settings screen (API keys, LLM selection, Strava re-auth)**:
  - User story: Runner wants to configure API keys, choose their preferred LLM, and manage Strava connection
  - Key acceptance criteria: Input fields for Claude and OpenAI API keys, API keys stored encrypted in SQLite (same encryption as OAuth tokens), dropdown/selector for active LLM backend (Claude/OpenAI/Local), Ollama endpoint URL configuration, button to re-authorize Strava, current connection status displayed
  - Technical notes: Lives in `frontend/` for UI, `internal/storage/` for encrypted key storage. API key encryption should use the same mechanism as OAuth token encryption (S01). Go bindings needed: `SaveSettings(settings Settings) error`, `GetSettings() (Settings, error)`, `ReauthorizeStrava() error`. Depends on S01 (Strava auth), S09-S11 (LLM backends).
  - Tests: Unit (settings CRUD, input validation, key masking in UI), Integration (save → reload settings, LLM switch), Edge cases (invalid API key format, empty fields, switching LLM while chat is active, Strava re-auth failure)
  - Out of scope: Theme/appearance settings, notification preferences, data export/import, account deletion, usage statistics

  **Cross-story consistency rules for this task**:
  - All UI stories specify behavior, not appearance (no colors, fonts, pixel sizes)
  - Wails Go bindings listed in technical notes of each story that needs them
  - SQLite table names consistent with other domains: `chat_sessions`, `chat_messages`, `activities` (same as S03), `settings`
  - All stories reference `frontend/` for Svelte and `app.go` for bindings
  - Error states specified: every story says what happens when backend calls fail
  - Empty states specified: every story says what shows when no data exists
  - Loading states specified: stories with async operations mention loading indicators

  **Must NOT do**:
  - Do not specify UI design (colors, fonts, layouts, pixel dimensions)
  - Do not write Svelte component code in stories
  - Do not write Go code in stories
  - Do not exceed 8 acceptance criteria per story
  - Do not exceed 15 lines in technical notes per story
  - Do not add sections beyond the 7 in the template
  - Do not include subjective criteria ("looks clean", "feels snappy")

  **Recommended Agent Profile**:
  - **Category**: `writing`
    - Reason: Technical writing requiring understanding of Wails v2 Go-to-frontend binding patterns, Svelte component architecture, and desktop app UX patterns.
  - **Skills**: []
    - No special skills needed.

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3, 4, 6)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `stories/_template.md` — Exact template to follow
  - `AGENTS.md:37-54` — Repository structure showing `frontend/` and `app.go`
  - `AGENTS.md:119-125` — Strava rules relevant to S16 re-auth
  - `AGENTS.md:109-113` — LLM interface relevant to S16 backend selection

  **API/Type References**: None beyond AGENTS.md

  **Test References**:
  - `AGENTS.md:82-89` — Testing rules

  **External References**:
  - Wails v2 docs: https://wails.io/docs/reference/runtime/intro — Wails Go↔JS binding pattern

  **WHY Each Reference Matters**:
  - `AGENTS.md:37-54` shows the canonical frontend structure — stories must reference `frontend/` and `app.go`, not invented paths
  - Wails v2 docs clarify how Go functions are exposed to Svelte (binding pattern)

  **Acceptance Criteria**:

  - [ ] `stories/S12-chat-ui.md` exists with valid frontmatter
  - [ ] `stories/S13-chat-history.md` exists with valid frontmatter
  - [ ] `stories/S14-save-insight.md` exists with valid frontmatter
  - [ ] `stories/S15-activity-dashboard.md` exists with valid frontmatter
  - [ ] `stories/S16-settings.md` exists with valid frontmatter
  - [ ] All 5 files have all 7 template sections with non-empty Out of Scope
  - [ ] No UI design specifications (colors, fonts, pixel sizes) in any story
  - [ ] All stories mention error/empty states
  - [ ] S16 mentions encrypted API key storage

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: All 5 UI story files exist with valid structure
    Tool: Bash
    Preconditions: Task 1 completed
    Steps:
      1. Run: for f in stories/S12-chat-ui.md stories/S13-chat-history.md stories/S14-save-insight.md stories/S15-activity-dashboard.md stories/S16-settings.md; do test -f "$f" && echo "EXISTS: $f" || echo "MISSING: $f"; done
      2. Run: for f in stories/S1{2,3,4,5,6}-*.md; do grep -q "status: draft" "$f" && echo "DRAFT: $f" || echo "NOT_DRAFT: $f"; done
    Expected Result: All 5 files exist, all have draft status
    Failure Indicators: Any file missing or wrong status
    Evidence: .sisyphus/evidence/task-5-ui-stories-structure.txt

  Scenario: No UI design specs in stories (behavior only)
    Tool: Bash
    Preconditions: All 5 files written
    Steps:
      1. Run: grep -in "px\|pixel\|#[0-9a-f]\{3,6\}\|font-size\|color:" stories/S1{2,3,4,5,6}-*.md || echo "NO_DESIGN_SPECS"
      2. Run: grep -in "looks good\|feels\|intuitive\|clean\|beautiful\|sleek" stories/S1{2,3,4,5,6}-*.md || echo "NO_SUBJECTIVE"
    Expected Result: NO_DESIGN_SPECS, NO_SUBJECTIVE
    Failure Indicators: Design specs or subjective language found
    Evidence: .sisyphus/evidence/task-5-ui-behavior-only.txt

  Scenario: Error and empty states mentioned
    Tool: Bash
    Preconditions: All 5 files written
    Steps:
      1. Run: for f in stories/S1{2,3,4,5,6}-*.md; do grep -ci "error\|fail\|empty" "$f"; done
    Expected Result: Each file returns ≥1 (mentions error handling or empty state at least once)
    Failure Indicators: Any file returns 0
    Evidence: .sisyphus/evidence/task-5-ui-error-states.txt
  ```

  **Evidence to Capture:**
  - [ ] task-5-ui-stories-structure.txt
  - [ ] task-5-ui-behavior-only.txt
  - [ ] task-5-ui-error-states.txt

  **Commit**: NO (groups with Wave 2 commit)

- [x] 6. Write FIT File Import story — S17

  **What to do**:
  Write 1 story file for optional FIT file import. This story must explicitly reference the same activity schema used in S03 (Strava ingestion) so both data sources map to identical storage.

  **S17 — FIT file import (optional manual import, maps to same schema as Strava sync)**:
  - User story: Runner can manually import activities from FIT files (e.g., from Garmin) so activities from non-Strava sources appear in the system
  - Key acceptance criteria: Parse FIT file format to extract HR/pace/cadence streams, map to same SQLite schema as Strava-synced activities (S03), deduplication check (don't reimport same activity), file picker UI for selecting FIT files, import status feedback (success/failure/duplicate)
  - Technical notes: Lives in `internal/fit/` for parsing, `internal/storage/` for persistence. CRITICAL: must use the exact same `activities` and `activity_streams` tables and schema as S03. The FIT SDK/parser is the main implementation challenge — consider using an existing Go FIT library. Deduplication: hash activity start time + duration + distance to detect reimports. This is an optional feature — the app works without it.
  - Tests: Unit (FIT parsing, schema mapping, dedup hash), Integration (file → parse → store pipeline), Edge cases (corrupted FIT file, FIT file with missing streams, very large file, non-running activity in FIT, file from unknown device)
  - Out of scope: TCX/GPX import, automatic file detection, activity editing after import, batch import UI

  **Must NOT do**:
  - Do not write Go code in stories
  - Do not exceed 8 acceptance criteria
  - Do not exceed 15 lines in technical notes
  - Do not add sections beyond the 7 in the template
  - Do not include subjective criteria
  - Do not use different table/schema names than S03

  **Recommended Agent Profile**:
  - **Category**: `writing`
    - Reason: Technical writing for a single focused story requiring knowledge of FIT file format and schema mapping.
  - **Skills**: []
    - No special skills needed.

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3, 4, 5)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `stories/_template.md` — Exact template to follow
  - `AGENTS.md:37-54` — Repository structure showing `internal/fit/` with "(optional import)" note

  **API/Type References**: None beyond AGENTS.md

  **Test References**:
  - `AGENTS.md:82-89` — Testing rules

  **External References**:
  - FIT SDK: https://developer.garmin.com/fit/overview/ — Garmin FIT file format specification
  - Go FIT libraries: github.com/tormoder/fit — Popular Go FIT parsing library

  **WHY Each Reference Matters**:
  - `AGENTS.md:37-54` confirms `internal/fit/` as the canonical package location
  - FIT SDK docs provide the real data structure names for accurate technical notes
  - The Go FIT library reference helps the implementing agent avoid writing a parser from scratch

  **Acceptance Criteria**:

  - [ ] `stories/S17-fit-file-import.md` exists with valid frontmatter (`id: S17`, `status: draft`, `created: 2026-03-16`)
  - [ ] File has all 7 template sections with non-empty Out of Scope
  - [ ] Story references same `activities` and `activity_streams` tables as S03
  - [ ] Story mentions deduplication strategy
  - [ ] Story explicitly notes this is an optional feature

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: FIT import story file exists with valid structure
    Tool: Bash
    Preconditions: Task 1 completed
    Steps:
      1. Run: test -f stories/S17-fit-file-import.md && echo "EXISTS" || echo "MISSING"
      2. Run: grep -q "status: draft" stories/S17-fit-file-import.md && echo "DRAFT" || echo "NOT_DRAFT"
      3. Run: grep -c "^## " stories/S17-fit-file-import.md
    Expected Result: EXISTS, DRAFT, 6 sections
    Failure Indicators: File missing, wrong status, or wrong section count
    Evidence: .sisyphus/evidence/task-6-fit-story-structure.txt

  Scenario: Schema consistency with S03
    Tool: Bash
    Preconditions: S17 file written
    Steps:
      1. Run: grep -c "activities\|activity_streams" stories/S17-fit-file-import.md
      2. Run: grep -ci "same schema\|same table\|identical schema\|S03" stories/S17-fit-file-import.md
    Expected Result: Both return ≥1
    Failure Indicators: No mention of shared schema or S03 reference
    Evidence: .sisyphus/evidence/task-6-fit-schema-consistency.txt
  ```

  **Evidence to Capture:**
  - [ ] task-6-fit-story-structure.txt
  - [ ] task-6-fit-schema-consistency.txt

  **Commit**: NO (groups with Wave 2 commit)

- [x] 7. QA Verification + Summary Table

  **What to do**:
  Run comprehensive verification across all 17 story files. Check structural compliance, cross-domain consistency, forbidden patterns, and AGENTS.md constraint encoding. Then generate and print the summary table.

  Verification steps:
  1. **Structural check**: All 17 files exist in `stories/`, all have valid frontmatter, all have 7 sections, all have `status: draft`
  2. **Content quality**: No empty Out of Scope sections, no subjective acceptance criteria, no Go code blocks (except LLM interface quote), acceptance criteria counts ≤8 per story
  3. **Cross-domain consistency**:
     - Strava stories (S01-S03): consistent table names (`oauth_tokens`, `activities`, `activity_streams`)
     - Context engine (S05-S08): consistent terminology (compression, token budget, block names)
     - LLM stories (S09-S11): identical interface reference, consistent naming
     - UI stories (S12-S16): error/empty states mentioned, no design specs
     - FIT import (S17): same schema as S03
  4. **AGENTS.md constraint encoding**: <2s webhook in S02, encrypted tokens in S01, dedup in S02, compression order in S06, pinned insights never dropped in S07, token budget in S08, LLM interface exact match in S09-S11
  5. **Summary table**: Generate markdown table with columns: ID, Title, Status, Key dependency

  **Must NOT do**:
  - Do not modify any story files (read-only verification)
  - Do not create additional stories
  - Do not change any file statuses

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Multi-step verification requiring reading all 17 files, running pattern checks, and generating a structured report. Not a simple task.
  - **Skills**: []
    - No special skills needed.

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (sole task)
  - **Blocks**: F1-F4
  - **Blocked By**: Tasks 2, 3, 4, 5, 6

  **References**:

  **Pattern References**:
  - `stories/_template.md` — Template structure to verify against
  - `AGENTS.md:96-103` — Context engine rules to verify encoding
  - `AGENTS.md:109-116` — LLM interface to verify exact match
  - `AGENTS.md:119-125` — Strava rules to verify encoding
  - `AGENTS.md:82-89` — Testing rules to verify test section completeness

  **WHY Each Reference Matters**:
  - Each AGENTS.md section contains constraints that MUST appear in the corresponding stories — this task verifies that

  **Acceptance Criteria**:

  - [ ] All 17 story files pass structural validation
  - [ ] All 17 files pass content quality checks
  - [ ] Cross-domain consistency verified with zero issues
  - [ ] All AGENTS.md constraints found in corresponding stories
  - [ ] Summary table generated with 17 rows
  - [ ] If any issues found: report them with file:line references for the implementing agent to fix

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Complete structural verification of all 17 stories
    Tool: Bash
    Preconditions: All writing tasks (2-6) completed
    Steps:
      1. Run: ls stories/S{01..17}-*.md 2>/dev/null | wc -l
      2. Run: for f in stories/S*.md; do grep -q "status: draft" "$f" && echo "OK" || echo "FAIL: $f"; done
      3. Run: for f in stories/S*.md; do for s in "User story" "Acceptance criteria" "Technical notes" "Tests required" "Out of scope" "Status history"; do grep -q "## $s" "$f" || echo "MISSING '$s' in $f"; done; done
    Expected Result: 17 files, all OK status, no missing sections
    Failure Indicators: Count ≠ 17, any FAIL status, any MISSING section
    Evidence: .sisyphus/evidence/task-7-structural-verification.txt

  Scenario: Forbidden patterns scan across all stories
    Tool: Bash
    Preconditions: All stories exist
    Steps:
      1. Run: grep -rn "```go" stories/S*.md | grep -v "S09\|S10\|S11" || echo "NO_UNEXPECTED_GO_CODE"
      2. Run: grep -rin "looks good\|feels fast\|is responsive\|intuitive\|beautiful" stories/S*.md || echo "NO_SUBJECTIVE"
      3. Run: for f in stories/S*.md; do ac=$(grep -c "^\- \[ \]" "$f"); if [ "$ac" -gt 8 ]; then echo "TOO_MANY: $f ($ac)"; fi; done; echo "DONE"
    Expected Result: NO_UNEXPECTED_GO_CODE, NO_SUBJECTIVE, DONE with no TOO_MANY
    Failure Indicators: Unexpected Go code, subjective language, or >8 acceptance criteria
    Evidence: .sisyphus/evidence/task-7-forbidden-patterns.txt

  Scenario: Summary table generation
    Tool: Bash
    Preconditions: All verification passed
    Steps:
      1. Generate a markdown table with columns: ID | Title | Status | Key dependency
      2. Print the table to stdout
      3. Verify table has exactly 17 data rows (header + separator + 17 rows)
    Expected Result: Table with 17 rows, all showing "draft" status
    Failure Indicators: Wrong row count or missing stories
    Evidence: .sisyphus/evidence/task-7-summary-table.txt
  ```

  **Evidence to Capture:**
  - [ ] task-7-structural-verification.txt
  - [ ] task-7-forbidden-patterns.txt
  - [ ] task-7-summary-table.txt

  **Commit**: YES
  - Message: `docs(stories): write S01-S17 user story files`
  - Files: `stories/S01-*.md` through `stories/S17-*.md`
  - Pre-commit: `ls stories/S{01..17}-*.md | wc -l` (expected: 17)

---

## Final Verification Wave

> 4 review agents run in PARALLEL. ALL must APPROVE. Rejection → fix → re-run.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read each story file, check section presence). For each "Must NOT Have": search stories for forbidden patterns (subjective criteria, Go code blocks, extra sections) — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Review all 17 story files for: vague acceptance criteria, missing test cases, empty Out of Scope, frontmatter errors, broken cross-references (e.g., referencing S18 which doesn't exist), inconsistent terminology within domains. Check AGENTS.md compliance: file paths match repository structure, LLM interface quoted correctly, Strava constraints respected.
  Output: `Stories [N/17 clean] | Frontmatter [N/17 valid] | Cross-refs [N valid/N broken] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high`
  Read every story file end-to-end. Verify: acceptance criteria are truly testable (could you write a Go test for each?), technical notes give enough context for an agent to implement without guessing, test cases cover unit + integration + edge cases, out of scope prevents likely creep. Check domain consistency: Strava stories use same table names, context engine stories use same compression terminology, LLM stories reference same interface.
  Output: `Stories [N/17 pass] | Testability [N/N criteria testable] | Consistency [N domains clean] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", verify actual output matches. Verify exactly 17 stories exist (no extra, no missing). Verify `_template.md` was moved to `stories/`. Verify no files created outside `stories/` directory. Check that summary table was generated with all 17 rows. Flag any unaccounted files or missing deliverables.
  Output: `Stories [17/17] | Template [MOVED/MISSING] | Summary [GENERATED/MISSING] | Extra files [CLEAN/N found] | VERDICT`

---

## Commit Strategy

- **Wave 1**: `chore(stories): create stories directory and move template` — stories/_template.md
- **Wave 2**: `docs(stories): write S01-S17 user story files` — stories/S01-*.md through stories/S17-*.md (single commit after all 5 writing tasks complete)

---

## Success Criteria

### Verification Commands
```bash
# All 17 stories exist
ls stories/S{01..17}-*.md 2>/dev/null | wc -l  # Expected: 17

# All have draft status
grep -c "status: draft" stories/S*.md | grep -v ":0$" | wc -l  # Expected: 17

# Template exists in stories/
test -f stories/_template.md && echo "OK" || echo "MISSING"  # Expected: OK

# No empty Out of Scope sections
for f in stories/S*.md; do
  awk '/^## Out of scope/{getline; getline; if ($0 ~ /^## |^---/) print FILENAME}' "$f"
done  # Expected: no output (all have content)

# All have required sections
for f in stories/S*.md; do
  for section in "User story" "Acceptance criteria" "Technical notes" "Tests required" "Out of scope" "Status history"; do
    grep -q "## $section" "$f" || echo "MISSING '$section' in $f"
  done
done  # Expected: no output
```

### Final Checklist
- [ ] All 17 story files present in `stories/`
- [ ] `_template.md` moved to `stories/_template.md`
- [ ] All frontmatter valid with status `draft`
- [ ] All sections present in every story
- [ ] No empty Out of Scope sections
- [ ] No subjective acceptance criteria
- [ ] Cross-story domain consistency verified
- [ ] Summary table generated
