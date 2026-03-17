---
id: S46
title: Fix free AI (Gemini) backend saving
status: draft
created: 2026-03-17
updated: 2026-03-17
---

# S46 — Fix free AI (Gemini) backend saving

## User story

As a **runner**,
I want **to save the "Free (Gemini Flash)" LLM backend option in settings**
so that **I can use the free tier without configuring API keys or local models**.

## Problem

When the user selects "Free (Gemini Flash)" as the active LLM backend in Settings or Onboarding and clicks Save, the settings fail to save properly. The issue stems from the free LLM backend implementation requiring an API key even though it should be optional for the free tier.

**Root cause**: The `NewFree` function in `internal/llm/free.go` (lines 30-46) requires an API key:

```go
func NewFree(config FreeConfig) (*Free, error) {
  // ... setup defaults ...
  if config.APIKey == "" {
    config.APIKey = os.Getenv("GEMINI_API_KEY")
  }
  if config.APIKey == "" {
    return nil, errors.New("free: API key is required (set GEMINI_API_KEY env var or build with key)")
  }
  // ...
}
```

The app initialization logic calls `NewFree` when `activeLlm == "free"` (see the LLM initialization in app.go, likely in `reloadLLMClient()` or similar), and this fails when no API key is present, preventing the settings from being saved.

Additionally, the error message in `internal/storage/settings.go` line 34 doesn't mention "free" as a valid option:

```go
return fmt.Errorf("active_llm must be one of claude, openai, local; got %q", s.ActiveLLM)
```

This error is outdated since "free" was added to the `validLLMs` map on line 26.

## Acceptance criteria

- [ ] User can select "Free (Gemini Flash)" in Settings backend selector
- [ ] User can save settings with "Free (Gemini Flash)" as the active backend
- [ ] No API key input is shown when "Free" backend is selected (current behavior is correct)
- [ ] Free backend works without requiring an API key from the user
- [ ] App initializes correctly with "free" as the active LLM backend
- [ ] Chat functionality works when using the free backend
- [ ] Update the error message in `settings.go` to include "free" as a valid option
- [ ] Onboarding wizard allows selecting and saving "Free" backend

## Technical notes

**Required changes**:

1. **Internal LLM validation** - `internal/storage/settings.go` line 34:
   Update error message to list all valid options including "free":
   ```go
   return fmt.Errorf("active_llm must be one of claude, openai, local, free; got %q", s.ActiveLLM)
   ```

2. **Free LLM initialization** - `internal/llm/free.go`:
   The free backend should either:
   - Provide a built-in free-tier API key (if available from Google's free tier)
   - OR handle the no-key case gracefully by using the public free endpoint
   - OR provide a clear error message guiding the user to set up an API key

   According to S34 story, the original intent was to use a built-in key or free endpoint. Check if Gemini provides a truly free unauthenticated endpoint, or if the app needs to embed a free-tier key at build time.

3. **App LLM initialization** - `app.go`:
   Locate where `NewFree` is called (likely in `initLLM()` or `reloadLLMClient()`) and ensure it handles the case where API key is not set by the user.

**Build-time key strategy** (from S34):
The story mentions using `go:embed` or `-ldflags -X` to inject the key at build time. If a free-tier key exists:
- Add build variable for `GEMINI_API_KEY`
- Inject at build time: `-ldflags "-X coachlm/internal/llm.geminiApiKey=YOUR_KEY"`
- Store in unexported variable, not in source

**Related files**:
- `internal/llm/free.go` - Modify `NewFree` to handle no-key case
- `internal/storage/settings.go` - Update error message on line 34
- `app.go` - Review LLM initialization logic
- `frontend/src/Settings.svelte` - Verify UI behavior (no changes needed)
- `frontend/src/Onboarding.svelte` - Verify UI behavior (no changes needed)

## Tests required

- Unit: Verify settings save succeeds with `activeLlm = "free"`
- Unit: Verify error message includes "free" as a valid option when validation fails
- Integration: Save settings with free backend → reload app → verify LLM initializes
- Manual test: Select "Free (Gemini Flash)" in Settings → Save → Settings persist after reload
- Manual test: Onboarding → Select "Free" → Complete wizard → Chat works
- Edge case: If no API key is available at all, provide clear error message guiding user

## Out of scope

- Adding streaming responses for free backend
- Usage tracking or rate limit UI
- Multiple free providers (only Gemini)
- Changing the free backend API/endpoint (keep existing implementation)
- Switching automatically to another backend on rate limit

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-17 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
