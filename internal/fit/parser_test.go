package fit

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
	"time"

	"github.com/tormoder/fit"
)

func buildTestFITFile(t *testing.T, sport fit.Sport, distance uint32, timerTime uint32, avgHR uint8, maxHR uint8, avgCadence uint8, records []*fit.RecordMsg) *fit.File {
	t.Helper()
	h := fit.NewHeader(fit.V10, false)
	f, err := fit.NewFile(fit.FileTypeActivity, h)
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	af, err := f.Activity()
	if err != nil {
		t.Fatalf("Activity: %v", err)
	}

	session := fit.NewSessionMsg()
	session.Sport = sport
	session.StartTime = time.Date(2026, 3, 15, 7, 30, 0, 0, time.UTC)
	session.TotalDistance = distance
	session.TotalTimerTime = timerTime
	session.AvgHeartRate = avgHR
	session.MaxHeartRate = maxHR
	session.AvgCadence = avgCadence

	af.Sessions = append(af.Sessions, session)
	af.Activity = fit.NewActivityMsg()
	af.Records = records

	return roundTripFIT(t, f)
}

func roundTripFIT(t *testing.T, f *fit.File) *fit.File {
	t.Helper()
	var buf bytes.Buffer
	if err := fit.Encode(&buf, f, binary.LittleEndian); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := fit.Decode(&buf)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	return decoded
}

func TestParseFITData_RunningActivity(t *testing.T) {
	rec1 := fit.NewRecordMsg()
	rec1.HeartRate = 145
	rec1.Speed = 3333 // 3.333 m/s
	rec1.Cadence = 85

	rec2 := fit.NewRecordMsg()
	rec2.HeartRate = 155
	rec2.Speed = 3500 // 3.5 m/s
	rec2.Cadence = 87

	// distance raw 500000 = 5000m, timerTime raw 1800000 = 1800s (30 min)
	f := buildTestFITFile(t, fit.SportRunning, 500000, 1800000, 150, 180, 85, []*fit.RecordMsg{rec1, rec2})

	parsed, err := ParseFITData(f)
	if err != nil {
		t.Fatalf("ParseFITData: %v", err)
	}

	if parsed.ActivityType != "Run" {
		t.Errorf("ActivityType = %q, want %q", parsed.ActivityType, "Run")
	}
	if parsed.Distance != 5000.0 {
		t.Errorf("Distance = %.1f, want 5000.0", parsed.Distance)
	}
	if parsed.DurationSecs != 1800 {
		t.Errorf("DurationSecs = %d, want 1800", parsed.DurationSecs)
	}
	// 1800s / 5km = 360 sec/km
	if parsed.AvgPaceSecs != 360 {
		t.Errorf("AvgPaceSecs = %d, want 360", parsed.AvgPaceSecs)
	}
	if parsed.AvgHR != 150 {
		t.Errorf("AvgHR = %d, want 150", parsed.AvgHR)
	}
	if parsed.MaxHR != 180 {
		t.Errorf("MaxHR = %d, want 180", parsed.MaxHR)
	}
	// Running cadence: 85 * 2 = 170 spm
	if parsed.AvgCadence != 170 {
		t.Errorf("AvgCadence = %.0f, want 170", parsed.AvgCadence)
	}
	if parsed.Name != "Morning Run" {
		t.Errorf("Name = %q, want %q", parsed.Name, "Morning Run")
	}
}

func TestParseFITData_CyclingActivity(t *testing.T) {
	rec := fit.NewRecordMsg()
	rec.HeartRate = 140
	rec.Speed = 8000 // 8 m/s
	rec.Cadence = 90

	// 20km in 45min
	f := buildTestFITFile(t, fit.SportCycling, 2000000, 2700000, 140, 170, 90, []*fit.RecordMsg{rec})

	parsed, err := ParseFITData(f)
	if err != nil {
		t.Fatalf("ParseFITData: %v", err)
	}

	if parsed.ActivityType != "Ride" {
		t.Errorf("ActivityType = %q, want %q", parsed.ActivityType, "Ride")
	}
	// Cycling cadence is NOT doubled
	if parsed.AvgCadence != 90 {
		t.Errorf("AvgCadence = %.0f, want 90 (cycling cadence should NOT be doubled)", parsed.AvgCadence)
	}
}

func TestParseFITData_SwimmingActivity(t *testing.T) {
	f := buildTestFITFile(t, fit.SportSwimming, 200000, 3600000, 130, 160, 30, nil)

	parsed, err := ParseFITData(f)
	if err != nil {
		t.Fatalf("ParseFITData: %v", err)
	}

	if parsed.ActivityType != "Swim" {
		t.Errorf("ActivityType = %q, want %q", parsed.ActivityType, "Swim")
	}
}

func TestParseFITData_MissingHRRecords(t *testing.T) {
	rec := fit.NewRecordMsg()
	// HeartRate stays at 0xFF (invalid) from NewRecordMsg
	rec.Speed = 3000
	rec.Cadence = 85

	f := buildTestFITFile(t, fit.SportRunning, 500000, 1800000, 0xFF, 0xFF, 85, []*fit.RecordMsg{rec})

	parsed, err := ParseFITData(f)
	if err != nil {
		t.Fatalf("ParseFITData: %v", err)
	}

	if parsed.AvgHR != 0 {
		t.Errorf("AvgHR = %d, want 0 for missing data", parsed.AvgHR)
	}
	if parsed.MaxHR != 0 {
		t.Errorf("MaxHR = %d, want 0 for missing data", parsed.MaxHR)
	}
	if parsed.HeartRate != nil {
		t.Errorf("HeartRate stream should be nil when no valid HR records exist")
	}
}

func TestParseFITData_MissingSpeedRecords(t *testing.T) {
	rec := fit.NewRecordMsg()
	rec.HeartRate = 150
	// Speed stays at 0xFFFF (invalid) from NewRecordMsg
	rec.Cadence = 85

	f := buildTestFITFile(t, fit.SportRunning, 500000, 1800000, 150, 180, 85, []*fit.RecordMsg{rec})

	parsed, err := ParseFITData(f)
	if err != nil {
		t.Fatalf("ParseFITData: %v", err)
	}

	if parsed.Pace != nil {
		t.Errorf("Pace stream should be nil when no valid speed records exist")
	}
}

func TestParseFITData_NoRecords(t *testing.T) {
	f := buildTestFITFile(t, fit.SportRunning, 500000, 1800000, 150, 180, 85, nil)

	parsed, err := ParseFITData(f)
	if err != nil {
		t.Fatalf("ParseFITData: %v", err)
	}

	if parsed.HeartRate != nil {
		t.Error("HeartRate should be nil with no records")
	}
	if parsed.Pace != nil {
		t.Error("Pace should be nil with no records")
	}
	if parsed.Cadence != nil {
		t.Error("Cadence should be nil with no records")
	}
}

func TestParseFITData_ZeroDistance(t *testing.T) {
	// Invalid distance (0xFFFFFFFF from NewSessionMsg means NaN when scaled)
	f := buildTestFITFile(t, fit.SportRunning, 0xFFFFFFFF, 1800000, 150, 180, 85, nil)

	parsed, err := ParseFITData(f)
	if err != nil {
		t.Fatalf("ParseFITData: %v", err)
	}

	if parsed.Distance != 0 {
		t.Errorf("Distance = %.1f, want 0 for invalid distance", parsed.Distance)
	}
	if parsed.AvgPaceSecs != 0 {
		t.Errorf("AvgPaceSecs = %d, want 0 when distance is zero", parsed.AvgPaceSecs)
	}
}

func TestParseFITData_NilFile(t *testing.T) {
	_, err := ParseFITData(nil)
	if err == nil {
		t.Error("expected error for nil file")
	}
}

func TestParseFITData_NonActivityFile(t *testing.T) {
	h := fit.NewHeader(fit.V10, false)
	f, err := fit.NewFile(fit.FileTypeSettings, h)
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	_, err = ParseFITData(f)
	if err == nil {
		t.Error("expected error for non-activity file")
	}
}

func TestParseFITFile_EmptyPath(t *testing.T) {
	_, err := ParseFITFile("")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestParseFITFile_NonExistentFile(t *testing.T) {
	_, err := ParseFITFile("/nonexistent/path/file.fit")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParseFITFile_InvalidFile(t *testing.T) {
	// Write garbage to a temp file
	tmpFile := t.TempDir() + "/garbage.fit"
	if err := writeGarbageFile(tmpFile); err != nil {
		t.Fatalf("write garbage file: %v", err)
	}

	_, err := ParseFITFile(tmpFile)
	if err == nil {
		t.Error("expected error for invalid FIT file")
	}
}

func writeGarbageFile(path string) error {
	return os.WriteFile(path, []byte("this is not a FIT file"), 0644)
}

func TestDeduplicationHash_Deterministic(t *testing.T) {
	a := &ParsedActivity{
		StartDate:    time.Date(2026, 3, 15, 7, 30, 0, 0, time.UTC),
		DurationSecs: 1800,
		Distance:     5000.0,
	}

	hash1 := DeduplicationHash(a)
	hash2 := DeduplicationHash(a)

	if hash1 == "" {
		t.Error("hash should not be empty")
	}
	if hash1 != hash2 {
		t.Errorf("same input produced different hashes: %q vs %q", hash1, hash2)
	}
}

func TestDeduplicationHash_DifferentInputs(t *testing.T) {
	a := &ParsedActivity{
		StartDate:    time.Date(2026, 3, 15, 7, 30, 0, 0, time.UTC),
		DurationSecs: 1800,
		Distance:     5000.0,
	}
	b := &ParsedActivity{
		StartDate:    time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		DurationSecs: 1800,
		Distance:     5000.0,
	}

	hashA := DeduplicationHash(a)
	hashB := DeduplicationHash(b)

	if hashA == hashB {
		t.Error("different inputs should produce different hashes")
	}
}

func TestDeduplicationHash_DifferentDistance(t *testing.T) {
	a := &ParsedActivity{
		StartDate:    time.Date(2026, 3, 15, 7, 30, 0, 0, time.UTC),
		DurationSecs: 1800,
		Distance:     5000.0,
	}
	b := &ParsedActivity{
		StartDate:    time.Date(2026, 3, 15, 7, 30, 0, 0, time.UTC),
		DurationSecs: 1800,
		Distance:     10000.0,
	}

	if DeduplicationHash(a) == DeduplicationHash(b) {
		t.Error("different distances should produce different hashes")
	}
}

func TestDeduplicationHash_Nil(t *testing.T) {
	if hash := DeduplicationHash(nil); hash != "" {
		t.Errorf("nil input should return empty hash, got %q", hash)
	}
}

func TestStreams_Running_CadenceDoubled(t *testing.T) {
	rec := fit.NewRecordMsg()
	rec.HeartRate = 150
	rec.Speed = 3333
	rec.Cadence = 85

	f := buildTestFITFile(t, fit.SportRunning, 500000, 1800000, 150, 180, 85, []*fit.RecordMsg{rec})

	parsed, err := ParseFITData(f)
	if err != nil {
		t.Fatalf("ParseFITData: %v", err)
	}

	if len(parsed.Cadence) != 1 {
		t.Fatalf("expected 1 cadence record, got %d", len(parsed.Cadence))
	}
	// Running cadence: 85 * 2 = 170 spm
	if parsed.Cadence[0] != 170 {
		t.Errorf("running cadence = %.0f, want 170 (85 * 2)", parsed.Cadence[0])
	}
}

func TestStreams_Cycling_CadenceNotDoubled(t *testing.T) {
	rec := fit.NewRecordMsg()
	rec.HeartRate = 140
	rec.Speed = 8000
	rec.Cadence = 90

	f := buildTestFITFile(t, fit.SportCycling, 2000000, 2700000, 140, 170, 90, []*fit.RecordMsg{rec})

	parsed, err := ParseFITData(f)
	if err != nil {
		t.Fatalf("ParseFITData: %v", err)
	}

	if len(parsed.Cadence) != 1 {
		t.Fatalf("expected 1 cadence record, got %d", len(parsed.Cadence))
	}
	if parsed.Cadence[0] != 90 {
		t.Errorf("cycling cadence = %.0f, want 90 (should NOT be doubled)", parsed.Cadence[0])
	}
}

func TestStreams_PaceFromSpeed(t *testing.T) {
	rec := fit.NewRecordMsg()
	rec.HeartRate = 150
	rec.Speed = 4000 // 4.0 m/s → 1000/4 = 250 sec/km
	rec.Cadence = 85

	f := buildTestFITFile(t, fit.SportRunning, 500000, 1800000, 150, 180, 85, []*fit.RecordMsg{rec})

	parsed, err := ParseFITData(f)
	if err != nil {
		t.Fatalf("ParseFITData: %v", err)
	}

	if len(parsed.Pace) != 1 {
		t.Fatalf("expected 1 pace record, got %d", len(parsed.Pace))
	}
	if parsed.Pace[0] != 250 {
		t.Errorf("pace = %d, want 250 sec/km (from 4.0 m/s)", parsed.Pace[0])
	}
}

func TestBuildActivityName_TimeOfDay(t *testing.T) {
	tests := []struct {
		hour int
		want string
	}{
		{3, "Early Morning Run"},
		{8, "Morning Run"},
		{14, "Afternoon Run"},
		{19, "Evening Run"},
		{22, "Night Run"},
	}

	for _, tt := range tests {
		session := fit.NewSessionMsg()
		session.Sport = fit.SportRunning
		session.StartTime = time.Date(2026, 3, 15, tt.hour, 0, 0, 0, time.UTC)

		got := buildActivityName(session)
		if got != tt.want {
			t.Errorf("hour=%d: got %q, want %q", tt.hour, got, tt.want)
		}
	}
}

func TestMapSport(t *testing.T) {
	tests := []struct {
		sport fit.Sport
		want  string
	}{
		{fit.SportRunning, "Run"},
		{fit.SportCycling, "Ride"},
		{fit.SportEBiking, "Ride"},
		{fit.SportSwimming, "Swim"},
		{fit.SportWalking, "Walk"},
		{fit.SportHiking, "Hike"},
		{fit.SportRowing, "Rowing"},
		{fit.SportTraining, "Workout"},
		{fit.SportCrossCountrySkiing, "NordicSki"},
		{fit.SportAlpineSkiing, "AlpineSki"},
		{fit.SportSnowboarding, "Snowboard"},
		{fit.SportGolf, "Golf"},
	}

	for _, tt := range tests {
		got := mapSport(tt.sport)
		if got != tt.want {
			t.Errorf("mapSport(%v) = %q, want %q", tt.sport, got, tt.want)
		}
	}
}
