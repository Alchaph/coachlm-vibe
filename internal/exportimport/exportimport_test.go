package exportimport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"coachlm/internal/storage"
)

func TestExport_EmptyDB(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "export.coachctx")

	if err := Export(db, filePath); err != nil {
		t.Fatalf("Export: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read export file: %v", err)
	}

	os.Remove(filePath)

	var envelope ImportEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("unmarshal export: %v", err)
	}

	if envelope.SchemaVersion != 1 {
		t.Errorf("schema version = %d, want 1", envelope.SchemaVersion)
	}
	if envelope.AthleteProfile != nil {
		t.Error("expected nil profile for empty DB")
	}
	if len(envelope.TrainingSummaries) != 0 {
		t.Errorf("training summaries = %d, want 0", len(envelope.TrainingSummaries))
	}
	if len(envelope.PinnedInsights) != 0 {
		t.Errorf("pinned insights = %d, want 0", len(envelope.PinnedInsights))
	}
	if envelope.SettingsMeta.ActiveLLM != "" {
		t.Error("expected empty active LLM for empty DB")
	}
}

func TestExport_PopulatedDB(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	profile := &storage.AthleteProfile{
		Age:                 30,
		MaxHR:               185,
		ThresholdPaceSecs:   270,
		WeeklyMileageTarget: 50.0,
		RaceGoals:           "Sub-3 marathon",
	}
	if err := db.SaveProfile(profile); err != nil {
		t.Fatalf("save profile: %v", err)
	}

	activity := &storage.Activity{
		StravaID:     123,
		Name:         "Test Run",
		ActivityType: "Run",
		StartDate:    time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		Distance:     10000.0,
		DurationSecs: 3000,
		AvgPaceSecs:  300,
		AvgHR:        150,
		MaxHR:        175,
		AvgCadence:   180.0,
		Source:       "test",
	}
	if err := db.SaveActivity(activity); err != nil {
		t.Fatalf("save activity: %v", err)
	}

	insight1 := "Test insight 1"
	insight2 := "Test insight 2"
	if _, err := db.SaveInsight(insight1, "s1"); err != nil {
		t.Fatalf("save insight 1: %v", err)
	}
	if _, err := db.SaveInsight(insight2, "s2"); err != nil {
		t.Fatalf("save insight 2: %v", err)
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "export.coachctx")

	if err := Export(db, filePath); err != nil {
		t.Fatalf("Export: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read export file: %v", err)
	}

	os.Remove(filePath)

	var envelope ImportEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("unmarshal export: %v", err)
	}

	if envelope.SchemaVersion != 1 {
		t.Errorf("schema version = %d, want 1", envelope.SchemaVersion)
	}
	if envelope.AthleteProfile == nil {
		t.Fatal("expected profile")
	}
	if envelope.AthleteProfile.Age != 30 {
		t.Errorf("profile age = %d, want 30", envelope.AthleteProfile.Age)
	}
	if len(envelope.TrainingSummaries) != 1 {
		t.Errorf("training summaries = %d, want 1", len(envelope.TrainingSummaries))
	}
	if len(envelope.PinnedInsights) != 2 {
		t.Errorf("pinned insights = %d, want 2", len(envelope.PinnedInsights))
	}
}

func TestImport_Additive(t *testing.T) {
	db1 := newTestDB(t)
	defer db1.Close()

	profile := &storage.AthleteProfile{
		Age:                 30,
		MaxHR:               180,
		ThresholdPaceSecs:   300,
		WeeklyMileageTarget: 50.0,
		RaceGoals:           "Sub-3 marathon",
	}
	if err := db1.SaveProfile(profile); err != nil {
		t.Fatalf("save profile: %v", err)
	}

	activity := &storage.Activity{
		StravaID:     1,
		Name:         "Existing Run",
		ActivityType: "Run",
		StartDate:    time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		Distance:     5000.0,
		DurationSecs: 1500,
		Source:       "existing",
	}
	if err := db1.SaveActivity(activity); err != nil {
		t.Fatalf("save activity: %v", err)
	}

	insight := "Existing insight"
	if _, err := db1.SaveInsight(insight, "s1"); err != nil {
		t.Fatalf("save insight: %v", err)
	}

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.coachctx")

	if err := Export(db1, exportPath); err != nil {
		t.Fatalf("Export: %v", err)
	}

	db2 := newTestDB(t)
	defer db2.Close()

	if err := Import(db2, exportPath, false); err != nil {
		t.Fatalf("Import: %v", err)
	}

	loadedProfile, _ := db2.GetProfile()
	if loadedProfile == nil || loadedProfile.Age != 30 {
		t.Errorf("profile not imported correctly")
	}

	loadedActivities, _ := db2.ListActivities(10, 0)
	if len(loadedActivities) != 1 {
		t.Errorf("activities = %d, want 1", len(loadedActivities))
	}

	loadedInsights, _ := db2.GetInsights()
	if len(loadedInsights) != 1 {
		t.Errorf("insights = %d, want 1", len(loadedInsights))
	}
}

func TestImport_ReplaceAll(t *testing.T) {
	db1 := newTestDB(t)
	defer db1.Close()

	profile := &storage.AthleteProfile{
		Age:                 35,
		MaxHR:               180,
		ThresholdPaceSecs:   300,
		WeeklyMileageTarget: 50.0,
		RaceGoals:           "Sub-3 marathon",
	}
	if err := db1.SaveProfile(profile); err != nil {
		t.Fatalf("save profile: %v", err)
	}

	activity := &storage.Activity{
		StravaID:     1,
		Name:         "Old Run",
		ActivityType: "Run",
		StartDate:    time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		Distance:     8000.0,
		DurationSecs: 2000,
		Source:       "old",
	}
	if err := db1.SaveActivity(activity); err != nil {
		t.Fatalf("save activity: %v", err)
	}

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.coachctx")

	if err := Export(db1, exportPath); err != nil {
		t.Fatalf("Export: %v", err)
	}

	db2 := newTestDB(t)
	defer db2.Close()

	if err := Import(db2, exportPath, true); err != nil {
		t.Fatalf("Import: %v", err)
	}

	loadedProfile, _ := db2.GetProfile()
	if loadedProfile == nil || loadedProfile.Age != 35 {
		t.Errorf("profile was not replaced correctly, got age = %d", loadedProfile.Age)
	}

	loadedActivities, _ := db2.ListActivities(10, 0)
	if len(loadedActivities) != 1 {
		t.Errorf("activities = %d, want 1", len(loadedActivities))
	}
	if loadedActivities[0].Name != "Old Run" {
		t.Errorf("activity name = %q, want %q", loadedActivities[0].Name, "Old Run")
	}
}

func TestImport_SchemaVersionRejectsHigher(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	exportData := `{"schema_version":2,"exported_at":"2026-03-16T00:00:00Z"}`
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.coachctx")

	if err := os.WriteFile(exportPath, []byte(exportData), 0644); err != nil {
		t.Fatalf("write test export file: %v", err)
	}
	defer os.Remove(exportPath)

	importErr := Import(db, exportPath, false)
	if importErr == nil {
		t.Error("expected error for schema version 2")
	}

	if importErr != nil {
		expectedErr := "unsupported schema version: 2 (max supported: 1)"
		if importErr.Error() != expectedErr {
			t.Errorf("wrong error message: %v", importErr)
		}
	}
}

func TestImport_NonJSONFile(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "invalid.txt")

	if err := os.WriteFile(exportPath, []byte("not json"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	defer os.Remove(exportPath)

	if err := Import(db, exportPath, false); err == nil {
		t.Error("expected error for non-JSON file")
	}
}

func newTestDB(t *testing.T) *storage.DB {
	db, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("create test DB: %v", err)
	}
	return db
}
