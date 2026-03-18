package main

import (
	"strings"
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
		UseLocalModel:  true,
		OllamaEndpoint: "http://localhost:11434",
		OllamaModel:    "llama3.1",
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

	if got.UseLocalModel != want.UseLocalModel {
		t.Errorf("UseLocalModel = %v, want %v", got.UseLocalModel, want.UseLocalModel)
	}
	if got.OllamaEndpoint != want.OllamaEndpoint {
		t.Errorf("OllamaEndpoint = %q, want %q", got.OllamaEndpoint, want.OllamaEndpoint)
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

func TestStartStravaAuth_NoCredentials(t *testing.T) {
	app := newTestApp(t)

	err := app.StartStravaAuth()
	if err == nil {
		t.Error("expected error when no credentials are available")
	}
	if err != nil && !strings.Contains(err.Error(), "no client credentials available") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGetStravaCredentialsAvailable_NoCredentials(t *testing.T) {
	app := newTestApp(t)
	if app.GetStravaCredentialsAvailable() {
		t.Error("expected credentials not available when no ldflags or env vars set")
	}
}

func TestResolveStravaCredentials_EnvVars(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "env-id-123")
	t.Setenv("STRAVA_CLIENT_SECRET", "env-secret-456")

	id, secret, ok := resolveStravaCredentials()
	if !ok {
		t.Fatal("expected ok=true with env vars set")
	}
	if id != "env-id-123" {
		t.Errorf("clientID = %q, want %q", id, "env-id-123")
	}
	if secret != "env-secret-456" {
		t.Errorf("clientSecret = %q, want %q", secret, "env-secret-456")
	}
}

func TestResolveStravaCredentials_Empty(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "")
	t.Setenv("STRAVA_CLIENT_SECRET", "")

	_, _, ok := resolveStravaCredentials()
	if ok {
		t.Error("expected ok=false with no credentials")
	}
}

func TestReloadLLMClient(t *testing.T) {
	app := newTestApp(t)

	if err := app.SaveSettingsData(SettingsData{
		UseLocalModel:  true,
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

func TestGetContextPreview_Empty(t *testing.T) {
	app := newTestApp(t)

	preview, err := app.GetContextPreview()
	if err != nil {
		t.Fatalf("GetContextPreview: %v", err)
	}
	if preview == "" {
		t.Error("expected non-empty preview even with no data")
	}
	if !strings.Contains(preview, "CoachLM") {
		t.Error("expected preview to contain 'CoachLM' preamble")
	}
}

func TestGetContextPreview_WithProfile(t *testing.T) {
	app := newTestApp(t)

	if err := app.SaveProfileData(ProfileData{
		Age:                 30,
		MaxHR:               185,
		ThresholdPaceSecs:   300,
		WeeklyMileageTarget: 50,
		RaceGoals:           "Sub-3:30 marathon",
		InjuryHistory:       "None",
	}); err != nil {
		t.Fatalf("SaveProfileData: %v", err)
	}

	preview, err := app.GetContextPreview()
	if err != nil {
		t.Fatalf("GetContextPreview: %v", err)
	}
	if !strings.Contains(preview, "30") {
		t.Error("expected preview to contain age")
	}
	if !strings.Contains(preview, "185") {
		t.Error("expected preview to contain max HR")
	}
}

func TestGetContextPreview_WithInsights(t *testing.T) {
	app := newTestApp(t)

	if err := app.SaveInsight("Always warm up before tempo runs"); err != nil {
		t.Fatalf("SaveInsight: %v", err)
	}

	preview, err := app.GetContextPreview()
	if err != nil {
		t.Fatalf("GetContextPreview: %v", err)
	}
	if !strings.Contains(preview, "warm up before tempo") {
		t.Error("expected preview to contain the saved insight")
	}
}
