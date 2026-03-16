package strava

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchStreamsSuccess(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activities/12345/streams" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Query().Get("keys") != "heartrate,velocity_smooth,cadence" {
			t.Errorf("keys = %q, want heartrate,velocity_smooth,cadence", r.URL.Query().Get("keys"))
		}
		if r.URL.Query().Get("key_type") != "time" {
			t.Errorf("key_type = %q, want time", r.URL.Query().Get("key_type"))
		}

		json.NewEncoder(w).Encode([]map[string]any{
			{"type": "heartrate", "data": []int{140, 145, 150, 155}},
			{"type": "velocity_smooth", "data": []float64{3.0, 3.5, 4.0, 0.0}},
			{"type": "cadence", "data": []int{170, 175, 180, 0}},
		})
	}))
	defer mockStrava.Close()

	wh := &WebhookHandler{
		stravaAPIBase: mockStrava.URL,
		httpClient:    mockStrava.Client(),
	}

	ss, err := wh.FetchStreams(t.Context(), "test-token", 12345)
	if err != nil {
		t.Fatalf("FetchStreams: %v", err)
	}

	if len(ss.HeartRate) != 4 {
		t.Fatalf("HeartRate len = %d, want 4", len(ss.HeartRate))
	}
	if ss.HeartRate[0] != 140 || ss.HeartRate[3] != 155 {
		t.Errorf("HeartRate = %v, want [140 145 150 155]", ss.HeartRate)
	}

	if len(ss.Pace) != 4 {
		t.Fatalf("Pace len = %d, want 4", len(ss.Pace))
	}
	expectedPace0 := 1000.0 / 3.0
	if math.Abs(ss.Pace[0]-expectedPace0) > 0.01 {
		t.Errorf("Pace[0] = %f, want %f", ss.Pace[0], expectedPace0)
	}
	if ss.Pace[3] != 0 {
		t.Errorf("Pace[3] = %f, want 0 (zero velocity)", ss.Pace[3])
	}

	if len(ss.Cadence) != 4 {
		t.Fatalf("Cadence len = %d, want 4", len(ss.Cadence))
	}
	if ss.Cadence[0] != 170 || ss.Cadence[2] != 180 {
		t.Errorf("Cadence = %v, want [170 175 180 0]", ss.Cadence)
	}
}

func TestFetchStreamsMissingStreams(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{"type": "heartrate", "data": []int{140, 145}},
		})
	}))
	defer mockStrava.Close()

	wh := &WebhookHandler{
		stravaAPIBase: mockStrava.URL,
		httpClient:    mockStrava.Client(),
	}

	ss, err := wh.FetchStreams(t.Context(), "token", 1)
	if err != nil {
		t.Fatalf("FetchStreams: %v", err)
	}

	if ss.HeartRate == nil {
		t.Error("HeartRate should not be nil")
	}
	if ss.Pace != nil {
		t.Errorf("Pace should be nil when velocity_smooth missing, got %v", ss.Pace)
	}
	if ss.Cadence != nil {
		t.Errorf("Cadence should be nil when cadence missing, got %v", ss.Cadence)
	}
}

func TestFetchStreamsEmptyResponse(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer mockStrava.Close()

	wh := &WebhookHandler{
		stravaAPIBase: mockStrava.URL,
		httpClient:    mockStrava.Client(),
	}

	ss, err := wh.FetchStreams(t.Context(), "token", 1)
	if err != nil {
		t.Fatalf("FetchStreams: %v", err)
	}

	if ss.HeartRate != nil || ss.Pace != nil || ss.Cadence != nil {
		t.Errorf("all streams should be nil for empty response, got HR=%v Pace=%v Cadence=%v",
			ss.HeartRate, ss.Pace, ss.Cadence)
	}
}

func TestFetchStreamsAPIError(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer mockStrava.Close()

	wh := &WebhookHandler{
		stravaAPIBase: mockStrava.URL,
		httpClient:    mockStrava.Client(),
	}

	_, err := wh.FetchStreams(t.Context(), "token", 1)
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestFetchStreamsUnauthorized(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer mockStrava.Close()

	wh := &WebhookHandler{
		stravaAPIBase: mockStrava.URL,
		httpClient:    mockStrava.Client(),
	}

	_, err := wh.FetchStreams(t.Context(), "bad-token", 1)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestVelocityToPace(t *testing.T) {
	tests := []struct {
		name       string
		velocities []float64
		wantLen    int
		checkIdx   int
		wantVal    float64
	}{
		{"normal speed", []float64{3.0}, 1, 0, 1000.0 / 3.0},
		{"zero velocity", []float64{0.0}, 1, 0, 0},
		{"negative velocity", []float64{-1.0}, 1, 0, 0},
		{"empty", []float64{}, 0, -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paces := velocityToPace(tt.velocities)
			if len(paces) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(paces), tt.wantLen)
			}
			if tt.checkIdx >= 0 && math.Abs(paces[tt.checkIdx]-tt.wantVal) > 0.01 {
				t.Errorf("pace[%d] = %f, want %f", tt.checkIdx, paces[tt.checkIdx], tt.wantVal)
			}
		})
	}
}

func TestFetchStreamsContextCancellation(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer mockStrava.Close()

	wh := &WebhookHandler{
		stravaAPIBase: mockStrava.URL,
		httpClient:    mockStrava.Client(),
	}

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	_, err := wh.FetchStreams(ctx, "token", 1)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
