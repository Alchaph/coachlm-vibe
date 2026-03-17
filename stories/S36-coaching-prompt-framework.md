---
id: S36
title: Coaching system prompt framework
status: done
created: 2026-03-16
updated: 2026-03-16
---

# S36 — Coaching system prompt framework

## User story

As a **runner chatting with CoachLM**,
I want **focused, actionable coaching responses instead of generic filler**
so that **every reply actually helps me train better instead of drowning me in caveats and motivational fluff**.

## Problem

The entire system prompt today is one sentence:

```
You are CoachLM, an AI running coach. You have access to the athlete's profile,
recent training data, and coaching insights. Provide personalized, evidence-based
training advice.
```

This gives the LLM almost no guidance on **how** to behave, so it falls back to its default habits:
- Walls of text with excessive hedging ("It's important to note…", "Always consult a doctor…")
- Generic advice that ignores the athlete data sitting right in the context
- Repeating the user's question back before answering
- Motivational cheerleading instead of concrete prescriptions
- No consistent response structure — sometimes bullet points, sometimes essays
- Failing to reference the actual training data (e.g., recent mileage, pace trends) even when it's right there in the prompt

A well-structured system prompt can eliminate most of these problems without changing a single line of LLM code.

## Acceptance criteria

- [ ] The `systemPreamble` constant in `assembler.go` is replaced with a multi-section prompt framework that defines:
  1. **Identity & tone** — who CoachLM is, communication style (direct, concise, no fluff)
  2. **Response rules** — concrete behavioral constraints (see examples below)
  3. **Data usage instructions** — explicit directive to reference the athlete's actual numbers from the context
  4. **Output format guidance** — default to short structured responses; use longer format only when the question genuinely requires it
- [ ] Response rules include at minimum:
  - Lead with the answer, not a preamble
  - Reference specific numbers from the athlete's profile and training data when relevant (e.g., "Your average pace this week was 5:12/km" not "based on your recent training")
  - Keep responses under ~200 words unless the user asks for a detailed plan
  - No generic safety disclaimers unless the user describes pain or injury symptoms
  - No motivational filler ("Great job!", "Keep it up!") unless the user is explicitly seeking encouragement
  - When prescribing workouts, include specific paces/distances/durations based on the athlete's threshold pace and recent volume
- [ ] The prompt framework is assembled as a Go template or string builder — NOT a single hardcoded string. Each section is a named constant or function so individual rules can be tested and modified independently.
- [ ] Existing `AssemblePrompt` signature and token budget logic are unchanged — only the preamble content changes
- [ ] Pinned insights priority is unchanged (insights are still never truncated)
- [ ] The prompt fits comfortably within the default 4000-token budget alongside typical profile + training data (the framework itself should be under ~600 tokens)

## Technical notes

- `internal/context/assembler.go`: Replace `systemPreamble` with a `buildSystemPreamble()` function that assembles sections
- Suggested prompt structure (adapt as needed):

  ```
  # CoachLM — Running Coach

  ## Role
  You are CoachLM, a direct and knowledgeable running coach.
  You have the athlete's profile, training log, and pinned coaching
  insights below. Use them.

  ## Response Rules
  - Lead with the answer. No preamble, no restating the question.
  - Reference the athlete's actual numbers (pace, mileage, HR) — never say "based on your data" without citing specifics.
  - Default to ≤150 words. Only go longer for detailed plans the user explicitly requests.
  - Prescribe specific paces and distances derived from the athlete's threshold pace and recent volume.
  - Skip generic safety disclaimers unless the user reports pain or injury.
  - No motivational filler unless asked for encouragement.
  - If data is missing (no profile, no activities), say so briefly and ask what they need.

  ## Output Format
  - Use bullet points or short paragraphs.
  - For workouts: specify warmup, main set (pace + distance/time), cooldown.
  - For questions: answer directly, then add brief reasoning if helpful.
  ```

- Keep the preamble testable: `assembler_test.go` should verify the assembled preamble contains key directives
- The context assembler already handles token budget — just make sure the new preamble + typical data stays within 4000 tokens
- Do NOT change the `LLM` interface or any backend code — this is purely a prompt content change

## Tests required

- Unit: `assembler_test.go` — `buildSystemPreamble()` output contains required sections (role, response rules, output format)
- Unit: assembled prompt with full profile + training data + new preamble fits within default 4000-token budget
- Unit: existing `AssemblePrompt` tests still pass (preamble content changed but structure/priority unchanged)
- Unit: preamble token count is under 600 tokens (prevents prompt bloat)
- Edge case: prompt with no profile and no activities — preamble still includes the "if data is missing" instruction

## Out of scope

- Per-backend prompt tuning (Claude vs OpenAI vs Ollama) — single prompt for all backends for now
- User-editable system prompt (possible future story)
- Prompt versioning / A/B testing
- Few-shot examples in the system prompt (would blow the token budget)
- Changing the context assembly priority order (insights > profile > training)

---

## Status history

| Date | Status | Notes |
|---|---|---|
| 2026-03-16 | draft | Created |

---

<!-- Agent: add a Blocker section here if status is set to failed -->
