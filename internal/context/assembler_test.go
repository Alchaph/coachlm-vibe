package context

import (
	"strings"
	"testing"
	"time"

	"coachlm/internal/storage"
)

func makeProfile() *storage.AthleteProfile {
	return &storage.AthleteProfile{
		Age:                 35,
		MaxHR:               185,
		ThresholdPaceSecs:   270,
		WeeklyMileageTarget: 50.0,
		RaceGoals:           "Sub-3:30 marathon",
		InjuryHistory:       "IT band 2024",
	}
}

func makeActivities(now time.Time) []storage.Activity {
	return []storage.Activity{
		{Name: "Morning Run", ActivityType: "Run", StartDate: now.AddDate(0, 0, -1), Distance: 10000, DurationSecs: 3000, AvgPaceSecs: 300, AvgHR: 145},
		{Name: "Tempo Run", ActivityType: "Run", StartDate: now.AddDate(0, 0, -3), Distance: 8000, DurationSecs: 2400, AvgPaceSecs: 270, AvgHR: 160},
		{Name: "Long Run", ActivityType: "Run", StartDate: now.AddDate(0, 0, -10), Distance: 20000, DurationSecs: 6000, AvgPaceSecs: 310, AvgHR: 140},
	}
}

func makeInsights() []storage.PinnedInsight {
	return []storage.PinnedInsight{
		{ID: 1, Content: "Athlete responds well to tempo intervals", SourceSessionID: "s1", CreatedAt: time.Now()},
		{ID: 2, Content: "Keep easy runs below HR 140", SourceSessionID: "s2", CreatedAt: time.Now()},
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"a", 1},
		{"ab", 1},
		{"abc", 1},
		{"abcd", 1},
		{"abcde", 2},
		{strings.Repeat("x", 100), 25},
		{strings.Repeat("x", 101), 26},
	}
	for _, tt := range tests {
		got := EstimateTokens(tt.input)
		if got != tt.expected {
			t.Errorf("EstimateTokens(%d chars) = %d, want %d", len(tt.input), got, tt.expected)
		}
	}
}

func TestDefaultPromptConfig(t *testing.T) {
	cfg := DefaultPromptConfig()
	if cfg.TokenBudget != 4000 {
		t.Errorf("DefaultPromptConfig().TokenBudget = %d, want 4000", cfg.TokenBudget)
	}
}

func TestAssemblePrompt_AllBlocksFit(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	input := PromptInput{
		Profile:    makeProfile(),
		Activities: makeActivities(now),
		Insights:   makeInsights(),
		Now:        now,
	}
	config := PromptConfig{TokenBudget: 10000}

	result := AssemblePrompt(input, config)

	if !strings.Contains(result, systemPreamble) {
		t.Error("missing system preamble")
	}
	if !strings.Contains(result, "## Coaching Insights") {
		t.Error("missing insights section")
	}
	if !strings.Contains(result, "## Athlete Profile") {
		t.Error("missing profile section")
	}
	if !strings.Contains(result, "## Training Summary") {
		t.Error("missing training summary section")
	}
	if EstimateTokens(result) > config.TokenBudget {
		t.Errorf("result exceeds token budget: %d > %d", EstimateTokens(result), config.TokenBudget)
	}
}

func TestAssemblePrompt_AssemblyOrder(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	input := PromptInput{
		Profile:    makeProfile(),
		Activities: makeActivities(now),
		Insights:   makeInsights(),
		Now:        now,
	}
	config := PromptConfig{TokenBudget: 10000}

	result := AssemblePrompt(input, config)

	preambleIdx := strings.Index(result, "You are CoachLM")
	insightsIdx := strings.Index(result, "## Coaching Insights")
	profileIdx := strings.Index(result, "## Athlete Profile")
	trainingIdx := strings.Index(result, "## Training Summary")

	if preambleIdx >= insightsIdx {
		t.Error("preamble should come before insights")
	}
	if insightsIdx >= profileIdx {
		t.Error("insights should come before profile")
	}
	if profileIdx >= trainingIdx {
		t.Error("profile should come before training summary")
	}
}

func TestAssemblePrompt_TrainingSummaryTruncated(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)

	var activities []storage.Activity
	for i := 0; i < 30; i++ {
		activities = append(activities, storage.Activity{
			Name:         "Run",
			ActivityType: "Run",
			StartDate:    now.AddDate(0, 0, -i),
			Distance:     10000,
			DurationSecs: 3000,
			AvgPaceSecs:  300,
			AvgHR:        145,
		})
	}

	input := PromptInput{
		Profile:    makeProfile(),
		Activities: activities,
		Insights:   makeInsights(),
		Now:        now,
	}

	fullResult := AssemblePrompt(input, PromptConfig{TokenBudget: 100000})
	fullTokens := EstimateTokens(fullResult)

	tightBudget := fullTokens / 2
	result := AssemblePrompt(input, PromptConfig{TokenBudget: tightBudget})

	if EstimateTokens(result) > tightBudget {
		t.Errorf("result exceeds tight budget: %d > %d", EstimateTokens(result), tightBudget)
	}
	if !strings.Contains(result, "## Athlete Profile") {
		t.Error("profile should still be present when only training is truncated")
	}
	if !strings.Contains(result, "## Coaching Insights") {
		t.Error("insights must always be present")
	}
}

func TestAssemblePrompt_NoTrainingData(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	input := PromptInput{
		Profile:    makeProfile(),
		Activities: nil,
		Insights:   makeInsights(),
		Now:        now,
	}
	config := PromptConfig{TokenBudget: 4000}

	result := AssemblePrompt(input, config)

	if !strings.Contains(result, systemPreamble) {
		t.Error("missing preamble")
	}
	if !strings.Contains(result, "## Coaching Insights") {
		t.Error("missing insights")
	}
	if !strings.Contains(result, "## Athlete Profile") {
		t.Error("missing profile")
	}
	if !strings.Contains(result, "No recent training data") {
		t.Error("expected 'No recent training data' text")
	}
}

func TestAssemblePrompt_PinnedInsightsNeverTruncated(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)

	var insights []storage.PinnedInsight
	for i := 0; i < 50; i++ {
		insights = append(insights, storage.PinnedInsight{
			ID:      int64(i),
			Content: strings.Repeat("Important insight content here. ", 10),
		})
	}

	input := PromptInput{
		Profile:    makeProfile(),
		Activities: makeActivities(now),
		Insights:   insights,
		Now:        now,
	}

	insightsBlock := formatInsightsBlock(insights)
	tinyBudget := EstimateTokens(systemPreamble) / 2

	result := AssemblePrompt(input, PromptConfig{TokenBudget: tinyBudget})

	for _, ins := range insights {
		if !strings.Contains(result, ins.Content) {
			t.Errorf("insight content was truncated: %q", ins.Content[:40])
		}
	}
	if !strings.Contains(result, insightsBlock) {
		t.Error("insights block was modified")
	}
}

func TestAssemblePrompt_EmptyEverything(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	input := PromptInput{
		Profile:    nil,
		Activities: nil,
		Insights:   nil,
		Now:        now,
	}
	config := PromptConfig{TokenBudget: 4000}

	result := AssemblePrompt(input, config)

	if !strings.Contains(result, systemPreamble) {
		t.Error("preamble must always be present")
	}
	if !strings.Contains(result, "No profile configured") {
		t.Error("expected 'No profile configured'")
	}
	if !strings.Contains(result, "No recent training data") {
		t.Error("expected 'No recent training data'")
	}
	if strings.Contains(result, "## Coaching Insights") {
		t.Error("insights section should not appear when there are no insights")
	}
}

func TestAssemblePrompt_LargeProfileTruncated(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)

	largeProfile := &storage.AthleteProfile{
		Age:                 35,
		MaxHR:               185,
		ThresholdPaceSecs:   270,
		WeeklyMileageTarget: 50.0,
		RaceGoals:           strings.Repeat("Marathon goal details. ", 200),
		InjuryHistory:       strings.Repeat("Injury detail. ", 200),
	}

	input := PromptInput{
		Profile:    largeProfile,
		Activities: nil,
		Insights:   makeInsights(),
		Now:        now,
	}

	preambleTokens := EstimateTokens(systemPreamble)
	insightsTokens := EstimateTokens(formatInsightsBlock(input.Insights))
	budget := preambleTokens + insightsTokens + 50

	result := AssemblePrompt(input, PromptConfig{TokenBudget: budget})

	if !strings.Contains(result, "## Coaching Insights") {
		t.Error("insights must be present")
	}
	if !strings.Contains(result, systemPreamble) {
		t.Error("preamble must be present")
	}

	profileBlock := FormatProfileBlock(largeProfile)
	if strings.Contains(result, profileBlock) {
		t.Error("full profile should have been truncated")
	}
}

func TestAssemblePrompt_BudgetSmallerThanPreamblePlusInsights(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	insights := makeInsights()

	input := PromptInput{
		Profile:    makeProfile(),
		Activities: makeActivities(now),
		Insights:   insights,
		Now:        now,
	}

	result := AssemblePrompt(input, PromptConfig{TokenBudget: 10})

	if !strings.Contains(result, systemPreamble) {
		t.Error("preamble must always be present")
	}
	for _, ins := range insights {
		if !strings.Contains(result, ins.Content) {
			t.Errorf("insight must never be dropped: %q", ins.Content)
		}
	}
}

func TestAssemblePrompt_ZeroBudgetUsesDefault(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	input := PromptInput{
		Profile: makeProfile(),
		Now:     now,
	}

	result := AssemblePrompt(input, PromptConfig{TokenBudget: 0})
	if !strings.Contains(result, systemPreamble) {
		t.Error("should use default budget and include preamble")
	}
}

func TestAssemblePrompt_NoInsights(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	input := PromptInput{
		Profile:    makeProfile(),
		Activities: makeActivities(now),
		Insights:   nil,
		Now:        now,
	}
	config := PromptConfig{TokenBudget: 10000}

	result := AssemblePrompt(input, config)

	if strings.Contains(result, "## Coaching Insights") {
		t.Error("insights section should not appear when empty")
	}
	if !strings.Contains(result, "## Athlete Profile") {
		t.Error("profile should be present")
	}
	if !strings.Contains(result, "## Training Summary") {
		t.Error("training summary should be present")
	}
}

func TestAssemblePrompt_SectionsSeparatedByDoubleNewline(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	input := PromptInput{
		Profile:    makeProfile(),
		Activities: makeActivities(now),
		Insights:   makeInsights(),
		Now:        now,
	}
	config := PromptConfig{TokenBudget: 10000}

	result := AssemblePrompt(input, config)

	parts := strings.Split(result, "\n\n")
	if len(parts) < 4 {
		t.Errorf("expected at least 4 sections separated by double newlines, got %d", len(parts))
	}
}

func TestTruncateToTokens(t *testing.T) {
	text := "line1\nline2\nline3\nline4\nline5"

	t.Run("fits within budget", func(t *testing.T) {
		result := truncateToTokens(text, 1000)
		if result != text {
			t.Error("should return full text when within budget")
		}
	})

	t.Run("removes lines from end", func(t *testing.T) {
		budget := EstimateTokens("line1\nline2")
		result := truncateToTokens(text, budget)
		if !strings.HasPrefix(result, "line1") {
			t.Error("should keep first lines")
		}
		if strings.Contains(result, "line5") {
			t.Error("should have removed later lines")
		}
		if EstimateTokens(result) > budget {
			t.Errorf("result exceeds budget: %d > %d", EstimateTokens(result), budget)
		}
	})

	t.Run("character-level truncation", func(t *testing.T) {
		result := truncateToTokens("abcdefghijklmnop", 2)
		if len(result) > 8 {
			t.Errorf("character truncation failed: got %d chars, want <= 8", len(result))
		}
	})

	t.Run("zero budget", func(t *testing.T) {
		result := truncateToTokens(text, 0)
		if result != "" {
			t.Error("zero budget should return empty string")
		}
	})
}

func TestFormatInsightsBlock(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		result := formatInsightsBlock(nil)
		if result != "" {
			t.Error("should return empty string for nil insights")
		}
	})

	t.Run("with insights", func(t *testing.T) {
		insights := makeInsights()
		result := formatInsightsBlock(insights)
		if !strings.HasPrefix(result, "## Coaching Insights\n") {
			t.Error("should start with heading")
		}
		for _, ins := range insights {
			if !strings.Contains(result, "- "+ins.Content) {
				t.Errorf("missing insight: %q", ins.Content)
			}
		}
	})
}
