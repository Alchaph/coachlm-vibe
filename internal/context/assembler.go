package context

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"coachlm/internal/storage"
)

// buildSystemPreamble constructs the coaching system prompt framework.
func buildSystemPreamble() string {
	return `# CoachLM — Running Coach

## Role
You are CoachLM, a direct and knowledgeable running coach.
You have the athlete's profile, training log, and saved coaching
insights below. Use them. When coaching insights are present,
weave them naturally into your advice — do not repeat them verbatim.

## Response Rules
- Lead with the answer. No preamble, no restating the question.
- Reference the athlete's actual numbers (pace, mileage, HR) — never say "based on your data" without citing specifics.
- Default to ≤150 words. Only go longer for detailed plans the user explicitly requests.
- Prescribe specific paces and distances derived from the athlete's threshold pace and recent volume.
- Skip generic safety disclaimers unless the user reports pain or injury.
- No motivational filler unless asked for encouragement.
- If data is missing (no profile, no activities), say so briefly and ask what they need.

## When to Generate Training Plans
- Do NOT generate a full training plan unless:
  1. The user explicitly asks for one ("Give me a plan", "Create a schedule")
  2. The user clicks the "Generate Training Plan" button
  3. The coach determines a plan is needed and asks for permission first
- Default to direct, principle-based advice (e.g., "Focus on threshold work to improve your 5K pace")
- Reference the athlete's threshold pace, recent mileage, HR zones when giving advice
- If the user's question is broad ("How do I get faster?"), explain the approach, then ask if they want a plan

## Output Format
- Use bullet points or short paragraphs.
- For workouts: specify warmup, main set (pace + distance/time), cooldown.
- For questions: answer directly, then add brief reasoning if helpful.
`
}

// PromptConfig holds configuration for prompt assembly.
type PromptConfig struct {
	TokenBudget int
}

// DefaultPromptConfig returns a PromptConfig with TokenBudget=4000.
func DefaultPromptConfig() PromptConfig {
	return PromptConfig{
		TokenBudget: 4000,
	}
}

// PromptInput holds all the data needed to assemble a prompt.
type PromptInput struct {
	Profile      *storage.AthleteProfile
	Activities   []storage.Activity
	Insights     []storage.PinnedInsight
	CustomPrompt string
	Now          time.Time
}

// EstimateTokens returns an estimated token count: 4 characters ≈ 1 token (ceiling division).
func EstimateTokens(text string) int {
	return (len(text) + 3) / 4
}

// maxInsightChars is the maximum character length for a single insight's content
// before it gets truncated with an ellipsis.
const maxInsightChars = 500

func formatInsightsBlock(insights []storage.PinnedInsight) string {
	if len(insights) == 0 {
		return ""
	}

	sorted := make([]storage.PinnedInsight, len(insights))
	copy(sorted, insights)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	var sb strings.Builder
	sb.WriteString("## Saved Coaching Insights\n")
	sb.WriteString("(Reference these when relevant. Build on prior guidance. Avoid repeating verbatim.)\n")
	for _, ins := range sorted {
		content := ins.Content
		if len(content) > maxInsightChars {
			content = content[:maxInsightChars] + "…"
		}
		if ins.CreatedAt.IsZero() {
			sb.WriteString(fmt.Sprintf("- %s\n", content))
		} else {
			sb.WriteString(fmt.Sprintf("- %s [%s]\n", content, ins.CreatedAt.Format("Jan 02")))
		}
	}
	return sb.String()
}

// AssemblePrompt builds a system prompt from the given input, enforcing the token budget.
//
// Priority (high → low): preamble > pinned insights (NEVER cut) > profile > training summary.
// Training summary is truncated first; profile is truncated next; insights are sacred.
func AssemblePrompt(input PromptInput, config PromptConfig) string {
	if config.TokenBudget <= 0 {
		config.TokenBudget = DefaultPromptConfig().TokenBudget
	}

	preamble := buildSystemPreamble()
	insightsBlock := formatInsightsBlock(input.Insights)
	profileBlock := "## Athlete Profile\n" + FormatProfileBlock(input.Profile)
	trainingBlock := FormatTrainingSummary(input.Activities, input.Now)

	customBlock := ""
	if input.CustomPrompt != "" {
		customBlock = "\n\n## Your Custom Instructions\n" + input.CustomPrompt
	}

	sacredText := preamble + customBlock
	if insightsBlock != "" {
		sacredText += "\n\n" + insightsBlock
	}
	sacredTokens := EstimateTokens(sacredText)

	full := joinSections(preamble, insightsBlock, profileBlock, trainingBlock)
	if EstimateTokens(full) <= config.TokenBudget {
		return full
	}

	remainingTokens := config.TokenBudget - sacredTokens
	separatorTokens := 0
	if profileBlock != "" {
		separatorTokens += EstimateTokens("\n\n")
	}
	if trainingBlock != "" {
		separatorTokens += EstimateTokens("\n\n")
	}
	remainingTokens -= separatorTokens

	profileTokens := EstimateTokens(profileBlock)

	if profileTokens+EstimateTokens(trainingBlock) <= remainingTokens {
		return joinSections(preamble, customBlock, insightsBlock, profileBlock, trainingBlock)
	}

	tokensForTraining := remainingTokens - profileTokens
	if tokensForTraining > 0 {
		truncatedTraining := truncateToTokens(trainingBlock, tokensForTraining)
		if truncatedTraining != "" {
			return joinSections(preamble, customBlock, insightsBlock, profileBlock, truncatedTraining)
		}
	}

	if tokensForTraining >= 0 {
		return joinSections(preamble, customBlock, insightsBlock, profileBlock, "")
	}

	tokensForProfile := remainingTokens
	if tokensForProfile > 0 {
		truncatedProfile := truncateToTokens(profileBlock, tokensForProfile)
		return joinSections(preamble, customBlock, insightsBlock, truncatedProfile, "")
	}

	return joinSections(preamble, customBlock, insightsBlock, "", "")
}

func joinSections(parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, "\n\n")
}

// truncateToTokens removes lines from the end until the text fits within maxTokens.
// Falls back to character-level truncation if even one line exceeds the budget.
func truncateToTokens(text string, maxTokens int) string {
	if EstimateTokens(text) <= maxTokens {
		return text
	}

	lines := strings.Split(text, "\n")
	for len(lines) > 0 {
		candidate := strings.Join(lines, "\n")
		if EstimateTokens(candidate) <= maxTokens {
			return candidate
		}
		lines = lines[:len(lines)-1]
	}

	maxChars := maxTokens * 4
	if maxChars <= 0 {
		return ""
	}
	if maxChars >= len(text) {
		return text
	}
	return text[:maxChars]
}
