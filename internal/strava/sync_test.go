package strava

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAthleteActivitiesSinglePage(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/athlete/activities" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode([]StravaActivitySummary{
			{
				ID:               100,
				Name:             "Morning Run",
				Type:             "Run",
				StartDate:        "2026-03-15T06:00:00Z",
				Distance:         10000.0,
				MovingTime:       3000,
				AverageSpeed:     3.33,
				AverageHeartrate: 150.0,
				MaxHeartrate:     175.0,
				AverageCadence:   180.0,
			},
			{
				ID:               101,
				Name:             "Afternoon Walk",
				Type:             "Walk",
				StartDate:        "2026-03-15T14:00:00Z",
				Distance:         3000.0,
				MovingTime:       1800,
				AverageSpeed:     1.67,
				AverageHeartrate: 100.0,
				MaxHeartrate:     120.0,
			},
		})
	}))
	defer mockStrava.Close()

	activities, _, err := FetchAthleteActivities(t.Context(), mockStrava.Client(), mockStrava.URL, "test-token")
	if err != nil {
		t.Fatalf("FetchAthleteActivities: %v", err)
	}

	if len(activities) != 2 {
		t.Fatalf("got %d activities, want 2", len(activities))
	}

	a := activities[0]
	if a.StravaID != 100 {
		t.Errorf("StravaID = %d, want 100", a.StravaID)
	}
	if a.Name != "Morning Run" {
		t.Errorf("Name = %q, want %q", a.Name, "Morning Run")
	}
	if a.ActivityType != "Run" {
		t.Errorf("ActivityType = %q, want %q", a.ActivityType, "Run")
	}
	if a.Distance != 10000.0 {
		t.Errorf("Distance = %f, want 10000.0", a.Distance)
	}
	if a.Source != "strava" {
		t.Errorf("Source = %q, want %q", a.Source, "strava")
	}
	if a.AvgPaceSecs != 300 {
		t.Errorf("AvgPaceSecs = %d, want 300 (1000/3.33)", a.AvgPaceSecs)
	}
}

func TestFetchAthleteActivitiesPagination(t *testing.T) {
	callCount := 0
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		page := r.URL.Query().Get("page")

		var activities []StravaActivitySummary
		switch page {
		case "1", "":
			for i := 0; i < defaultPerPage; i++ {
				activities = append(activities, StravaActivitySummary{
					ID:   int64(i + 1),
					Name: "Run",
					Type: "Run",
				})
			}
		case "2":
			activities = append(activities, StravaActivitySummary{
				ID:   int64(defaultPerPage + 1),
				Name: "Last Run",
				Type: "Run",
			})
		default:
			activities = []StravaActivitySummary{}
		}

		json.NewEncoder(w).Encode(activities)
	}))
	defer mockStrava.Close()

	activities, _, err := FetchAthleteActivities(t.Context(), mockStrava.Client(), mockStrava.URL, "token")
	if err != nil {
		t.Fatalf("FetchAthleteActivities: %v", err)
	}

	expected := defaultPerPage + 1
	if len(activities) != expected {
		t.Errorf("got %d activities, want %d", len(activities), expected)
	}

	if callCount != 2 {
		t.Errorf("API called %d times, want 2 (page 1 full + page 2 partial = stop)", callCount)
	}
}

func TestFetchAthleteActivitiesEmptyResponse(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]StravaActivitySummary{})
	}))
	defer mockStrava.Close()

	activities, _, err := FetchAthleteActivities(t.Context(), mockStrava.Client(), mockStrava.URL, "token")
	if err != nil {
		t.Fatalf("FetchAthleteActivities: %v", err)
	}

	if len(activities) != 0 {
		t.Errorf("got %d activities, want 0", len(activities))
	}
}

func TestFetchAthleteActivitiesAPIError(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer mockStrava.Close()

	_, _, err := FetchAthleteActivities(t.Context(), mockStrava.Client(), mockStrava.URL, "token")
	if err == nil {
		t.Error("expected error for non-200 response")
	}
}

func TestFetchAthleteActivitiesAuthHeader(t *testing.T) {
	var receivedAuth string
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		json.NewEncoder(w).Encode([]StravaActivitySummary{})
	}))
	defer mockStrava.Close()

	_, _, err := FetchAthleteActivities(t.Context(), mockStrava.Client(), mockStrava.URL, "my-secret-token")
	if err != nil {
		t.Fatalf("FetchAthleteActivities: %v", err)
	}

	if receivedAuth != "Bearer my-secret-token" {
		t.Errorf("Authorization = %q, want %q", receivedAuth, "Bearer my-secret-token")
	}
}
