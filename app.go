package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"coachlm/internal/cloudsync"
	coachctx "coachlm/internal/context"
	"coachlm/internal/exportimport"
	"coachlm/internal/fit"
	"coachlm/internal/llm"
	"coachlm/internal/storage"
	"coachlm/internal/strava"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	db          *storage.DB
	llmClient   llm.LLM
	sessionID   string
	syncManager *cloudsync.Manager
}

type ProfileData struct {
	Age                 int     `json:"age"`
	MaxHR               int     `json:"maxHR"`
	ThresholdPaceSecs   int     `json:"thresholdPaceSecs"`
	WeeklyMileageTarget float64 `json:"weeklyMileageTarget"`
	RaceGoals           string  `json:"raceGoals"`
	InjuryHistory       string  `json:"injuryHistory"`
	ExperienceLevel     string  `json:"experienceLevel"`
	TrainingDaysPerWeek int     `json:"trainingDaysPerWeek"`
	RestingHR           int     `json:"restingHR"`
	PreferredTerrain    string  `json:"preferredTerrain"`
	HeartRateZones      string  `json:"heartRateZones"`
}

type InsightData struct {
	ID              int64  `json:"id"`
	Content         string `json:"content"`
	SourceSessionID string `json:"sourceSessionId"`
	CreatedAt       string `json:"createdAt"`
}

type SettingsData struct {
	OllamaEndpoint     string `json:"ollamaEndpoint"`
	OllamaModel        string `json:"ollamaModel"`
	CustomSystemPrompt string `json:"customSystemPrompt"`
}

func NewApp(db *storage.DB, llmClient llm.LLM) *App {
	return &App{db: db, llmClient: llmClient}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if err := a.initSyncManagerFromSettings(); err != nil {
		fmt.Printf("Warning: cloud sync init failed: %v\n", err)
	}
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
	settings, _ := a.db.GetSettings()

	customPrompt := ""
	if settings != nil {
		customPrompt = settings.CustomSystemPrompt
	}

	systemPrompt := coachctx.AssemblePrompt(coachctx.PromptInput{
		Profile:      profile,
		Activities:   activities,
		Insights:     insights,
		CustomPrompt: customPrompt,
		Now:          time.Now(),
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

type StatsData struct {
	TotalCount      int     `json:"totalCount"`
	TotalDistanceKm float64 `json:"totalDistanceKm"`
	EarliestDate    string  `json:"earliestDate"`
	LatestDate      string  `json:"latestDate"`
}

func (a *App) GetActivityStats() (*StatsData, error) {
	stats, err := a.db.GetActivityStats()
	if err != nil {
		return nil, fmt.Errorf("get activity stats: %w", err)
	}
	if stats == nil {
		return &StatsData{}, nil
	}
	return &StatsData{
		TotalCount:      stats.TotalCount,
		TotalDistanceKm: stats.TotalDistanceKm / 1000.0,
		EarliestDate:    stats.EarliestDate,
		LatestDate:      stats.LatestDate,
	}, nil
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
		OllamaEndpoint:     s.OllamaEndpoint,
		OllamaModel:        s.OllamaModel,
		CustomSystemPrompt: s.CustomSystemPrompt,
	}, nil
}

func (a *App) SaveSettingsData(data SettingsData) error {
	existing, err := a.db.GetSettings()
	if err != nil {
		return fmt.Errorf("get existing settings: %w", err)
	}

	s := &storage.Settings{
		ActiveLLM:          "local",
		OllamaEndpoint:     data.OllamaEndpoint,
		OllamaModel:        data.OllamaModel,
		CustomSystemPrompt: data.CustomSystemPrompt,
	}

	if existing != nil {
		s.CloudProvider = existing.CloudProvider
		s.CloudEndpoint = existing.CloudEndpoint
		s.CloudBucket = existing.CloudBucket
		s.CloudAccessKey = existing.CloudAccessKey
		s.CloudSecretKey = existing.CloudSecretKey
		s.GDriveAccessToken = existing.GDriveAccessToken
		s.GDriveRefreshToken = existing.GDriveRefreshToken
		s.GDriveTokenExpiry = existing.GDriveTokenExpiry
		s.GDriveClientID = existing.GDriveClientID
	}

	if err := a.db.SaveSettings(s); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}
	if err := a.reloadLLMClient(); err != nil {
		// Log but don't fail — settings were saved successfully.
		// The LLM may fail later at chat time (e.g., free backend with no API key).
		fmt.Printf("Warning: LLM reload after settings save: %v\n", err)
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

func resolveStravaCredentials() (clientID, clientSecret string, ok bool) {
	id, secret := strava.BuiltinCredentials()
	if id == "" {
		id = os.Getenv("STRAVA_CLIENT_ID")
	}
	if secret == "" {
		secret = os.Getenv("STRAVA_CLIENT_SECRET")
	}
	return id, secret, id != "" && secret != ""
}

func (a *App) GetStravaCredentialsAvailable() bool {
	_, _, ok := resolveStravaCredentials()
	return ok
}

func (a *App) StartStravaAuth() error {
	clientID, clientSecret, ok := resolveStravaCredentials()
	if !ok {
		return errors.New("strava: no client credentials available in this build")
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
		ExperienceLevel:     p.ExperienceLevel,
		TrainingDaysPerWeek: p.TrainingDaysPerWeek,
		RestingHR:           p.RestingHR,
		PreferredTerrain:    p.PreferredTerrain,
		HeartRateZones:      p.HeartRateZones,
	}, nil
}

func (a *App) SaveProfileData(data ProfileData) error {
	existing, _ := a.db.GetProfile()
	heartRateZones := ""
	if existing != nil {
		heartRateZones = existing.HeartRateZones
	}

	p := &storage.AthleteProfile{
		Age:                 data.Age,
		MaxHR:               data.MaxHR,
		ThresholdPaceSecs:   data.ThresholdPaceSecs,
		WeeklyMileageTarget: data.WeeklyMileageTarget,
		RaceGoals:           data.RaceGoals,
		InjuryHistory:       data.InjuryHistory,
		ExperienceLevel:     data.ExperienceLevel,
		TrainingDaysPerWeek: data.TrainingDaysPerWeek,
		RestingHR:           data.RestingHR,
		PreferredTerrain:    data.PreferredTerrain,
		HeartRateZones:      heartRateZones,
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
	settings, _ := a.db.GetSettings()

	customPrompt := ""
	if settings != nil {
		customPrompt = settings.CustomSystemPrompt
	}

	prompt := coachctx.AssemblePrompt(coachctx.PromptInput{
		Profile:      profile,
		Activities:   activities,
		Insights:     insights,
		CustomPrompt: customPrompt,
		Now:          time.Now(),
	}, coachctx.DefaultPromptConfig())

	return prompt, nil
}

func (a *App) SyncStravaActivities() error {
	clientID, clientSecret, ok := resolveStravaCredentials()
	if !ok {
		return errors.New("strava: no client credentials available in this build")
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
		clientID, clientSecret,
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

	if hrZones, err := strava.FetchAthleteZones(a.ctx, httpClient, "https://www.strava.com/api/v3", accessToken); err == nil && len(hrZones.Zones) > 0 {
		zonesJSON, err := json.Marshal(hrZones.Zones)
		if err == nil {
			profile, _ := a.db.GetProfile()
			if profile != nil {
				profile.HeartRateZones = string(zonesJSON)
				_ = a.db.SaveProfile(profile)
			}
		}
	}

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

func (a *App) ExportContext(filePath string) error {
	return exportimport.Export(a.db, filePath)
}

func (a *App) ImportContext(filePath string, replaceAll bool) error {
	err := exportimport.Import(a.db, filePath, replaceAll)
	if err == nil && a.syncManager != nil {
		a.syncManager.TriggerSync()
	}
	return err
}

func (a *App) ConnectS3(endpoint, bucket, accessKey, secretKey string) error {
	provider, err := cloudsync.NewS3(cloudsync.S3Config{
		Endpoint:  endpoint,
		Bucket:    bucket,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if err != nil {
		return fmt.Errorf("validate s3 config: %w", err)
	}

	encKey := sha256.Sum256([]byte("coachlm-encryption-key"))

	encAccessKey, err := strava.Encrypt([]byte(accessKey), encKey[:])
	if err != nil {
		return fmt.Errorf("encrypt access key: %w", err)
	}
	encSecretKey, err := strava.Encrypt([]byte(secretKey), encKey[:])
	if err != nil {
		return fmt.Errorf("encrypt secret key: %w", err)
	}

	settings, err := a.db.GetSettings()
	if err != nil {
		return fmt.Errorf("get settings: %w", err)
	}
	if settings == nil {
		return errors.New("settings not initialized")
	}

	settings.CloudProvider = "s3"
	settings.CloudEndpoint = endpoint
	settings.CloudBucket = bucket
	settings.CloudAccessKey = encAccessKey
	settings.CloudSecretKey = encSecretKey
	settings.GDriveAccessToken = nil
	settings.GDriveRefreshToken = nil
	settings.GDriveClientID = ""

	if err := a.db.SaveSettings(settings); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}

	mgr, err := a.createSyncManager(provider)
	if err != nil {
		return fmt.Errorf("init sync manager: %w", err)
	}

	if a.syncManager != nil {
		a.syncManager.Stop()
	}
	a.syncManager = mgr

	return mgr.SyncNow()
}

func (a *App) ConnectGoogleDrive() error {
	port, err := randomPort()
	if err != nil {
		return fmt.Errorf("find port: %w", err)
	}

	clientID := "YOUR_GOOGLE_CLIENT_ID"
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	codeVerifier, codeChallenge := generatePKCE()

	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&code_challenge=%s&code_challenge_method=S256&access_type=offline&prompt=consent",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape("https://www.googleapis.com/auth/drive.file"),
		url.QueryEscape(codeChallenge),
	)

	resultCh := make(chan gdriveAuthResult, 1)
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("start gdrive callback server: %w", err)
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing authorization code", http.StatusBadRequest)
			resultCh <- gdriveAuthResult{err: errors.New("no authorization code from Google")}
			return
		}

		tokens, exchangeErr := exchangeGoogleCode(r.Context(), clientID, code, redirectURI, codeVerifier)
		if exchangeErr != nil {
			http.Error(w, "token exchange failed", http.StatusInternalServerError)
			resultCh <- gdriveAuthResult{err: exchangeErr}
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h2>Google Drive connected!</h2><p>You can close this tab.</p></body></html>")
		resultCh <- gdriveAuthResult{tokens: tokens}
	})

	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			resultCh <- gdriveAuthResult{err: fmt.Errorf("gdrive callback server: %w", serveErr)}
		}
	}()

	wailsRuntime.BrowserOpenURL(a.ctx, authURL)

	var result gdriveAuthResult
	select {
	case result = <-resultCh:
		_ = server.Close()
	case <-time.After(2 * time.Minute):
		_ = server.Close()
		return errors.New("google drive authorization timed out after 2 minutes")
	}

	if result.err != nil {
		return result.err
	}

	encKey := sha256.Sum256([]byte("coachlm-encryption-key"))

	encAccessToken, err := strava.Encrypt([]byte(result.tokens.AccessToken), encKey[:])
	if err != nil {
		return fmt.Errorf("encrypt gdrive access token: %w", err)
	}
	encRefreshToken, err := strava.Encrypt([]byte(result.tokens.RefreshToken), encKey[:])
	if err != nil {
		return fmt.Errorf("encrypt gdrive refresh token: %w", err)
	}

	settings, err := a.db.GetSettings()
	if err != nil {
		return fmt.Errorf("get settings: %w", err)
	}
	if settings == nil {
		return errors.New("settings not initialized")
	}

	settings.CloudProvider = "gdrive"
	settings.CloudEndpoint = ""
	settings.CloudBucket = ""
	settings.CloudAccessKey = nil
	settings.CloudSecretKey = nil
	settings.GDriveAccessToken = encAccessToken
	settings.GDriveRefreshToken = encRefreshToken
	settings.GDriveTokenExpiry = result.tokens.Expiry
	settings.GDriveClientID = clientID

	if err := a.db.SaveSettings(settings); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}

	provider, err := cloudsync.NewGDrive(cloudsync.GDriveConfig{
		AccessToken: result.tokens.AccessToken,
		ClientID:    clientID,
	})
	if err != nil {
		return fmt.Errorf("create gdrive provider: %w", err)
	}

	mgr, err := a.createSyncManager(provider)
	if err != nil {
		return fmt.Errorf("init sync manager: %w", err)
	}

	if a.syncManager != nil {
		a.syncManager.Stop()
	}
	a.syncManager = mgr

	return mgr.SyncNow()
}

func (a *App) DisconnectCloud() error {
	if a.syncManager != nil {
		a.syncManager.Stop()
		a.syncManager = nil
	}

	settings, err := a.db.GetSettings()
	if err != nil {
		return fmt.Errorf("get settings: %w", err)
	}
	if settings == nil {
		return nil
	}

	settings.CloudProvider = ""
	settings.CloudEndpoint = ""
	settings.CloudBucket = ""
	settings.CloudAccessKey = nil
	settings.CloudSecretKey = nil
	settings.GDriveAccessToken = nil
	settings.GDriveRefreshToken = nil
	settings.GDriveClientID = ""

	if err := a.db.SaveSettings(settings); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}

	return a.db.DeleteCloudSyncState()
}

func (a *App) SyncNow() error {
	if a.syncManager == nil {
		return errors.New("cloud sync is not enabled")
	}
	return a.syncManager.SyncNow()
}

func (a *App) GetSyncStatus() cloudsync.SyncStatus {
	if a.syncManager == nil {
		return cloudsync.SyncStatus{Enabled: false}
	}
	return a.syncManager.GetStatus()
}

func (a *App) ExportChatSessions() ([]byte, error) {
	return exportimport.ExportChatData(a.db)
}

func (a *App) ImportChatSessions(data []byte, replaceAll bool) error {
	err := exportimport.ImportChatData(a.db, data, replaceAll)
	if err == nil && a.syncManager != nil {
		a.syncManager.TriggerChatSync()
	}
	return err
}

func (a *App) createSyncManager(provider cloudsync.CloudProvider) (*cloudsync.Manager, error) {
	return cloudsync.NewManager(cloudsync.ManagerConfig{
		Provider: provider,
		ExportContext: func() ([]byte, error) {
			return exportimport.ExportData(a.db)
		},
		ImportContext: func(data []byte, replaceAll bool) error {
			return exportimport.ImportData(a.db, data, replaceAll)
		},
		ExportChat: func() ([]byte, error) {
			return exportimport.ExportChatData(a.db)
		},
		ImportChat: func(data []byte, replaceAll bool) error {
			return exportimport.ImportChatData(a.db, data, replaceAll)
		},
		StateStore: cloudsync.NewStateAdapter(a.db),
	})
}

func (a *App) initSyncManagerFromSettings() error {
	settings, err := a.db.GetSettings()
	if err != nil || settings == nil || settings.CloudProvider == "" {
		return nil
	}

	encKey := sha256.Sum256([]byte("coachlm-encryption-key"))

	var provider cloudsync.CloudProvider
	switch settings.CloudProvider {
	case "s3":
		accessKey, err := strava.Decrypt(settings.CloudAccessKey, encKey[:])
		if err != nil {
			return fmt.Errorf("decrypt s3 access key: %w", err)
		}
		secretKey, err := strava.Decrypt(settings.CloudSecretKey, encKey[:])
		if err != nil {
			return fmt.Errorf("decrypt s3 secret key: %w", err)
		}
		provider, err = cloudsync.NewS3(cloudsync.S3Config{
			Endpoint:  settings.CloudEndpoint,
			Bucket:    settings.CloudBucket,
			AccessKey: string(accessKey),
			SecretKey: string(secretKey),
		})
		if err != nil {
			return fmt.Errorf("create s3 provider: %w", err)
		}
	case "gdrive":
		accessToken, err := strava.Decrypt(settings.GDriveAccessToken, encKey[:])
		if err != nil {
			return fmt.Errorf("decrypt gdrive access token: %w", err)
		}
		provider, err = cloudsync.NewGDrive(cloudsync.GDriveConfig{
			AccessToken: string(accessToken),
			ClientID:    settings.GDriveClientID,
		})
		if err != nil {
			return fmt.Errorf("create gdrive provider: %w", err)
		}
	default:
		return nil
	}

	mgr, err := a.createSyncManager(provider)
	if err != nil {
		return fmt.Errorf("init sync manager: %w", err)
	}
	a.syncManager = mgr
	return nil
}

type gdriveAuthResult struct {
	tokens *gdriveTokens
	err    error
}

type gdriveTokens struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

func exchangeGoogleCode(ctx context.Context, clientID, code, redirectURI, codeVerifier string) (*gdriveTokens, error) {
	data := url.Values{
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
		"code_verifier": {codeVerifier},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("token exchange: status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}

	return &gdriveTokens{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}, nil
}

func randomPort() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 49152, nil
}

func generatePKCE() (verifier, challenge string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	verifier = strings.TrimRight(
		strings.NewReplacer("+", "-", "/", "_").Replace(
			fmt.Sprintf("%x", b),
		), "=",
	)
	h := sha256.Sum256([]byte(verifier))
	challenge = strings.TrimRight(
		strings.NewReplacer("+", "-", "/", "_").Replace(
			fmt.Sprintf("%x", h),
		), "=",
	)
	return verifier, challenge
}
