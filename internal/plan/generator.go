package plan

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"coachlm/internal/llm"
	"coachlm/internal/storage"

	coachctx "coachlm/internal/context"
)

// GeneratorConfig holds configuration for plan generation.
type GeneratorConfig struct {
	// TrainingHistoryWeeks is how many weeks of activities to include (default: 8).
	TrainingHistoryWeeks int
}

// DefaultGeneratorConfig returns sensible defaults.
func DefaultGeneratorConfig() GeneratorConfig {
	return GeneratorConfig{
		TrainingHistoryWeeks: 8,
	}
}

// Generator orchestrates training plan creation via an LLM.
type Generator struct {
	llmClient llm.LLM
	store     *Storage
	db        *storage.DB
	config    GeneratorConfig
}

// NewGenerator creates a Generator.
func NewGenerator(client llm.LLM, store *Storage, db *storage.DB, config GeneratorConfig) *Generator {
	if config.TrainingHistoryWeeks <= 0 {
		config.TrainingHistoryWeeks = 8
	}
	return &Generator{
		llmClient: client,
		store:     store,
		db:        db,
		config:    config,
	}
}

// maxRetries is the number of LLM parse retries before giving up.
const maxRetries = 3

// Generate creates a training plan for the given race.
func (g *Generator) Generate(ctx context.Context, raceID string) (*TrainingPlan, error) {
	race, err := g.store.GetRace(raceID)
	if err != nil {
		return nil, fmt.Errorf("get race: %w", err)
	}
	if race == nil {
		return nil, fmt.Errorf("race not found: %s", raceID)
	}

	prompt, err := g.buildPrompt(race)
	if err != nil {
		return nil, fmt.Errorf("build prompt: %w", err)
	}

	promptHash := fmt.Sprintf("%x", sha256.Sum256([]byte(prompt)))

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: prompt},
		{Role: llm.RoleUser, Content: "Generate the training plan now. Respond with only valid JSON matching the schema described above. No prose, no markdown."},
	}

	var llmResp LLMPlanResponse
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		response, err := g.llmClient.Chat(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("llm chat (attempt %d): %w", attempt+1, err)
		}

		parsed, parseErr := parseLLMResponse(response)
		if parseErr != nil {
			lastErr = fmt.Errorf("attempt %d: %w", attempt+1, parseErr)
			fmt.Printf("[plan/generator] parse failure (attempt %d): raw response:\n%s\n", attempt+1, response)
			continue
		}

		llmResp = *parsed
		lastErr = nil
		break
	}
	if lastErr != nil {
		return nil, fmt.Errorf("failed to parse LLM response after %d attempts: %w", maxRetries, lastErr)
	}

	// Compute weeks-to-race and validate week count.
	weeksToRace := weeksUntil(time.Now(), race.RaceDate)
	if weeksToRace < 1 {
		weeksToRace = 1
	}
	if len(llmResp.Weeks) != weeksToRace {
		// Accept if close (LLM may round), but reject large mismatches.
		diff := len(llmResp.Weeks) - weeksToRace
		if diff < -2 || diff > 2 {
			return nil, fmt.Errorf("LLM returned %d weeks, expected ~%d", len(llmResp.Weeks), weeksToRace)
		}
	}

	plan := g.buildPlan(race, &llmResp, promptHash)
	if err := g.store.SavePlan(plan); err != nil {
		return nil, fmt.Errorf("save plan: %w", err)
	}
	return plan, nil
}

// buildPrompt assembles the system prompt for plan generation.
func (g *Generator) buildPrompt(race *Race) (string, error) {
	profile, _ := g.db.GetProfile()
	days := int(math.Ceil(time.Until(race.RaceDate).Hours() / 24))
	activities, _ := g.db.ListActivities(g.config.TrainingHistoryWeeks*7, 0)
	insights, _ := g.db.GetInsights()

	var sb strings.Builder

	sb.WriteString("You are an expert running coach creating a structured training plan.\n\n")
	sb.WriteString("## Instructions\n")
	sb.WriteString("- Return ONLY a valid JSON object with the schema described below.\n")
	sb.WriteString("- No prose, no markdown code fences, no explanation — just JSON.\n")
	sb.WriteString("- The plan must have exactly the right number of weeks to cover the period from now to race day.\n")
	sb.WriteString("- Each week has sessions for days 1 (Monday) through 7 (Sunday).\n")
	sb.WriteString("- Rest days should be type \"rest\" with duration_min 0.\n")
	sb.WriteString("- The final week should include a \"race\" session on race day and be a taper week.\n\n")

	sb.WriteString("## JSON Schema\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{
  "weeks": [
    {
      "week_number": 1,
      "sessions": [
        {
          "day_of_week": 1,
          "type": "easy|tempo|intervals|long_run|strength|rest|race",
          "duration_min": 45,
          "distance_km": 8.0,
          "hr_zone": 2,
          "pace_low": 5.0,
          "pace_high": 5.5,
          "notes": "Easy aerobic run at conversational pace"
        }
      ]
    }
  ]
}`)
	sb.WriteString("\n```\n\n")

	// Race details
	weeksToRace := weeksUntil(time.Now(), race.RaceDate)
	sb.WriteString("## Race Details\n")
	sb.WriteString(fmt.Sprintf("- Name: %s\n", race.Name))
	sb.WriteString(fmt.Sprintf("- Distance: %.1f km\n", race.DistanceKm))
	sb.WriteString(fmt.Sprintf("- Date: %s (%d days away, ~%d weeks)\n", race.RaceDate.Format("2006-01-02"), days, weeksToRace))
	sb.WriteString(fmt.Sprintf("- Terrain: %s\n", race.Terrain))
	if race.ElevationM != nil {
		sb.WriteString(fmt.Sprintf("- Elevation: %.0f m\n", *race.ElevationM))
	}
	if race.GoalTimeSec != nil {
		h := *race.GoalTimeSec / 3600
		m := (*race.GoalTimeSec % 3600) / 60
		s := *race.GoalTimeSec % 60
		sb.WriteString(fmt.Sprintf("- Goal time: %d:%02d:%02d\n", h, m, s))
	}
	sb.WriteString(fmt.Sprintf("- Priority: %s\n", race.Priority))

	if weeksToRace < 4 {
		sb.WriteString("\nWARNING: Less than 4 weeks to race. Create an abbreviated plan focused on taper and race preparation.\n")
	}

	sb.WriteString("\n")

	// Athlete profile
	if profile != nil {
		sb.WriteString("## Athlete Profile\n")
		sb.WriteString(coachctx.FormatProfileBlock(profile))
		sb.WriteString("\n\n")
	} else {
		sb.WriteString("## Athlete Profile\nNo profile data available. Use conservative pacing.\n\n")
	}

	// Training history
	if len(activities) > 0 {
		sb.WriteString("## Recent Training History\n")
		sb.WriteString(coachctx.FormatTrainingSummary(activities, time.Now()))
		sb.WriteString("\n\n")
	} else {
		sb.WriteString("## Recent Training History\nNo recent activity data. Build plan from profile data and race requirements only.\n\n")
	}

	// Insights
	if len(insights) > 0 {
		sb.WriteString("## Coaching Insights\n")
		for _, ins := range insights {
			sb.WriteString(fmt.Sprintf("- %s\n", ins.Content))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Generate a %d-week training plan as JSON.\n", weeksToRace))

	return sb.String(), nil
}

// parseLLMResponse parses the LLM's JSON response, stripping fences if needed.
func parseLLMResponse(raw string) (*LLMPlanResponse, error) {
	cleaned := strings.TrimSpace(raw)

	// Try direct parse first.
	var resp LLMPlanResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err == nil {
		if len(resp.Weeks) == 0 {
			return nil, fmt.Errorf("parsed plan has zero weeks")
		}
		return &resp, nil
	}

	// Strip markdown code fences and retry.
	fenceRe := regexp.MustCompile("(?s)```(?:json)?\\s*\n?(.*?)\\s*```")
	matches := fenceRe.FindStringSubmatch(cleaned)
	if len(matches) >= 2 {
		stripped := strings.TrimSpace(matches[1])
		if err := json.Unmarshal([]byte(stripped), &resp); err == nil {
			if len(resp.Weeks) == 0 {
				return nil, fmt.Errorf("parsed plan has zero weeks after fence strip")
			}
			return &resp, nil
		}
	}

	return nil, fmt.Errorf("response is not valid JSON: %.200s", cleaned)
}

// buildPlan converts the LLM response into a fully-structured TrainingPlan.
func (g *Generator) buildPlan(race *Race, resp *LLMPlanResponse, promptHash string) *TrainingPlan {
	now := time.Now()
	planID := fmt.Sprintf("plan_%d", now.UnixNano())

	plan := &TrainingPlan{
		ID:          planID,
		RaceID:      race.ID,
		GeneratedAt: now,
		LLMBackend:  g.llmClient.Name(),
		PromptHash:  promptHash,
	}

	// Calculate Monday of the current week as the plan start.
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	mondayOffset := weekday - 1
	planStart := now.AddDate(0, 0, -mondayOffset).Truncate(24 * time.Hour)

	for _, lw := range resp.Weeks {
		weekStart := planStart.AddDate(0, 0, (lw.WeekNumber-1)*7)
		weekID := fmt.Sprintf("%s_w%d", planID, lw.WeekNumber)

		week := Week{
			ID:         weekID,
			PlanID:     planID,
			WeekNumber: lw.WeekNumber,
			WeekStart:  weekStart,
		}

		for i, ls := range lw.Sessions {
			sessID := fmt.Sprintf("%s_s%d", weekID, i+1)
			sess := Session{
				ID:          sessID,
				WeekID:      weekID,
				DayOfWeek:   ls.DayOfWeek,
				Type:        SessionType(ls.Type),
				DurationMin: ls.DurationMin,
				DistanceKm:  ls.DistanceKm,
				HRZone:      ls.HRZone,
				PaceMinLow:  ls.PaceLow,
				PaceMinHigh: ls.PaceHigh,
				Notes:       truncateNotes(ls.Notes, 300),
				Status:      StatusPlanned,
			}
			// Default invalid types to "easy".
			if !validSessionTypes[sess.Type] {
				sess.Type = SessionEasy
			}
			// Clamp day_of_week.
			if sess.DayOfWeek < 1 {
				sess.DayOfWeek = 1
			}
			if sess.DayOfWeek > 7 {
				sess.DayOfWeek = 7
			}
			week.Sessions = append(week.Sessions, sess)
		}

		plan.Weeks = append(plan.Weeks, week)
	}

	return plan
}

// weeksUntil returns the number of full or partial weeks between now and target.
func weeksUntil(now, target time.Time) int {
	days := int(math.Ceil(target.Sub(now).Hours() / 24))
	weeks := (days + 6) / 7
	if weeks < 1 {
		return 1
	}
	return weeks
}

// truncateNotes trims notes to maxLen characters.
func truncateNotes(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
