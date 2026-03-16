package context

import (
	"strings"
	"testing"

	"coachlm/internal/storage"
)

func TestFormatProfileBlock_FullProfile(t *testing.T) {
	p := &storage.AthleteProfile{
		Age:                 35,
		MaxHR:               185,
		ThresholdPaceSecs:   270,
		WeeklyMileageTarget: 50.0,
		RaceGoals:           "Sub-3 marathon",
		InjuryHistory:       "IT band 2024-06",
	}

	got := FormatProfileBlock(p)

	expected := []string{
		"Age: 35",
		"Max Heart Rate: 185 bpm",
		"Threshold Pace: 4:30/km",
		"Weekly Mileage Target: 50.0 km",
		"Race Goals: Sub-3 marathon",
		"Injury History: IT band 2024-06",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestFormatProfileBlock_PartialProfile(t *testing.T) {
	p := &storage.AthleteProfile{
		Age:               28,
		MaxHR:             190,
		ThresholdPaceSecs: 300,
	}

	got := FormatProfileBlock(p)

	if !strings.Contains(got, "Age: 28") {
		t.Errorf("expected Age present, got:\n%s", got)
	}
	if !strings.Contains(got, "Threshold Pace: 5:00/km") {
		t.Errorf("expected Threshold Pace present, got:\n%s", got)
	}
	if strings.Contains(got, "Race Goals") {
		t.Errorf("expected Race Goals omitted, got:\n%s", got)
	}
	if strings.Contains(got, "Injury History") {
		t.Errorf("expected Injury History omitted, got:\n%s", got)
	}
	if strings.Contains(got, "Weekly Mileage Target") {
		t.Errorf("expected Weekly Mileage Target omitted, got:\n%s", got)
	}
}

func TestFormatProfileBlock_NilProfile(t *testing.T) {
	got := FormatProfileBlock(nil)
	if got != "No profile configured." {
		t.Errorf("expected 'No profile configured.', got %q", got)
	}
}

func TestFormatProfileBlock_ZeroMileageOmitted(t *testing.T) {
	p := &storage.AthleteProfile{
		Age:                 30,
		MaxHR:               180,
		ThresholdPaceSecs:   240,
		WeeklyMileageTarget: 0,
	}

	got := FormatProfileBlock(p)
	if strings.Contains(got, "Weekly Mileage Target") {
		t.Errorf("expected Weekly Mileage Target omitted when zero, got:\n%s", got)
	}
}

func TestFormatProfileBlock_AllZeroFields(t *testing.T) {
	p := &storage.AthleteProfile{}
	got := FormatProfileBlock(p)
	if got != "No profile configured." {
		t.Errorf("all-zero profile should return 'No profile configured.', got %q", got)
	}
}

func TestFormatPace(t *testing.T) {
	tests := []struct {
		secs int
		want string
	}{
		{270, "4:30/km"},
		{300, "5:00/km"},
		{195, "3:15/km"},
		{360, "6:00/km"},
		{245, "4:05/km"},
	}
	for _, tt := range tests {
		got := FormatPace(tt.secs)
		if got != tt.want {
			t.Errorf("FormatPace(%d) = %q, want %q", tt.secs, got, tt.want)
		}
	}
}

func TestFormatProfileBlock_Deterministic(t *testing.T) {
	p := &storage.AthleteProfile{
		Age:                 35,
		MaxHR:               185,
		ThresholdPaceSecs:   270,
		WeeklyMileageTarget: 50.0,
		RaceGoals:           "Sub-3 marathon",
		InjuryHistory:       "IT band 2024-06",
	}

	first := FormatProfileBlock(p)
	second := FormatProfileBlock(p)

	if first != second {
		t.Errorf("non-deterministic output:\nfirst:  %q\nsecond: %q", first, second)
	}
}
