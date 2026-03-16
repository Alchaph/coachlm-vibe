package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	coachctx "coachlm/internal/context"
	"coachlm/internal/fit"
	"coachlm/internal/llm"
	"coachlm/internal/storage"
	"coachlm/internal/strava"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	db        *storage.DB
	llmClient llm.LLM
	sessionID string
}

type ProfileData struct {
	Age                 int     `json:"age"`
	MaxHR               int     `json:"maxHR"`
	ThresholdPaceSecs   int     `json:"thresholdPaceSecs"`
	WeeklyMileageTarget float64 `json:"weeklyMileageTarget"`
	RaceGoals           string  `json:"raceGoals"`
	InjuryHistory       string  `json:"injuryHistory"`
}

type InsightData struct {
	ID              int64  `json:"id"`
	Content         string `json:"content"`
	SourceSessionID string `json:"sourceSessionId"`
	CreatedAt       string `json:"createdAt"`
}

// SettingsData is the frontend-friendly representation of settings.
type SettingsData struct {
	ClaudeAPIKey       string `json:"claudeApiKey"`
	OpenAIAPIKey       string `json:"openaiApiKey"`
	ActiveLLM          string `json:"activeLlm"`
	OllamaEndpoint     string `json:"ollamaEndpoint"`
	StravaClientID     string `json:"stravaClientId"`
	StravaClientSecret string `json:"stravaClientSecret"`
	ClaudeModel        string `json:"claudeModel"`
	OpenAIModel        string `json:"openaiModel"`
	OllamaModel        string `json:"ollamaModel"`
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

func (a *App) GetSettingsData() (*SettingsData, error) {
	s, err := a.db.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	if s == nil {
		return nil, nil
	}
	return &SettingsData{
		ClaudeAPIKey:       string(s.ClaudeAPIKey),
		OpenAIAPIKey:       string(s.OpenAIAPIKey),
		ActiveLLM:          s.ActiveLLM,
		OllamaEndpoint:     s.OllamaEndpoint,
		StravaClientID:     string(s.StravaClientID),
		StravaClientSecret: string(s.StravaClientSecret),
		ClaudeModel:        s.ClaudeModel,
		OpenAIModel:        s.OpenAIModel,
		OllamaModel:        s.OllamaModel,
	}, nil
}

func (a *App) SaveSettingsData(data SettingsData) error {
	s := &storage.Settings{
		ClaudeAPIKey:       []byte(data.ClaudeAPIKey),
		OpenAIAPIKey:       []byte(data.OpenAIAPIKey),
		ActiveLLM:          data.ActiveLLM,
		OllamaEndpoint:     data.OllamaEndpoint,
		StravaClientID:     []byte(data.StravaClientID),
		StravaClientSecret: []byte(data.StravaClientSecret),
		ClaudeModel:        data.ClaudeModel,
		OpenAIModel:        data.OpenAIModel,
		OllamaModel:        data.OllamaModel,
	}
	if err := a.db.SaveSettings(s); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}
	if err := a.reloadLLMClient(); err != nil {
		return fmt.Errorf("reload LLM after save: %w", err)
	}
	return nil
}

func (a *App) IsFirstRun() (bool, error) {
	s, err := a.db.GetSettings()
	if err != nil {
		return false, fmt.Errorf("check first run: %w", err)
	}
	return s == nil, nil
}

func (a *App) GetStravaAuthStatus() (map[string]interface{}, error) {
	accessToken, _, _, err := a.db.GetTokens()
	if err != nil {
		return nil, fmt.Errorf("get strava auth status: %w", err)
	}
	return map[string]interface{}{
		"connected": accessToken != nil,
	}, nil
}

func (a *App) StartStravaAuth() error {
	s, err := a.db.GetSettings()
	if err != nil {
		return fmt.Errorf("get settings for strava auth: %w", err)
	}
	if s == nil {
		return errors.New("no settings configured; save settings first")
	}

	clientID := string(s.StravaClientID)
	clientSecret := string(s.StravaClientSecret)
	if clientID == "" || clientSecret == "" {
		return errors.New("strava client ID and secret must be configured in settings")
	}

	encKey := sha256.Sum256([]byte("coachlm-encryption-key"))
	oauthClient := strava.NewOAuthClient(clientID, clientSecret, "http://localhost:9876/callback", encKey[:])

	authURL := oauthClient.AuthURL()

	resultCh := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	listener, err := net.Listen("tcp", ":9876")
	if err != nil {
		return fmt.Errorf("start callback server: %w", err)
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing authorization code", http.StatusBadRequest)
			resultCh <- errors.New("no authorization code received from Strava")
			return
		}

		tokens, exchangeErr := oauthClient.Exchange(r.Context(), code)
		if exchangeErr != nil {
			http.Error(w, "token exchange failed", http.StatusInternalServerError)
			resultCh <- fmt.Errorf("exchange auth code: %w", exchangeErr)
			return
		}

		encAccess, encErr := oauthClient.EncryptToken(tokens.AccessToken)
		if encErr != nil {
			resultCh <- fmt.Errorf("encrypt access token: %w", encErr)
			return
		}
		encRefresh, encErr := oauthClient.EncryptToken(tokens.RefreshToken)
		if encErr != nil {
			resultCh <- fmt.Errorf("encrypt refresh token: %w", encErr)
			return
		}

		if saveErr := a.db.SaveTokens(encAccess, encRefresh, tokens.ExpiresAt); saveErr != nil {
			resultCh <- fmt.Errorf("save tokens: %w", saveErr)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h2>Authorization successful!</h2><p>You can close this tab.</p></body></html>")
		resultCh <- nil
	})

	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			resultCh <- fmt.Errorf("callback server: %w", serveErr)
		}
	}()

	wailsRuntime.BrowserOpenURL(a.ctx, authURL)

	select {
	case authErr := <-resultCh:
		_ = server.Close()
		return authErr
	case <-time.After(2 * time.Minute):
		_ = server.Close()
		return errors.New("strava authorization timed out after 2 minutes")
	}
}

func (a *App) DisconnectStrava() error {
	if err := a.db.DeleteTokens(); err != nil {
		return fmt.Errorf("disconnect strava: %w", err)
	}
	return nil
}

func (a *App) GetProfileData() (*ProfileData, error) {
	p, err := a.db.GetProfile()
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	if p == nil {
		return nil, nil
	}
	return &ProfileData{
		Age:                 p.Age,
		MaxHR:               p.MaxHR,
		ThresholdPaceSecs:   p.ThresholdPaceSecs,
		WeeklyMileageTarget: p.WeeklyMileageTarget,
		RaceGoals:           p.RaceGoals,
		InjuryHistory:       p.InjuryHistory,
	}, nil
}

func (a *App) SaveProfileData(data ProfileData) error {
	p := &storage.AthleteProfile{
		Age:                 data.Age,
		MaxHR:               data.MaxHR,
		ThresholdPaceSecs:   data.ThresholdPaceSecs,
		WeeklyMileageTarget: data.WeeklyMileageTarget,
		RaceGoals:           data.RaceGoals,
		InjuryHistory:       data.InjuryHistory,
	}
	if err := a.db.SaveProfile(p); err != nil {
		return fmt.Errorf("save profile: %w", err)
	}
	return nil
}

func (a *App) GetPinnedInsights() ([]InsightData, error) {
	insights, err := a.db.GetInsights()
	if err != nil {
		return nil, fmt.Errorf("get insights: %w", err)
	}
	result := make([]InsightData, 0, len(insights))
	for _, i := range insights {
		result = append(result, InsightData{
			ID:              i.ID,
			Content:         i.Content,
			SourceSessionID: i.SourceSessionID,
			CreatedAt:       i.CreatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

func (a *App) DeletePinnedInsight(id int64) error {
	if err := a.db.DeleteInsight(id); err != nil {
		return fmt.Errorf("delete insight: %w", err)
	}
	return nil
}

func (a *App) GetContextPreview() (string, error) {
	profile, _ := a.db.GetProfile()
	activities, _ := a.db.ListActivities(28, 0)
	insights, _ := a.db.GetInsights()

	prompt := coachctx.AssemblePrompt(coachctx.PromptInput{
		Profile:    profile,
		Activities: activities,
		Insights:   insights,
		Now:        time.Now(),
	}, coachctx.DefaultPromptConfig())

	return prompt, nil
}

func (a *App) SyncStravaActivities() error {
	s, err := a.db.GetSettings()
	if err != nil {
		return fmt.Errorf("get settings: %w", err)
	}
	if s == nil {
		return errors.New("no settings configured")
	}

	accessTokenEnc, refreshTokenEnc, expiresAt, err := a.db.GetTokens()
	if err != nil {
		return fmt.Errorf("get tokens: %w", err)
	}
	if accessTokenEnc == nil {
		return errors.New("strava not connected")
	}

	encKey := sha256.Sum256([]byte("coachlm-encryption-key"))
	oauthClient := strava.NewOAuthClient(
		string(s.StravaClientID), string(s.StravaClientSecret),
		"http://localhost:9876/callback", encKey[:],
	)

	accessToken, err := oauthClient.DecryptToken(accessTokenEnc)
	if err != nil {
		return fmt.Errorf("decrypt access token: %w", err)
	}

	if oauthClient.IsExpired(expiresAt) {
		refreshToken, err := oauthClient.DecryptToken(refreshTokenEnc)
		if err != nil {
			return fmt.Errorf("decrypt refresh token: %w", err)
		}
		tokens, err := oauthClient.Refresh(a.ctx, refreshToken)
		if err != nil {
			return fmt.Errorf("refresh token: %w", err)
		}
		encAccess, err := oauthClient.EncryptToken(tokens.AccessToken)
		if err != nil {
			return fmt.Errorf("encrypt new access token: %w", err)
		}
		encRefresh, err := oauthClient.EncryptToken(tokens.RefreshToken)
		if err != nil {
			return fmt.Errorf("encrypt new refresh token: %w", err)
		}
		if err := a.db.SaveTokens(encAccess, encRefresh, tokens.ExpiresAt); err != nil {
			return fmt.Errorf("save refreshed tokens: %w", err)
		}
		accessToken = tokens.AccessToken
	}

	wailsRuntime.EventsEmit(a.ctx, "strava:sync:start", nil)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	activities, err := strava.FetchAthleteActivities(a.ctx, httpClient, "https://www.strava.com/api/v3", accessToken)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "strava:sync:error", err.Error())
		return fmt.Errorf("fetch activities: %w", err)
	}

	total := len(activities)
	saved := 0
	for i, act := range activities {
		wailsRuntime.EventsEmit(a.ctx, "strava:sync:progress", map[string]int{
			"current": i + 1,
			"total":   total,
		})

		existing, err := a.db.GetActivityByStravaID(act.StravaID)
		if err != nil {
			continue
		}
		if existing != nil {
			continue
		}

		if err := a.db.SaveActivity(act); err != nil {
			continue
		}
		saved++
	}

	wailsRuntime.EventsEmit(a.ctx, "strava:sync:complete", map[string]int{
		"total": total,
		"saved": saved,
	})

	if preview, err := a.GetContextPreview(); err == nil {
		wailsRuntime.EventsEmit(a.ctx, "strava:sync:context-ready", preview)
	}

	return nil
}

func (a *App) GetOllamaModels(endpoint string) ([]string, error) {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	return llm.ListOllamaModels(a.ctx, endpoint)
}

func (a *App) reloadLLMClient() error {
	client, err := createLLMClient(a.db)
	if err != nil {
		return fmt.Errorf("reload LLM client: %w", err)
	}
	a.llmClient = client
	return nil
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
