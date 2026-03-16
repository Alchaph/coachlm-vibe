package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	coachctx "coachlm/internal/context"
	"coachlm/internal/fit"
	"coachlm/internal/llm"
	"coachlm/internal/storage"
)

type App struct {
	ctx       context.Context
	db        *storage.DB
	llmClient llm.LLM
	sessionID string
}

func NewApp(db *storage.DB, llmClient llm.LLM) *App {
	return &App{db: db, llmClient: llmClient}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SendMessage(message string) (string, error) {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "", errors.New("message cannot be empty")
	}

	if a.sessionID == "" {
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		sess, err := a.db.CreateSession(id)
		if err != nil {
			return "", fmt.Errorf("create session: %w", err)
		}
		a.sessionID = sess.ID
	}

	if _, err := a.db.SaveMessage(a.sessionID, "user", trimmed); err != nil {
		return "", fmt.Errorf("save user message: %w", err)
	}

	profile, _ := a.db.GetProfile()
	activities, _ := a.db.ListActivities(28, 0)
	insights, _ := a.db.GetInsights()

	systemPrompt := coachctx.AssemblePrompt(coachctx.PromptInput{
		Profile:    profile,
		Activities: activities,
		Insights:   insights,
		Now:        time.Now(),
	}, coachctx.DefaultPromptConfig())

	history, _ := a.db.GetMessages(a.sessionID)
	var msgs []llm.Message
	msgs = append(msgs, llm.Message{Role: llm.RoleSystem, Content: systemPrompt})
	for _, m := range history {
		msgs = append(msgs, llm.Message{Role: m.Role, Content: m.Content})
	}

	response, err := a.llmClient.Chat(a.ctx, msgs)
	if err != nil {
		return "", fmt.Errorf("llm chat: %w", err)
	}

	if _, err := a.db.SaveMessage(a.sessionID, "assistant", response); err != nil {
		return "", fmt.Errorf("save assistant message: %w", err)
	}

	return response, nil
}

func (a *App) SaveInsight(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return errors.New("insight content must not be empty")
	}

	exists, err := a.db.InsightExists(trimmed)
	if err != nil {
		return fmt.Errorf("check insight exists: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := a.db.SaveInsight(trimmed, a.sessionID); err != nil {
		return fmt.Errorf("save insight: %w", err)
	}
	return nil
}

type ActivityRecord struct {
	Name         string  `json:"name"`
	ActivityType string  `json:"activityType"`
	StartDate    string  `json:"startDate"`
	Distance     float64 `json:"distance"`
	DurationSecs int     `json:"durationSecs"`
	AvgPaceSecs  int     `json:"avgPaceSecs"`
	AvgHR        int     `json:"avgHR"`
}

func (a *App) GetRecentActivities(limit int) ([]ActivityRecord, error) {
	activities, err := a.db.ListActivities(limit, 0)
	if err != nil {
		return nil, fmt.Errorf("list activities: %w", err)
	}

	records := make([]ActivityRecord, 0, len(activities))
	for _, act := range activities {
		records = append(records, ActivityRecord{
			Name:         act.Name,
			ActivityType: act.ActivityType,
			StartDate:    act.StartDate.Format(time.RFC3339),
			Distance:     act.Distance / 1000.0,
			DurationSecs: act.DurationSecs,
			AvgPaceSecs:  act.AvgPaceSecs,
			AvgHR:        act.AvgHR,
		})
	}
	return records, nil
}

func (a *App) ImportFITFile(filePath string) error {
	trimmed := strings.TrimSpace(filePath)
	if trimmed == "" {
		return errors.New("file path must not be empty")
	}

	parsed, err := fit.ParseFITFile(trimmed)
	if err != nil {
		return fmt.Errorf("parse FIT file: %w", err)
	}

	// Negative StravaID from content hash: avoids collision with real Strava IDs
	// (always positive) while enabling deduplication for re-imported FIT files.
	hashHex := fit.DeduplicationHash(parsed)
	hashInt, _ := strconv.ParseUint(hashHex[:16], 16, 64)
	negativeID := -int64(hashInt>>1) - 1

	activity := &storage.Activity{
		StravaID:     negativeID,
		Name:         parsed.Name,
		ActivityType: parsed.ActivityType,
		StartDate:    parsed.StartDate,
		Distance:     parsed.Distance,
		DurationSecs: parsed.DurationSecs,
		AvgPaceSecs:  parsed.AvgPaceSecs,
		AvgHR:        parsed.AvgHR,
		MaxHR:        parsed.MaxHR,
		AvgCadence:   parsed.AvgCadence,
		Source:       "fit_import",
	}

	if err := a.db.SaveActivity(activity); err != nil {
		return fmt.Errorf("save activity: %w", err)
	}

	saved, err := a.db.GetActivityByStravaID(negativeID)
	if err != nil || saved == nil {
		return nil
	}

	if parsed.HeartRate != nil {
		if data, err := json.Marshal(parsed.HeartRate); err == nil {
			_ = a.db.SaveActivityStream(saved.ID, "heartrate", data)
		}
	}
	if parsed.Pace != nil {
		if data, err := json.Marshal(parsed.Pace); err == nil {
			_ = a.db.SaveActivityStream(saved.ID, "pace", data)
		}
	}
	if parsed.Cadence != nil {
		if data, err := json.Marshal(parsed.Cadence); err == nil {
			_ = a.db.SaveActivityStream(saved.ID, "cadence", data)
		}
	}

	return nil
}
