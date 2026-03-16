package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"coachlm/internal/storage"
)

const defaultStravaAPIBase = "https://www.strava.com/api/v3"

type WebhookEvent struct {
	ObjectType     string `json:"object_type"`
	ObjectID       int64  `json:"object_id"`
	AspectType     string `json:"aspect_type"`
	OwnerID        int64  `json:"owner_id"`
	SubscriptionID int64  `json:"subscription_id"`
	EventTime      int64  `json:"event_time"`
}

type stravaActivityResponse struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Type             string  `json:"type"`
	StartDate        string  `json:"start_date"`
	Distance         float64 `json:"distance"`
	MovingTime       int     `json:"moving_time"`
	AverageSpeed     float64 `json:"average_speed"`
	AverageHeartrate float64 `json:"average_heartrate"`
	MaxHeartrate     float64 `json:"max_heartrate"`
	AverageCadence   float64 `json:"average_cadence"`
}

type TokenProvider func() (string, error)

type WebhookHandler struct {
	DB            *storage.DB
	VerifyToken   string
	TokenProvider TokenProvider
	stravaAPIBase string
	httpClient    *http.Client
}

func NewWebhookHandler(db *storage.DB, verifyToken string, tokenProvider TokenProvider) *WebhookHandler {
	return &WebhookHandler{
		DB:            db,
		VerifyToken:   verifyToken,
		TokenProvider: tokenProvider,
		stravaAPIBase: defaultStravaAPIBase,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (wh *WebhookHandler) HandleChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("hub.verify_token")
	if token != wh.VerifyToken {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	challenge := r.URL.Query().Get("hub.challenge")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"hub.challenge": challenge,
	})
}

func (wh *WebhookHandler) HandleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event WebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)

	go wh.processEvent(event)
}

func (wh *WebhookHandler) processEvent(event WebhookEvent) {
	if event.ObjectType != "activity" {
		return
	}
	if event.AspectType != "create" && event.AspectType != "update" {
		return
	}

	existing, err := wh.DB.GetActivityByStravaID(event.ObjectID)
	if err != nil {
		log.Printf("webhook: dedup check failed for strava_id %d: %v", event.ObjectID, err)
		return
	}
	if existing != nil {
		return
	}

	accessToken, err := wh.TokenProvider()
	if err != nil {
		log.Printf("webhook: get access token: %v", err)
		return
	}

	activity, err := wh.FetchActivity(context.Background(), accessToken, event.ObjectID)
	if err != nil {
		log.Printf("webhook: fetch activity %d: %v", event.ObjectID, err)
		return
	}

	if err := wh.DB.SaveActivity(activity); err != nil {
		log.Printf("webhook: save activity %d: %v", event.ObjectID, err)
	}
}

func (wh *WebhookHandler) FetchActivity(ctx context.Context, accessToken string, activityID int64) (*storage.Activity, error) {
	url := fmt.Sprintf("%s/activities/%d", wh.stravaAPIBase, activityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := wh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch activity: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava API error: HTTP %d", resp.StatusCode)
	}

	var raw stravaActivityResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode activity: %w", err)
	}

	startDate, _ := time.Parse(time.RFC3339, raw.StartDate)

	var avgPaceSecs int
	if raw.AverageSpeed > 0 {
		avgPaceSecs = int(1000.0 / raw.AverageSpeed)
	}

	return &storage.Activity{
		StravaID:     raw.ID,
		Name:         raw.Name,
		ActivityType: raw.Type,
		StartDate:    startDate,
		Distance:     raw.Distance,
		DurationSecs: raw.MovingTime,
		AvgPaceSecs:  avgPaceSecs,
		AvgHR:        int(raw.AverageHeartrate),
		MaxHR:        int(raw.MaxHeartrate),
		AvgCadence:   raw.AverageCadence,
		Source:       "strava",
	}, nil
}
