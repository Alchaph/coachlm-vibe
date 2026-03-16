package context

import (
	"fmt"
	"strings"
	"time"

	"coachlm/internal/storage"
)

const systemPreamble = "You are CoachLM, an AI running coach. You have access to the athlete's profile, recent training data, and coaching insights. Provide personalized, evidence-based training advice."

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
	Profile    *storage.AthleteProfile
	Activities []storage.Activity
	Insights   []storage.PinnedInsight
	Now        time.Time
}

// EstimateTokens returns an estimated token count: 4 characters ≈ 1 token (ceiling division).
func EstimateTokens(text string) int {
	return (len(text) + 3) / 4
}

func formatInsightsBlock(insights []storage.PinnedInsight) string {
	if len(insights) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Coaching Insights\n")
	for _, ins := range insights {
		sb.WriteString(fmt.Sprintf("- %s\n", ins.Content))
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

	preamble := systemPreamble
	insightsBlock := formatInsightsBlock(input.Insights)
	profileBlock := "## Athlete Profile\n" + FormatProfileBlock(input.Profile)
	trainingBlock := FormatTrainingSummary(input.Activities, input.Now)

	sacredText := preamble
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
		return full
	}

	tokensForTraining := remainingTokens - profileTokens
	if tokensForTraining > 0 {
		truncatedTraining := truncateToTokens(trainingBlock, tokensForTraining)
		if truncatedTraining != "" {
			return joinSections(preamble, insightsBlock, profileBlock, truncatedTraining)
		}
	}

	if tokensForTraining >= 0 {
		return joinSections(preamble, insightsBlock, profileBlock, "")
	}

	tokensForProfile := remainingTokens
	if tokensForProfile > 0 {
		truncatedProfile := truncateToTokens(profileBlock, tokensForProfile)
		return joinSections(preamble, insightsBlock, truncatedProfile, "")
	}

	return joinSections(preamble, insightsBlock, "", "")
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
