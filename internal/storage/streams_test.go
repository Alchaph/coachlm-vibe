package storage

import (
	"encoding/json"
	"testing"
	"time"
)

func testActivity(t *testing.T, db *DB) int64 {
	t.Helper()
	activity := &Activity{
		StravaID:     55555,
		Name:         "Test Run",
		ActivityType: "Run",
		StartDate:    time.Date(2026, 3, 15, 7, 0, 0, 0, time.UTC),
		Distance:     10000.0,
		DurationSecs: 3000,
		Source:       "strava",
	}
	if err := db.SaveActivity(activity); err != nil {
		t.Fatalf("SaveActivity: %v", err)
	}
	got, err := db.GetActivityByStravaID(55555)
	if err != nil {
		t.Fatalf("GetActivityByStravaID: %v", err)
	}
	return got.ID
}

func TestSaveAndGetActivityStream(t *testing.T) {
	db := newTestDB(t)
	activityID := testActivity(t, db)

	hrData, _ := json.Marshal([]int{140, 145, 150, 155, 160})
	if err := db.SaveActivityStream(activityID, "heartrate", hrData); err != nil {
		t.Fatalf("SaveActivityStream heartrate: %v", err)
	}

	streams, err := db.GetActivityStreams(activityID)
	if err != nil {
		t.Fatalf("GetActivityStreams: %v", err)
	}
	if len(streams) != 1 {
		t.Fatalf("got %d streams, want 1", len(streams))
	}
	if streams[0].StreamType != "heartrate" {
		t.Errorf("StreamType = %q, want heartrate", streams[0].StreamType)
	}
	if streams[0].ActivityID != activityID {
		t.Errorf("ActivityID = %d, want %d", streams[0].ActivityID, activityID)
	}

	var hrValues []int
	if err := json.Unmarshal(streams[0].Data, &hrValues); err != nil {
		t.Fatalf("unmarshal HR data: %v", err)
	}
	if len(hrValues) != 5 {
		t.Fatalf("HR values len = %d, want 5", len(hrValues))
	}
	if hrValues[0] != 140 || hrValues[4] != 160 {
		t.Errorf("HR values = %v, want [140 145 150 155 160]", hrValues)
	}
}

func TestSaveMultipleStreamTypes(t *testing.T) {
	db := newTestDB(t)
	activityID := testActivity(t, db)

	hrData, _ := json.Marshal([]int{140, 145})
	paceData, _ := json.Marshal([]float64{333.3, 285.7})
	cadData, _ := json.Marshal([]int{170, 175})

	for _, tc := range []struct {
		streamType string
		data       json.RawMessage
	}{
		{"heartrate", hrData},
		{"pace", paceData},
		{"cadence", cadData},
	} {
		if err := db.SaveActivityStream(activityID, tc.streamType, tc.data); err != nil {
			t.Fatalf("SaveActivityStream %s: %v", tc.streamType, err)
		}
	}

	streams, err := db.GetActivityStreams(activityID)
	if err != nil {
		t.Fatalf("GetActivityStreams: %v", err)
	}
	if len(streams) != 3 {
		t.Fatalf("got %d streams, want 3", len(streams))
	}

	types := map[string]bool{}
	for _, s := range streams {
		types[s.StreamType] = true
	}
	for _, want := range []string{"heartrate", "pace", "cadence"} {
		if !types[want] {
			t.Errorf("missing stream type %q", want)
		}
	}
}

func TestGetActivityStreamByType(t *testing.T) {
	db := newTestDB(t)
	activityID := testActivity(t, db)

	hrData, _ := json.Marshal([]int{140, 145, 150})
	cadData, _ := json.Marshal([]int{170, 175, 180})

	if err := db.SaveActivityStream(activityID, "heartrate", hrData); err != nil {
		t.Fatalf("SaveActivityStream heartrate: %v", err)
	}
	if err := db.SaveActivityStream(activityID, "cadence", cadData); err != nil {
		t.Fatalf("SaveActivityStream cadence: %v", err)
	}

	hr, err := db.GetActivityStreamByType(activityID, "heartrate")
	if err != nil {
		t.Fatalf("GetActivityStreamByType heartrate: %v", err)
	}
	if hr == nil {
		t.Fatal("expected heartrate stream, got nil")
	}
	if hr.StreamType != "heartrate" {
		t.Errorf("StreamType = %q, want heartrate", hr.StreamType)
	}

	cad, err := db.GetActivityStreamByType(activityID, "cadence")
	if err != nil {
		t.Fatalf("GetActivityStreamByType cadence: %v", err)
	}
	if cad == nil {
		t.Fatal("expected cadence stream, got nil")
	}
}

func TestGetActivityStreamByTypeNotFound(t *testing.T) {
	db := newTestDB(t)
	activityID := testActivity(t, db)

	got, err := db.GetActivityStreamByType(activityID, "heartrate")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for non-existent stream, got %+v", got)
	}
}

func TestGetActivityStreamsEmpty(t *testing.T) {
	db := newTestDB(t)
	activityID := testActivity(t, db)

	streams, err := db.GetActivityStreams(activityID)
	if err != nil {
		t.Fatalf("GetActivityStreams: %v", err)
	}
	if streams != nil {
		t.Errorf("expected nil for no streams, got %d", len(streams))
	}
}

func TestDeleteActivityStreams(t *testing.T) {
	db := newTestDB(t)
	activityID := testActivity(t, db)

	hrData, _ := json.Marshal([]int{140, 145})
	cadData, _ := json.Marshal([]int{170, 175})

	if err := db.SaveActivityStream(activityID, "heartrate", hrData); err != nil {
		t.Fatalf("SaveActivityStream: %v", err)
	}
	if err := db.SaveActivityStream(activityID, "cadence", cadData); err != nil {
		t.Fatalf("SaveActivityStream: %v", err)
	}

	if err := db.DeleteActivityStreams(activityID); err != nil {
		t.Fatalf("DeleteActivityStreams: %v", err)
	}

	streams, err := db.GetActivityStreams(activityID)
	if err != nil {
		t.Fatalf("GetActivityStreams: %v", err)
	}
	if streams != nil {
		t.Errorf("expected nil after delete, got %d streams", len(streams))
	}
}

func TestSaveActivityStreamValidation(t *testing.T) {
	db := newTestDB(t)
	activityID := testActivity(t, db)

	if err := db.SaveActivityStream(activityID, "", json.RawMessage(`[1,2,3]`)); err == nil {
		t.Error("expected error for empty stream type")
	}

	if err := db.SaveActivityStream(activityID, "heartrate", nil); err == nil {
		t.Error("expected error for nil data")
	}

	if err := db.SaveActivityStream(activityID, "heartrate", json.RawMessage{}); err == nil {
		t.Error("expected error for empty data")
	}
}

func TestSaveActivityStreamForeignKeyConstraint(t *testing.T) {
	db := newTestDB(t)

	hrData, _ := json.Marshal([]int{140, 145})
	err := db.SaveActivityStream(99999, "heartrate", hrData)
	if err == nil {
		t.Error("expected error for non-existent activity_id (foreign key constraint)")
	}
}

func TestActivityStreamLargeDataset(t *testing.T) {
	db := newTestDB(t)
	activityID := testActivity(t, db)

	largeHR := make([]int, 36000)
	for i := range largeHR {
		largeHR[i] = 120 + (i % 60)
	}
	hrData, _ := json.Marshal(largeHR)

	if err := db.SaveActivityStream(activityID, "heartrate", hrData); err != nil {
		t.Fatalf("SaveActivityStream large dataset: %v", err)
	}

	hr, err := db.GetActivityStreamByType(activityID, "heartrate")
	if err != nil {
		t.Fatalf("GetActivityStreamByType: %v", err)
	}
	if hr == nil {
		t.Fatal("expected stream, got nil")
	}

	var retrieved []int
	if err := json.Unmarshal(hr.Data, &retrieved); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(retrieved) != 36000 {
		t.Errorf("retrieved len = %d, want 36000", len(retrieved))
	}
}
