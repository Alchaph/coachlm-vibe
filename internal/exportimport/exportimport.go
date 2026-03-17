package exportimport

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"coachlm/internal/storage"
)

const schemaVersion = 1

type ExportEnvelope struct {
	SchemaVersion     int                     `json:"schema_version"`
	ExportedAt        string                  `json:"exported_at"`
	AthleteProfile    *storage.AthleteProfile `json:"athlete_profile"`
	TrainingSummaries []TrainingSummaryExport `json:"training_summaries"`
	PinnedInsights    []storage.PinnedInsight `json:"pinned_insights"`
	SettingsMeta      SettingsMeta            `json:"settings_meta"`
}

type TrainingSummaryExport struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	ActivityType string  `json:"activity_type"`
	StartDate    string  `json:"start_date"`
	Distance     float64 `json:"distance"`
	DurationSecs int     `json:"duration_secs"`
	AvgPaceSecs  int     `json:"avg_pace_secs"`
	AvgHR        int     `json:"avg_hr"`
	MaxHR        int     `json:"max_hr"`
	AvgCadence   float64 `json:"avg_cadence"`
	Source       string  `json:"source"`
}

type SettingsMeta struct {
	ActiveLLM string `json:"active_llm"`
}

type ImportEnvelope struct {
	SchemaVersion     int                     `json:"schema_version"`
	ExportedAt        string                  `json:"exported_at"`
	AthleteProfile    *storage.AthleteProfile `json:"athlete_profile"`
	TrainingSummaries []TrainingSummaryExport `json:"training_summaries"`
	PinnedInsights    []storage.PinnedInsight `json:"pinned_insights"`
	SettingsMeta      SettingsMeta            `json:"settings_meta"`
}

func Export(db *storage.DB, filePath string) error {
	if db == nil {
		return errors.New("database is nil")
	}

	profile, _ := db.GetProfile()
	activities, _ := db.ListActivities(1000, 0)
	insights, _ := db.GetInsights()
	settings, _ := db.GetSettings()

	summaries := make([]TrainingSummaryExport, 0, len(activities))
	for i := range activities {
		summaries = append(summaries, TrainingSummaryExport{
			ID:           activities[i].ID,
			Name:         activities[i].Name,
			ActivityType: activities[i].ActivityType,
			StartDate:    activities[i].StartDate.Format(time.RFC3339),
			Distance:     activities[i].Distance,
			DurationSecs: activities[i].DurationSecs,
			AvgPaceSecs:  activities[i].AvgPaceSecs,
			AvgHR:        activities[i].AvgHR,
			MaxHR:        activities[i].MaxHR,
			AvgCadence:   activities[i].AvgCadence,
			Source:       activities[i].Source,
		})
	}

	var activeLLM string
	if settings != nil {
		activeLLM = settings.ActiveLLM
	}

	envelope := ExportEnvelope{
		SchemaVersion:     schemaVersion,
		ExportedAt:        time.Now().Format(time.RFC3339),
		AthleteProfile:    profile,
		TrainingSummaries: summaries,
		PinnedInsights:    insights,
		SettingsMeta: SettingsMeta{
			ActiveLLM: activeLLM,
		},
	}

	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal export: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}

	return nil
}

func Import(db *storage.DB, filePath string, replaceAll bool) error {
	if db == nil {
		return errors.New("database is nil")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read import file: %w", err)
	}

	var envelope ImportEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("unmarshal import file: %w", err)
	}

	if envelope.SchemaVersion > schemaVersion {
		return fmt.Errorf("unsupported schema version: %d (max supported: %d)", envelope.SchemaVersion, schemaVersion)
	}

	if replaceAll {
		if err := db.ReplaceAllContext(); err != nil {
			return fmt.Errorf("clear context before import: %w", err)
		}
	}

	if envelope.AthleteProfile != nil {
		if err := db.SaveProfile(envelope.AthleteProfile); err != nil {
			return fmt.Errorf("import profile: %w", err)
		}
	}

	for _, summary := range envelope.TrainingSummaries {
		startDate, err := time.Parse(time.RFC3339, summary.StartDate)
		if err != nil {
			return fmt.Errorf("parse training start date: %w", err)
		}
		activity := &storage.Activity{
			ID:           summary.ID,
			Name:         summary.Name,
			ActivityType: summary.ActivityType,
			StartDate:    startDate,
			Distance:     summary.Distance,
			DurationSecs: summary.DurationSecs,
			AvgPaceSecs:  summary.AvgPaceSecs,
			AvgHR:        summary.AvgHR,
			MaxHR:        summary.MaxHR,
			AvgCadence:   summary.AvgCadence,
			Source:       summary.Source,
		}
		if err := db.SaveActivity(activity); err != nil {
			return fmt.Errorf("import activity %q: %w", activity.Name, err)
		}
	}

	for _, insight := range envelope.PinnedInsights {
		if _, err := db.SaveInsight(insight.Content, insight.SourceSessionID); err != nil {
			return fmt.Errorf("import insight: %w", err)
		}
	}

	if envelope.SettingsMeta.ActiveLLM != "" {
		settings, err := db.GetSettings()
		if err != nil {
			return fmt.Errorf("get settings for LLM import: %w", err)
		}
		if settings == nil {
			return errors.New("settings table not initialized")
		}
		settings.ActiveLLM = envelope.SettingsMeta.ActiveLLM
		if err := db.SaveSettings(settings); err != nil {
			return fmt.Errorf("import LLM setting: %w", err)
		}
	}

	return nil
}
