---
id: S47
title: End-to-end test infrastructure with Playwright
status: in-progress
created: 2026-03-17
updated: 2026-03-17
---

# S47 — End-to-end test infrastructure with Playwright

## User story

As a **developer**,
I want **a Playwright-based e2e test suite that exercises the full UI as a real user would**
so that **regressions in UI interactions are caught automatically before release**.

## Problem

The project has zero frontend tests. The entire test suite is Go unit tests only.
A real user interacts entirely through the browser-based Svelte UI, not through Go functions directly.
E2e tests ensure that:
- Navigation between tabs works
- Forms load, accept input, and submit correctly
- Feedback messages appear after actions
- Onboarding wizard progresses step by step
- Chat UI renders messages and shows loading states
- Dashboard shows activities/stats correctly
- Context tab saves and deletes data
- Settings persist and show correct state

## Acceptance criteria

- [ ] Playwright is installed as a dev dependency in `frontend/`
- [ ] `playwright.config.ts` is at `frontend/playwright.config.ts`
- [ ] E2e tests live in `frontend/e2e/`
- [ ] Test runner mocks Wails Go bindings (via `window.go` mock in a global setup)
- [ ] Tests can be run with `npm run e2e` from the `frontend/` directory
- [ ] Coverage for: navigation, chat, dashboard, context tab, settings, onboarding

## Technical notes

Since this is a Wails desktop app, Playwright tests run against the Vite dev server
(not the compiled binary). The Wails JS bindings (`window.go.*`) must be mocked
globally so tests don't require a running Go backend.

Mock strategy:
- Create `frontend/e2e/mocks/wails.ts` that stubs all `window.go.main.App.*` calls
- Also mock `window.runtime` for dialog functions
- Inject mocks via `page.addInitScript()` in test setup

Test structure:
```
frontend/e2e/
  mocks/
    wails.ts        ← all backend mocks
  navigation.spec.ts
  chat.spec.ts
  dashboard.spec.ts
  context.spec.ts
  settings.spec.ts
  onboarding.spec.ts
  s39-custom-prompt.spec.ts
  s45-export-context.spec.ts
  s46-free-ai.spec.ts
```

## Tests required

- Navigation: clicking each sidebar item shows the correct panel
- Chat: empty state, typing message, send button enables/disables, loading indicator
- Dashboard: shows empty state when no activities; shows stats bar when activities present
- Context: profile form renders with correct fields; save shows feedback
- Settings: backend selector changes visible fields; save shows feedback
- Onboarding: 5-step flow, step progression, skip buttons work
- S39: custom prompt textarea visible in settings; persists on save
- S45: export button triggers dialog (mocked); success message shows
- S46: selecting "free" and saving settings works without error

## Out of scope

- Testing actual LLM responses (requires live API keys)
- Testing actual Strava OAuth flow (requires browser redirect)
- Visual regression / screenshot tests
- Mobile layout testing

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-17 | in-progress | Created alongside e2e implementation |
