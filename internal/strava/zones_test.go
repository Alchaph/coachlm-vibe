package strava

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAthleteZones(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/athlete/zones" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(AthleteZonesResponse{
			HeartRate: HeartRateZones{
				CustomZones: true,
				Zones: []ZoneRange{
					{Min: 0, Max: 115},
					{Min: 115, Max: 152},
					{Min: 152, Max: 171},
					{Min: 171, Max: 190},
					{Min: 190, Max: -1},
				},
			},
		})
	}))
	defer mockStrava.Close()

	zones, err := FetchAthleteZones(t.Context(), mockStrava.Client(), mockStrava.URL, "test-token")
	if err != nil {
		t.Fatalf("FetchAthleteZones: %v", err)
	}

	if !zones.CustomZones {
		t.Error("expected CustomZones = true")
	}
	if len(zones.Zones) != 5 {
		t.Fatalf("got %d zones, want 5", len(zones.Zones))
	}
	if zones.Zones[0].Min != 0 || zones.Zones[0].Max != 115 {
		t.Errorf("zone 0 = {%d, %d}, want {0, 115}", zones.Zones[0].Min, zones.Zones[0].Max)
	}
	if zones.Zones[4].Min != 190 || zones.Zones[4].Max != -1 {
		t.Errorf("zone 4 = {%d, %d}, want {190, -1}", zones.Zones[4].Min, zones.Zones[4].Max)
	}
}

func TestFetchAthleteZonesDefaultZones(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AthleteZonesResponse{
			HeartRate: HeartRateZones{
				CustomZones: false,
				Zones: []ZoneRange{
					{Min: 0, Max: 123},
					{Min: 123, Max: 153},
					{Min: 153, Max: 169},
					{Min: 169, Max: 184},
					{Min: 184, Max: -1},
				},
			},
		})
	}))
	defer mockStrava.Close()

	zones, err := FetchAthleteZones(t.Context(), mockStrava.Client(), mockStrava.URL, "token")
	if err != nil {
		t.Fatalf("FetchAthleteZones: %v", err)
	}

	if zones.CustomZones {
		t.Error("expected CustomZones = false for default zones")
	}
	if len(zones.Zones) != 5 {
		t.Fatalf("got %d zones, want 5", len(zones.Zones))
	}
}

func TestFetchAthleteZonesAPIError(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer mockStrava.Close()

	_, err := FetchAthleteZones(t.Context(), mockStrava.Client(), mockStrava.URL, "token")
	if err == nil {
		t.Error("expected error for non-200 response")
	}
}

func TestFetchAthleteZonesEmptyZones(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AthleteZonesResponse{
			HeartRate: HeartRateZones{
				CustomZones: false,
				Zones:       []ZoneRange{},
			},
		})
	}))
	defer mockStrava.Close()

	zones, err := FetchAthleteZones(t.Context(), mockStrava.Client(), mockStrava.URL, "token")
	if err != nil {
		t.Fatalf("FetchAthleteZones: %v", err)
	}

	if len(zones.Zones) != 0 {
		t.Errorf("got %d zones, want 0", len(zones.Zones))
	}
}

func TestFetchAthleteZonesNilHTTPClient(t *testing.T) {
	mockStrava := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AthleteZonesResponse{
			HeartRate: HeartRateZones{
				Zones: []ZoneRange{{Min: 0, Max: 120}},
			},
		})
	}))
	defer mockStrava.Close()

	zones, err := FetchAthleteZones(t.Context(), nil, mockStrava.URL, "token")
	if err != nil {
		t.Fatalf("FetchAthleteZones with nil client: %v", err)
	}
	if len(zones.Zones) != 1 {
		t.Errorf("got %d zones, want 1", len(zones.Zones))
	}
}
