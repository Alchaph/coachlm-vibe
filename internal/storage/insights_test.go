package storage

import (
	"fmt"
	"testing"
)

func TestSaveInsightAndGetInsights(t *testing.T) {
	db := newTestDB(t)

	insight, err := db.SaveInsight("Run easy days truly easy", "session-1")
	if err != nil {
		t.Fatalf("SaveInsight: %v", err)
	}
	if insight.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if insight.Content != "Run easy days truly easy" {
		t.Errorf("content = %q, want %q", insight.Content, "Run easy days truly easy")
	}
	if insight.SourceSessionID != "session-1" {
		t.Errorf("source_session_id = %q, want %q", insight.SourceSessionID, "session-1")
	}
	if insight.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	all, err := db.GetInsights()
	if err != nil {
		t.Fatalf("GetInsights: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("len = %d, want 1", len(all))
	}
	if all[0].Content != insight.Content {
		t.Errorf("round-trip content mismatch: got %q", all[0].Content)
	}
}

func TestDeleteInsight(t *testing.T) {
	db := newTestDB(t)

	insight, err := db.SaveInsight("Tempo runs build lactate threshold", "s2")
	if err != nil {
		t.Fatalf("SaveInsight: %v", err)
	}

	if err := db.DeleteInsight(insight.ID); err != nil {
		t.Fatalf("DeleteInsight: %v", err)
	}

	all, err := db.GetInsights()
	if err != nil {
		t.Fatalf("GetInsights: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 insights after delete, got %d", len(all))
	}
}

func TestDeleteInsightNotFound(t *testing.T) {
	db := newTestDB(t)

	err := db.DeleteInsight(9999)
	if err == nil {
		t.Fatal("expected error deleting non-existent insight")
	}
}

func TestInsightExistsTrue(t *testing.T) {
	db := newTestDB(t)

	_, err := db.SaveInsight("Long runs build aerobic base", "s1")
	if err != nil {
		t.Fatalf("SaveInsight: %v", err)
	}

	exists, err := db.InsightExists("Long runs build aerobic base")
	if err != nil {
		t.Fatalf("InsightExists: %v", err)
	}
	if !exists {
		t.Error("expected InsightExists to return true for duplicate")
	}
}

func TestInsightExistsFalse(t *testing.T) {
	db := newTestDB(t)

	exists, err := db.InsightExists("This was never saved")
	if err != nil {
		t.Fatalf("InsightExists: %v", err)
	}
	if exists {
		t.Error("expected InsightExists to return false for non-existent content")
	}
}

func TestGetInsightsOrdering(t *testing.T) {
	db := newTestDB(t)

	contents := []string{"first", "second", "third"}
	for _, c := range contents {
		if _, err := db.SaveInsight(c, "s1"); err != nil {
			t.Fatalf("SaveInsight(%q): %v", c, err)
		}
	}

	all, err := db.GetInsights()
	if err != nil {
		t.Fatalf("GetInsights: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("len = %d, want 3", len(all))
	}
	for i, want := range contents {
		if all[i].Content != want {
			t.Errorf("insight[%d].Content = %q, want %q", i, all[i].Content, want)
		}
	}
}

func TestManyInsights(t *testing.T) {
	db := newTestDB(t)

	const n = 120
	for i := 0; i < n; i++ {
		_, err := db.SaveInsight(fmt.Sprintf("insight-%d", i), "bulk")
		if err != nil {
			t.Fatalf("SaveInsight #%d: %v", i, err)
		}
	}

	all, err := db.GetInsights()
	if err != nil {
		t.Fatalf("GetInsights: %v", err)
	}
	if len(all) != n {
		t.Errorf("len = %d, want %d", len(all), n)
	}
	if all[0].Content != "insight-0" {
		t.Errorf("first insight = %q, want %q", all[0].Content, "insight-0")
	}
	if all[n-1].Content != fmt.Sprintf("insight-%d", n-1) {
		t.Errorf("last insight = %q, want %q", all[n-1].Content, fmt.Sprintf("insight-%d", n-1))
	}
}

func TestSaveInsightEmptyContent(t *testing.T) {
	db := newTestDB(t)

	cases := []string{"", "   ", "\t\n"}
	for _, c := range cases {
		_, err := db.SaveInsight(c, "s1")
		if err == nil {
			t.Errorf("expected error for empty content %q", c)
		}
	}
}

func TestGetInsightsEmpty(t *testing.T) {
	db := newTestDB(t)

	all, err := db.GetInsights()
	if err != nil {
		t.Fatalf("GetInsights: %v", err)
	}
	if all == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(all) != 0 {
		t.Errorf("expected 0 insights, got %d", len(all))
	}
}
