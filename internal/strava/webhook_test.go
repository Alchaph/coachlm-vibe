package strava

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"coachlm/internal/storage"
)

func newTestDB(t *testing.T) *storage.DB {
	t.Helper()
	db, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("newTestDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestHandleChallengeSuccess(t *testing.T) {
	wh := NewWebhookHandler(nil, "my-secret", nil)

	req := httptest.NewRequest(http.MethodGet,
		"/webhook?hub.verify_token=my-secret&hub.challenge=challenge-abc&hub.mode=subscribe", nil)
	rec := httptest.NewRecorder()

	wh.HandleChallenge(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["hub.challenge"] != "challenge-abc" {
		t.Errorf("challenge = %q, want %q", body["hub.challenge"], "challenge-abc")
	}
}

func TestHandleChallengeWrongToken(t *testing.T) {
	wh := NewWebhookHandler(nil, "my-secret", nil)

	req := httptest.NewRequest(http.MethodGet,
		"/webhook?hub.verify_token=wrong-token&hub.challenge=test", nil)
	rec := httptest.NewRecorder()

	wh.HandleChallenge(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestHandleEventCreatesActivity(t *testing.T) {
	db := newTestDB(t)

	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(stravaActivityResponse{
			ID:               77777,
			Name:             "Evening Run",
			Type:             "Run",
			StartDate:        "2026-03-15T18:00:00Z",
			Distance:         8000.0,
			MovingTime:       2400,
			AverageSpeed:     3.33,
			AverageHeartrate: 155.0,
			MaxHeartrate:     180.0,
			AverageCadence:   178.0,
		})
	}))
	defer mockStrava.Close()

	wh := NewWebhookHandler(db, "token", func() (string, error) {
		return "test-token", nil
	})
	wh.stravaAPIBase = mockStrava.URL

	event := WebhookEvent{
		ObjectType: "activity",
		ObjectID:   77777,
		AspectType: "create",
		OwnerID:    1,
	}
	body, _ := json.Marshal(event)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(body)))
	rec := httptest.NewRecorder()

	wh.HandleEvent(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	time.Sleep(500 * time.Millisecond)

	activity, err := db.GetActivityByStravaID(77777)
	if err != nil {
		t.Fatalf("GetActivityByStravaID: %v", err)
	}
	if activity == nil {
		t.Fatal("expected activity to be saved, got nil")
	}
	if activity.Name != "Evening Run" {
		t.Errorf("Name = %q, want %q", activity.Name, "Evening Run")
	}
	if activity.Distance != 8000.0 {
		t.Errorf("Distance = %f, want 8000.0", activity.Distance)
	}
}

func TestHandleEventDuplicateSkipped(t *testing.T) {
	db := newTestDB(t)

	existing := &storage.Activity{
		StravaID:     88888,
		Name:         "Already Saved",
		ActivityType: "Run",
		StartDate:    time.Date(2026, 3, 14, 7, 0, 0, 0, time.UTC),
		Source:       "strava",
	}
	if err := db.SaveActivity(existing); err != nil {
		t.Fatalf("SaveActivity: %v", err)
	}

	fetchCalled := false
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCalled = true
		json.NewEncoder(w).Encode(stravaActivityResponse{ID: 88888, Name: "Should Not Fetch"})
	}))
	defer mockStrava.Close()

	wh := NewWebhookHandler(db, "token", func() (string, error) {
		return "test-token", nil
	})
	wh.stravaAPIBase = mockStrava.URL

	event := WebhookEvent{
		ObjectType: "activity",
		ObjectID:   88888,
		AspectType: "create",
		OwnerID:    1,
	}
	body, _ := json.Marshal(event)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(body)))
	rec := httptest.NewRecorder()

	wh.HandleEvent(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	time.Sleep(200 * time.Millisecond)

	if fetchCalled {
		t.Error("Strava API should not be called for duplicate activity")
	}
}

func TestHandleEventMalformedBody(t *testing.T) {
	wh := NewWebhookHandler(nil, "token", nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("not-json"))
	rec := httptest.NewRecorder()

	wh.HandleEvent(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (must respond quickly even on bad input)", rec.Code)
	}
}

func TestHandleEventNonActivityIgnored(t *testing.T) {
	db := newTestDB(t)

	fetchCalled := false
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCalled = true
		fmt.Fprintln(w, "{}")
	}))
	defer mockStrava.Close()

	wh := NewWebhookHandler(db, "token", func() (string, error) {
		return "test-token", nil
	})
	wh.stravaAPIBase = mockStrava.URL

	event := WebhookEvent{
		ObjectType: "athlete",
		ObjectID:   123,
		AspectType: "update",
	}
	body, _ := json.Marshal(event)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(body)))
	rec := httptest.NewRecorder()

	wh.HandleEvent(rec, req)
	time.Sleep(200 * time.Millisecond)

	if fetchCalled {
		t.Error("Strava API should not be called for non-activity events")
	}
}

func TestFetchActivity(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activities/55555" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer my-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(stravaActivityResponse{
			ID:               55555,
			Name:             "Tempo Run",
			Type:             "Run",
			StartDate:        "2026-03-16T06:30:00Z",
			Distance:         12000.0,
			MovingTime:       3200,
			AverageSpeed:     3.75,
			AverageHeartrate: 165.0,
			MaxHeartrate:     185.0,
			AverageCadence:   182.0,
		})
	}))
	defer mockStrava.Close()

	wh := &WebhookHandler{
		stravaAPIBase: mockStrava.URL,
		httpClient:    mockStrava.Client(),
	}

	activity, err := wh.FetchActivity(t.Context(), "my-token", 55555)
	if err != nil {
		t.Fatalf("FetchActivity: %v", err)
	}
	if activity.StravaID != 55555 {
		t.Errorf("StravaID = %d, want 55555", activity.StravaID)
	}
	if activity.Name != "Tempo Run" {
		t.Errorf("Name = %q, want %q", activity.Name, "Tempo Run")
	}
	if activity.Source != "strava" {
		t.Errorf("Source = %q, want %q", activity.Source, "strava")
	}
	if activity.AvgPaceSecs != 266 {
		t.Errorf("AvgPaceSecs = %d, want 266 (1000/3.75)", activity.AvgPaceSecs)
	}
}

func TestFetchActivityAPIError(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer mockStrava.Close()

	wh := &WebhookHandler{
		stravaAPIBase: mockStrava.URL,
		httpClient:    mockStrava.Client(),
	}

	_, err := wh.FetchActivity(t.Context(), "token", 123)
	if err == nil {
		t.Error("expected error for non-200 response")
	}
}
