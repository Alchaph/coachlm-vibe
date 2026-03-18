package storage

import (
	"testing"
	"time"
)

func sampleStats() *AthleteStats {
	return &AthleteStats{
		RecentRunCount:      5,
		RecentRunDistance:   50000,
		RecentRunMovingTime: 18000,
		RecentRunElevation:  400,
		YTDRunCount:         80,
		YTDRunDistance:      800000,
		YTDRunMovingTime:    288000,
		YTDRunElevation:     6400,
		AllRunCount:         300,
		AllRunDistance:      3000000,
		AllRunMovingTime:    1080000,
		AllRunElevation:     24000,
	}
}

func TestSaveAndGetAthleteStats(t *testing.T) {
	db := newTestDB(t)
	s := sampleStats()

	if err := db.SaveAthleteStats(s); err != nil {
		t.Fatalf("SaveAthleteStats: %v", err)
	}

	got, err := db.GetAthleteStats()
	if err != nil {
		t.Fatalf("GetAthleteStats: %v", err)
	}
	if got == nil {
		t.Fatal("GetAthleteStats returned nil after save")
	}

	if got.RecentRunCount != s.RecentRunCount {
		t.Errorf("RecentRunCount = %d, want %d", got.RecentRunCount, s.RecentRunCount)
	}
	if got.RecentRunDistance != s.RecentRunDistance {
		t.Errorf("RecentRunDistance = %f, want %f", got.RecentRunDistance, s.RecentRunDistance)
	}
	if got.RecentRunMovingTime != s.RecentRunMovingTime {
		t.Errorf("RecentRunMovingTime = %d, want %d", got.RecentRunMovingTime, s.RecentRunMovingTime)
	}
	if got.RecentRunElevation != s.RecentRunElevation {
		t.Errorf("RecentRunElevation = %f, want %f", got.RecentRunElevation, s.RecentRunElevation)
	}
	if got.YTDRunCount != s.YTDRunCount {
		t.Errorf("YTDRunCount = %d, want %d", got.YTDRunCount, s.YTDRunCount)
	}
	if got.YTDRunDistance != s.YTDRunDistance {
		t.Errorf("YTDRunDistance = %f, want %f", got.YTDRunDistance, s.YTDRunDistance)
	}
	if got.YTDRunMovingTime != s.YTDRunMovingTime {
		t.Errorf("YTDRunMovingTime = %d, want %d", got.YTDRunMovingTime, s.YTDRunMovingTime)
	}
	if got.YTDRunElevation != s.YTDRunElevation {
		t.Errorf("YTDRunElevation = %f, want %f", got.YTDRunElevation, s.YTDRunElevation)
	}
	if got.AllRunCount != s.AllRunCount {
		t.Errorf("AllRunCount = %d, want %d", got.AllRunCount, s.AllRunCount)
	}
	if got.AllRunDistance != s.AllRunDistance {
		t.Errorf("AllRunDistance = %f, want %f", got.AllRunDistance, s.AllRunDistance)
	}
	if got.AllRunMovingTime != s.AllRunMovingTime {
		t.Errorf("AllRunMovingTime = %d, want %d", got.AllRunMovingTime, s.AllRunMovingTime)
	}
	if got.AllRunElevation != s.AllRunElevation {
		t.Errorf("AllRunElevation = %f, want %f", got.AllRunElevation, s.AllRunElevation)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestGetAthleteStatsEmpty(t *testing.T) {
	db := newTestDB(t)

	got, err := db.GetAthleteStats()
	if err != nil {
		t.Fatalf("GetAthleteStats on empty db: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil on empty db, got %+v", got)
	}
}

func TestSaveAthleteStatsNil(t *testing.T) {
	db := newTestDB(t)

	err := db.SaveAthleteStats(nil)
	if err == nil {
		t.Error("SaveAthleteStats(nil) should return an error")
	}
}

func TestSaveAthleteStatsUpdate(t *testing.T) {
	db := newTestDB(t)
	s := sampleStats()

	if err := db.SaveAthleteStats(s); err != nil {
		t.Fatalf("first SaveAthleteStats: %v", err)
	}

	first, err := db.GetAthleteStats()
	if err != nil {
		t.Fatalf("first GetAthleteStats: %v", err)
	}

	time.Sleep(2 * time.Millisecond)

	s.RecentRunCount = 10
	s.RecentRunDistance = 100000
	if err := db.SaveAthleteStats(s); err != nil {
		t.Fatalf("second SaveAthleteStats: %v", err)
	}

	second, err := db.GetAthleteStats()
	if err != nil {
		t.Fatalf("second GetAthleteStats: %v", err)
	}

	if second.RecentRunCount != 10 {
		t.Errorf("RecentRunCount after update = %d, want 10", second.RecentRunCount)
	}
	if second.RecentRunDistance != 100000 {
		t.Errorf("RecentRunDistance after update = %f, want 100000", second.RecentRunDistance)
	}

	if !second.CreatedAt.Equal(first.CreatedAt) {
		t.Errorf("CreatedAt changed after update: %v → %v", first.CreatedAt, second.CreatedAt)
	}
}
