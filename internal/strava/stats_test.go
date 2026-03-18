package strava

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAthleteStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/athletes/12345/stats" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(AthleteStats{
			RecentRunTotals: ActivityTotals{
				Count:         5,
				Distance:      45000.0,
				MovingTime:    14400,
				ElevationGain: 300.0,
			},
			YTDRunTotals: ActivityTotals{
				Count:         42,
				Distance:      400000.0,
				MovingTime:    120000,
				ElevationGain: 2500.0,
			},
			AllRunTotals: ActivityTotals{
				Count:         500,
				Distance:      5000000.0,
				MovingTime:    1500000,
				ElevationGain: 30000.0,
			},
		})
	}))
	defer server.Close()

	stats, err := FetchAthleteStats(t.Context(), server.Client(), server.URL, "test-token", 12345)
	if err != nil {
		t.Fatalf("FetchAthleteStats: %v", err)
	}

	if stats.RecentRunTotals.Count != 5 {
		t.Errorf("RecentRunTotals.Count = %d, want 5", stats.RecentRunTotals.Count)
	}
	if stats.RecentRunTotals.Distance != 45000.0 {
		t.Errorf("RecentRunTotals.Distance = %f, want 45000.0", stats.RecentRunTotals.Distance)
	}
	if stats.YTDRunTotals.Count != 42 {
		t.Errorf("YTDRunTotals.Count = %d, want 42", stats.YTDRunTotals.Count)
	}
	if stats.AllRunTotals.Count != 500 {
		t.Errorf("AllRunTotals.Count = %d, want 500", stats.AllRunTotals.Count)
	}
	if stats.AllRunTotals.ElevationGain != 30000.0 {
		t.Errorf("AllRunTotals.ElevationGain = %f, want 30000.0", stats.AllRunTotals.ElevationGain)
	}
}

func TestFetchAthleteStatsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	_, err := FetchAthleteStats(t.Context(), server.Client(), server.URL, "token", 99)
	if err == nil {
		t.Error("expected error for 403 response")
	}
}

func TestFetchAthleteStatsNilClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AthleteStats{
			RecentRunTotals: ActivityTotals{Count: 1},
		})
	}))
	defer server.Close()

	stats, err := FetchAthleteStats(t.Context(), nil, server.URL, "token", 1)
	if err != nil {
		t.Fatalf("FetchAthleteStats with nil client: %v", err)
	}
	if stats.RecentRunTotals.Count != 1 {
		t.Errorf("RecentRunTotals.Count = %d, want 1", stats.RecentRunTotals.Count)
	}
}

func TestFetchGear(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/gear/g123456" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(GearDetail{
			ID:          "g123456",
			Name:        "Nike Vaporfly",
			Distance:    350000.0,
			BrandName:   "Nike",
			ModelName:   "Vaporfly 4%",
			Description: "Race day shoes",
			Primary:     true,
			Retired:     false,
		})
	}))
	defer server.Close()

	gear, err := FetchGear(t.Context(), server.Client(), server.URL, "test-token", "g123456")
	if err != nil {
		t.Fatalf("FetchGear: %v", err)
	}

	if gear.ID != "g123456" {
		t.Errorf("ID = %q, want %q", gear.ID, "g123456")
	}
	if gear.Name != "Nike Vaporfly" {
		t.Errorf("Name = %q, want %q", gear.Name, "Nike Vaporfly")
	}
	if gear.Distance != 350000.0 {
		t.Errorf("Distance = %f, want 350000.0", gear.Distance)
	}
	if gear.BrandName != "Nike" {
		t.Errorf("BrandName = %q, want %q", gear.BrandName, "Nike")
	}
	if !gear.Primary {
		t.Error("expected Primary = true")
	}
	if gear.Retired {
		t.Error("expected Retired = false")
	}
}

func TestFetchGearNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	_, err := FetchGear(t.Context(), server.Client(), server.URL, "token", "g000000")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestFetchGearNilClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(GearDetail{
			ID:   "g999",
			Name: "Test Shoe",
		})
	}))
	defer server.Close()

	gear, err := FetchGear(t.Context(), nil, server.URL, "token", "g999")
	if err != nil {
		t.Fatalf("FetchGear with nil client: %v", err)
	}
	if gear.ID != "g999" {
		t.Errorf("ID = %q, want %q", gear.ID, "g999")
	}
}
