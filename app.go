package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// App struct holds the application state and dependencies.
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the Wails runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SendMessage sends a user message to the LLM backend and returns the response.
// This is a stub that echoes the message until the LLM router (S08/S09) is wired up.
func (a *App) SendMessage(message string) (string, error) {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "", errors.New("message cannot be empty")
	}
	return fmt.Sprintf("Echo: %s", trimmed), nil
}

// SaveInsight saves a chat message as a pinned insight for future context.
// This is a stub — the real implementation will use storage.DB.SaveInsight()
// and storage.DB.InsightExists() once wired up.
func (a *App) SaveInsight(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return errors.New("insight content must not be empty")
	}
	// TODO: wire up storage.DB.InsightExists + SaveInsight
	return nil
}

// ActivityRecord is a serializable representation of an activity for the frontend dashboard.
type ActivityRecord struct {
	Name         string  `json:"name"`
	ActivityType string  `json:"activityType"`
	StartDate    string  `json:"startDate"`
	Distance     float64 `json:"distance"`
	DurationSecs int     `json:"durationSecs"`
	AvgPaceSecs  int     `json:"avgPaceSecs"`
	AvgHR        int     `json:"avgHR"`
}

// GetRecentActivities returns the most recent activities for the dashboard.
// This is a stub that returns an empty slice until the storage layer is wired up.
func (a *App) GetRecentActivities(limit int) ([]ActivityRecord, error) {
	return []ActivityRecord{}, nil
}

func (a *App) ImportFITFile(filePath string) error {
	trimmed := strings.TrimSpace(filePath)
	if trimmed == "" {
		return errors.New("file path must not be empty")
	}
	return nil
}
