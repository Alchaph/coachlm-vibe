package main

import (
	"testing"
	"time"

	"coachlm/internal/llm"
	"coachlm/internal/storage"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	db, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("newTestApp: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &App{
		db:        db,
		llmClient: llm.NewLocal(llm.LocalConfig{}),
	}
}

func TestIsFirstRun_NoSettings(t *testing.T) {
	app := newTestApp(t)

	first, err := app.IsFirstRun()
	if err != nil {
		t.Fatalf("IsFirstRun: %v", err)
	}
	if !first {
		t.Error("expected IsFirstRun=true on empty DB")
	}
}

func TestIsFirstRun_WithSettings(t *testing.T) {
	app := newTestApp(t)

	if err := app.SaveSettingsData(SettingsData{
		ActiveLLM:      "local",
		OllamaEndpoint: "http://localhost:11434",
	}); err != nil {
		t.Fatalf("SaveSettingsData: %v", err)
	}

	first, err := app.IsFirstRun()
	if err != nil {
		t.Fatalf("IsFirstRun: %v", err)
	}
	if first {
		t.Error("expected IsFirstRun=false after saving settings")
	}
}

func TestGetSettingsData_Empty(t *testing.T) {
	app := newTestApp(t)

	data, err := app.GetSettingsData()
	if err != nil {
		t.Fatalf("GetSettingsData: %v", err)
	}
	if data != nil {
		t.Error("expected nil settings on empty DB")
	}
}

func TestSaveAndGetSettingsData_RoundTrip(t *testing.T) {
	app := newTestApp(t)

	want := SettingsData{
		ClaudeAPIKey:       "sk-claude-key",
		OpenAIAPIKey:       "sk-openai-key",
		ActiveLLM:          "claude",
		OllamaEndpoint:     "http://localhost:11434",
		StravaClientID:     "strava-id",
		StravaClientSecret: "strava-secret",
		ClaudeModel:        "claude-opus-4-20250514",
		OpenAIModel:        "gpt-4o-mini",
		OllamaModel:        "llama3.1",
	}

	if err := app.SaveSettingsData(want); err != nil {
		t.Fatalf("SaveSettingsData: %v", err)
	}

	got, err := app.GetSettingsData()
	if err != nil {
		t.Fatalf("GetSettingsData: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil settings")
	}

	if got.ClaudeAPIKey != want.ClaudeAPIKey {
		t.Errorf("ClaudeAPIKey = %q, want %q", got.ClaudeAPIKey, want.ClaudeAPIKey)
	}
	if got.OpenAIAPIKey != want.OpenAIAPIKey {
		t.Errorf("OpenAIAPIKey = %q, want %q", got.OpenAIAPIKey, want.OpenAIAPIKey)
	}
	if got.ActiveLLM != want.ActiveLLM {
		t.Errorf("ActiveLLM = %q, want %q", got.ActiveLLM, want.ActiveLLM)
	}
	if got.OllamaEndpoint != want.OllamaEndpoint {
		t.Errorf("OllamaEndpoint = %q, want %q", got.OllamaEndpoint, want.OllamaEndpoint)
	}
	if got.StravaClientID != want.StravaClientID {
		t.Errorf("StravaClientID = %q, want %q", got.StravaClientID, want.StravaClientID)
	}
	if got.StravaClientSecret != want.StravaClientSecret {
		t.Errorf("StravaClientSecret = %q, want %q", got.StravaClientSecret, want.StravaClientSecret)
	}
	if got.ClaudeModel != want.ClaudeModel {
		t.Errorf("ClaudeModel = %q, want %q", got.ClaudeModel, want.ClaudeModel)
	}
	if got.OpenAIModel != want.OpenAIModel {
		t.Errorf("OpenAIModel = %q, want %q", got.OpenAIModel, want.OpenAIModel)
	}
	if got.OllamaModel != want.OllamaModel {
		t.Errorf("OllamaModel = %q, want %q", got.OllamaModel, want.OllamaModel)
	}
}

func TestGetStravaAuthStatus_NotConnected(t *testing.T) {
	app := newTestApp(t)

	status, err := app.GetStravaAuthStatus()
	if err != nil {
		t.Fatalf("GetStravaAuthStatus: %v", err)
	}
	connected, ok := status["connected"].(bool)
	if !ok {
		t.Fatal("expected connected to be a bool")
	}
	if connected {
		t.Error("expected connected=false on empty DB")
	}
}

func TestDisconnectStrava(t *testing.T) {
	app := newTestApp(t)

	if err := app.db.SaveTokens([]byte("access"), []byte("refresh"), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}

	status, err := app.GetStravaAuthStatus()
	if err != nil {
		t.Fatalf("GetStravaAuthStatus: %v", err)
	}
	if !status["connected"].(bool) {
		t.Fatal("expected connected=true after saving tokens")
	}

	if err := app.DisconnectStrava(); err != nil {
		t.Fatalf("DisconnectStrava: %v", err)
	}

	status, err = app.GetStravaAuthStatus()
	if err != nil {
		t.Fatalf("GetStravaAuthStatus after disconnect: %v", err)
	}
	if status["connected"].(bool) {
		t.Error("expected connected=false after disconnect")
	}
}

func TestStartStravaAuth_NoSettings(t *testing.T) {
	app := newTestApp(t)

	err := app.StartStravaAuth()
	if err == nil {
		t.Error("expected error when no settings exist")
	}
}

func TestStartStravaAuth_MissingCredentials(t *testing.T) {
	app := newTestApp(t)

	if err := app.SaveSettingsData(SettingsData{
		ActiveLLM:      "local",
		OllamaEndpoint: "http://localhost:11434",
	}); err != nil {
		t.Fatalf("SaveSettingsData: %v", err)
	}

	err := app.StartStravaAuth()
	if err == nil {
		t.Error("expected error when strava credentials are empty")
	}
}

func TestReloadLLMClient(t *testing.T) {
	app := newTestApp(t)

	if err := app.SaveSettingsData(SettingsData{
		ActiveLLM:      "local",
		OllamaEndpoint: "http://localhost:11434",
	}); err != nil {
		t.Fatalf("SaveSettingsData: %v", err)
	}

	if app.llmClient == nil {
		t.Error("expected llmClient to be non-nil after save")
	}
	if app.llmClient.Name() != "local" {
		t.Errorf("expected llmClient.Name()=local, got %q", app.llmClient.Name())
	}
}
