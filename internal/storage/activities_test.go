package storage

import (
	"testing"
	"time"
)

func TestSaveAndGetActivity(t *testing.T) {
	db := newTestDB(t)

	activity := &Activity{
		StravaID:     12345,
		Name:         "Morning Run",
		ActivityType: "Run",
		StartDate:    time.Date(2026, 3, 15, 7, 0, 0, 0, time.UTC),
		Distance:     10000.0,
		DurationSecs: 3000,
		AvgPaceSecs:  300,
		AvgHR:        150,
		MaxHR:        175,
		AvgCadence:   180.0,
		Source:       "strava",
	}

	if err := db.SaveActivity(activity); err != nil {
		t.Fatalf("SaveActivity: %v", err)
	}

	got, err := db.GetActivityByStravaID(12345)
	if err != nil {
		t.Fatalf("GetActivityByStravaID: %v", err)
	}
	if got == nil {
		t.Fatal("expected activity, got nil")
	}
	if got.Name != "Morning Run" {
		t.Errorf("Name = %q, want %q", got.Name, "Morning Run")
	}
	if got.StravaID != 12345 {
		t.Errorf("StravaID = %d, want 12345", got.StravaID)
	}
	if got.Distance != 10000.0 {
		t.Errorf("Distance = %f, want 10000.0", got.Distance)
	}
	if got.Source != "strava" {
		t.Errorf("Source = %q, want %q", got.Source, "strava")
	}
}

func TestSaveActivityDedup(t *testing.T) {
	db := newTestDB(t)

	activity := &Activity{
		StravaID:     99999,
		Name:         "First Save",
		ActivityType: "Run",
		StartDate:    time.Date(2026, 3, 15, 7, 0, 0, 0, time.UTC),
		Distance:     5000.0,
		DurationSecs: 1500,
		Source:       "strava",
	}

	if err := db.SaveActivity(activity); err != nil {
		t.Fatalf("first SaveActivity: %v", err)
	}

	activity.Name = "Second Save"
	if err := db.SaveActivity(activity); err != nil {
		t.Fatalf("second SaveActivity should not error: %v", err)
	}

	got, err := db.GetActivityByStravaID(99999)
	if err != nil {
		t.Fatalf("GetActivityByStravaID: %v", err)
	}
	if got.Name != "First Save" {
		t.Errorf("Name = %q, want %q (INSERT OR IGNORE should keep first)", got.Name, "First Save")
	}
}

func TestGetActivityByStravaIDNotFound(t *testing.T) {
	db := newTestDB(t)

	got, err := db.GetActivityByStravaID(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for non-existent activity, got %+v", got)
	}
}

func TestListActivitiesOrderAndPagination(t *testing.T) {
	db := newTestDB(t)

	activities := []Activity{
		{StravaID: 1, Name: "Oldest", StartDate: time.Date(2026, 3, 10, 7, 0, 0, 0, time.UTC), Source: "strava"},
		{StravaID: 2, Name: "Middle", StartDate: time.Date(2026, 3, 12, 7, 0, 0, 0, time.UTC), Source: "strava"},
		{StravaID: 3, Name: "Newest", StartDate: time.Date(2026, 3, 14, 7, 0, 0, 0, time.UTC), Source: "strava"},
	}
	for i := range activities {
		if err := db.SaveActivity(&activities[i]); err != nil {
			t.Fatalf("SaveActivity %d: %v", i, err)
		}
	}

	got, err := db.ListActivities(10, 0)
	if err != nil {
		t.Fatalf("ListActivities: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d activities, want 3", len(got))
	}
	if got[0].Name != "Newest" {
		t.Errorf("first activity = %q, want %q (DESC order)", got[0].Name, "Newest")
	}
	if got[2].Name != "Oldest" {
		t.Errorf("last activity = %q, want %q", got[2].Name, "Oldest")
	}

	page, err := db.ListActivities(2, 0)
	if err != nil {
		t.Fatalf("ListActivities page 1: %v", err)
	}
	if len(page) != 2 {
		t.Fatalf("page 1: got %d activities, want 2", len(page))
	}

	page2, err := db.ListActivities(2, 2)
	if err != nil {
		t.Fatalf("ListActivities page 2: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("page 2: got %d activities, want 1", len(page2))
	}
	if page2[0].Name != "Oldest" {
		t.Errorf("page 2 first = %q, want %q", page2[0].Name, "Oldest")
	}
}

func TestSaveActivityNil(t *testing.T) {
	db := newTestDB(t)

	if err := db.SaveActivity(nil); err == nil {
		t.Error("expected error for nil activity")
	}
}

func TestListActivitiesEmpty(t *testing.T) {
	db := newTestDB(t)

	got, err := db.ListActivities(10, 0)
	if err != nil {
		t.Fatalf("ListActivities: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for empty list, got %d activities", len(got))
	}
}

func TestGetActivityStatsEmpty(t *testing.T) {
	db := newTestDB(t)

	stats, err := db.GetActivityStats()
	if err != nil {
		t.Fatalf("GetActivityStats: %v", err)
	}
	if stats.TotalCount != 0 {
		t.Errorf("TotalCount = %d, want 0", stats.TotalCount)
	}
	if stats.TotalDistanceKm != 0 {
		t.Errorf("TotalDistanceKm = %f, want 0", stats.TotalDistanceKm)
	}
	if stats.EarliestDate != "" {
		t.Errorf("EarliestDate = %q, want empty", stats.EarliestDate)
	}
	if stats.LatestDate != "" {
		t.Errorf("LatestDate = %q, want empty", stats.LatestDate)
	}
}

func TestGetActivityStatsMultiple(t *testing.T) {
	db := newTestDB(t)

	activities := []Activity{
		{StravaID: 1, Name: "Run 1", StartDate: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), Distance: 5000.0, DurationSecs: 1800, Source: "strava"},
		{StravaID: 2, Name: "Run 2", StartDate: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC), Distance: 10000.0, DurationSecs: 3600, Source: "strava"},
		{StravaID: 3, Name: "Run 3", StartDate: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), Distance: 7500.0, DurationSecs: 2700, Source: "strava"},
	}
	for i := range activities {
		if err := db.SaveActivity(&activities[i]); err != nil {
			t.Fatalf("SaveActivity %d: %v", i, err)
		}
	}

	stats, err := db.GetActivityStats()
	if err != nil {
		t.Fatalf("GetActivityStats: %v", err)
	}
	if stats.TotalCount != 3 {
		t.Errorf("TotalCount = %d, want 3", stats.TotalCount)
	}
	if stats.TotalDistanceKm != 22500.0 {
		t.Errorf("TotalDistanceKm = %f, want 22500.0", stats.TotalDistanceKm)
	}
	if stats.EarliestDate != "2026-01-10" {
		t.Errorf("EarliestDate = %q, want 2026-01-10", stats.EarliestDate)
	}
	if stats.LatestDate != "2026-03-20" {
		t.Errorf("LatestDate = %q, want 2026-03-20", stats.LatestDate)
	}
}

func TestGetActivityStatsSingle(t *testing.T) {
	db := newTestDB(t)

	activity := &Activity{
		StravaID:     1,
		Name:         "Solo Run",
		StartDate:    time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
		Distance:     8000.0,
		DurationSecs: 3000,
		Source:       "strava",
	}
	if err := db.SaveActivity(activity); err != nil {
		t.Fatalf("SaveActivity: %v", err)
	}

	stats, err := db.GetActivityStats()
	if err != nil {
		t.Fatalf("GetActivityStats: %v", err)
	}
	if stats.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", stats.TotalCount)
	}
	if stats.TotalDistanceKm != 8000.0 {
		t.Errorf("TotalDistanceKm = %f, want 8000.0", stats.TotalDistanceKm)
	}
	if stats.EarliestDate != "2026-05-05" {
		t.Errorf("EarliestDate = %q, want 2026-05-05", stats.EarliestDate)
	}
	if stats.LatestDate != "2026-05-05" {
		t.Errorf("LatestDate = %q, want 2026-05-05", stats.LatestDate)
	}
}

func TestGetActivityStatsZeroDistance(t *testing.T) {
	db := newTestDB(t)

	activity := &Activity{
		StravaID:     1,
		Name:         "Treadmill Run",
		StartDate:    time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Distance:     0,
		DurationSecs: 1800,
		Source:       "strava",
	}
	if err := db.SaveActivity(activity); err != nil {
		t.Fatalf("SaveActivity: %v", err)
	}

	stats, err := db.GetActivityStats()
	if err != nil {
		t.Fatalf("GetActivityStats: %v", err)
	}
	if stats.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", stats.TotalCount)
	}
	if stats.TotalDistanceKm != 0 {
		t.Errorf("TotalDistanceKm = %f, want 0 for zero-distance activity", stats.TotalDistanceKm)
	}
}
