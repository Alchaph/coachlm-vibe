package plan

import (
	"strings"
	"testing"
	"time"
)

func TestParseLLMResponse_ValidJSON(t *testing.T) {
	raw := `{"weeks":[{"week_number":1,"sessions":[{"day_of_week":1,"type":"easy","duration_min":45,"notes":"Easy run"}]}]}`
	resp, err := parseLLMResponse(raw)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(resp.Weeks) != 1 {
		t.Fatalf("expected 1 week, got %d", len(resp.Weeks))
	}
	if resp.Weeks[0].WeekNumber != 1 {
		t.Errorf("week_number = %d, want 1", resp.Weeks[0].WeekNumber)
	}
	if len(resp.Weeks[0].Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(resp.Weeks[0].Sessions))
	}
	if resp.Weeks[0].Sessions[0].Type != "easy" {
		t.Errorf("session type = %q, want %q", resp.Weeks[0].Sessions[0].Type, "easy")
	}
}

func TestParseLLMResponse_WithMarkdownFences(t *testing.T) {
	raw := "```json\n{\"weeks\":[{\"week_number\":1,\"sessions\":[{\"day_of_week\":1,\"type\":\"tempo\",\"duration_min\":30,\"notes\":\"\"}]}]}\n```"
	resp, err := parseLLMResponse(raw)
	if err != nil {
		t.Fatalf("expected no error after fence strip, got: %v", err)
	}
	if len(resp.Weeks) != 1 {
		t.Fatalf("expected 1 week, got %d", len(resp.Weeks))
	}
}

func TestParseLLMResponse_FencesNolabel(t *testing.T) {
	raw := "```\n{\"weeks\":[{\"week_number\":1,\"sessions\":[{\"day_of_week\":2,\"type\":\"rest\",\"duration_min\":0,\"notes\":\"Rest day\"}]}]}\n```"
	resp, err := parseLLMResponse(raw)
	if err != nil {
		t.Fatalf("expected no error for fences without json label, got: %v", err)
	}
	if len(resp.Weeks) != 1 {
		t.Fatal("expected 1 week")
	}
}

func TestParseLLMResponse_InvalidJSON(t *testing.T) {
	raw := "This is not JSON at all"
	_, err := parseLLMResponse(raw)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "not valid JSON") {
		t.Errorf("error message should mention invalid JSON, got: %v", err)
	}
}

func TestParseLLMResponse_ZeroWeeks(t *testing.T) {
	raw := `{"weeks":[]}`
	_, err := parseLLMResponse(raw)
	if err == nil {
		t.Error("expected error for zero weeks")
	}
	if !strings.Contains(err.Error(), "zero weeks") {
		t.Errorf("error should mention zero weeks, got: %v", err)
	}
}

func TestParseLLMResponse_ZeroWeeksAfterFenceStrip(t *testing.T) {
	raw := "```json\n{\"weeks\":[]}\n```"
	_, err := parseLLMResponse(raw)
	if err == nil {
		t.Error("expected error for zero weeks after fence strip")
	}
}

func TestParseLLMResponse_WhitespaceAroundJSON(t *testing.T) {
	raw := "   \n  {\"weeks\":[{\"week_number\":1,\"sessions\":[{\"day_of_week\":1,\"type\":\"easy\",\"duration_min\":30,\"notes\":\"\"}]}]}  \n  "
	resp, err := parseLLMResponse(raw)
	if err != nil {
		t.Fatalf("expected no error with whitespace, got: %v", err)
	}
	if len(resp.Weeks) != 1 {
		t.Fatalf("expected 1 week, got %d", len(resp.Weeks))
	}
}

func TestParseLLMResponse_MultipleWeeks(t *testing.T) {
	raw := `{"weeks":[
		{"week_number":1,"sessions":[{"day_of_week":1,"type":"easy","duration_min":30,"notes":""}]},
		{"week_number":2,"sessions":[{"day_of_week":1,"type":"tempo","duration_min":40,"notes":""}]},
		{"week_number":3,"sessions":[{"day_of_week":1,"type":"intervals","duration_min":35,"notes":""}]}
	]}`
	resp, err := parseLLMResponse(raw)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(resp.Weeks) != 3 {
		t.Fatalf("expected 3 weeks, got %d", len(resp.Weeks))
	}
}

func TestParseLLMResponse_OptionalFields(t *testing.T) {
	raw := `{"weeks":[{"week_number":1,"sessions":[{"day_of_week":1,"type":"easy","duration_min":45,"distance_km":8.5,"hr_zone":2,"pace_low":5.0,"pace_high":5.5,"notes":"Zone 2 run"}]}]}`
	resp, err := parseLLMResponse(raw)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	sess := resp.Weeks[0].Sessions[0]
	if sess.DistanceKm != 8.5 {
		t.Errorf("distance_km = %f, want 8.5", sess.DistanceKm)
	}
	if sess.HRZone != 2 {
		t.Errorf("hr_zone = %d, want 2", sess.HRZone)
	}
	if sess.PaceLow != 5.0 {
		t.Errorf("pace_low = %f, want 5.0", sess.PaceLow)
	}
	if sess.PaceHigh != 5.5 {
		t.Errorf("pace_high = %f, want 5.5", sess.PaceHigh)
	}
}

func TestWeeksUntil_FutureDate(t *testing.T) {
	tests := []struct {
		name     string
		daysAway int
		want     int
	}{
		{"1 day", 1, 1},
		{"6 days", 6, 1},
		{"7 days", 7, 1},
		{"8 days", 8, 2},
		{"14 days", 14, 2},
		{"15 days", 15, 3},
		{"28 days", 28, 4},
		{"56 days", 56, 8},
	}
	now := fixedTime()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := now.AddDate(0, 0, tt.daysAway)
			got := weeksUntil(now, target)
			if got != tt.want {
				t.Errorf("weeksUntil(%d days) = %d, want %d", tt.daysAway, got, tt.want)
			}
		})
	}
}

func TestWeeksUntil_PastDate(t *testing.T) {
	now := fixedTime()
	target := now.AddDate(0, 0, -10)
	got := weeksUntil(now, target)
	if got != 1 {
		t.Errorf("weeksUntil(past date) = %d, want 1 (minimum)", got)
	}
}

func TestWeeksUntil_SameDay(t *testing.T) {
	now := fixedTime()
	got := weeksUntil(now, now)
	if got != 1 {
		t.Errorf("weeksUntil(same day) = %d, want 1 (minimum)", got)
	}
}

func TestTruncateNotes_Short(t *testing.T) {
	s := "Short note"
	got := truncateNotes(s, 300)
	if got != s {
		t.Errorf("truncateNotes = %q, want %q", got, s)
	}
}

func TestTruncateNotes_ExactLength(t *testing.T) {
	s := strings.Repeat("x", 300)
	got := truncateNotes(s, 300)
	if got != s {
		t.Error("truncateNotes should not truncate at exact length")
	}
}

func TestTruncateNotes_Long(t *testing.T) {
	s := strings.Repeat("x", 500)
	got := truncateNotes(s, 300)
	if len(got) != 300 {
		t.Errorf("truncateNotes length = %d, want 300", len(got))
	}
}

func TestTruncateNotes_Empty(t *testing.T) {
	got := truncateNotes("", 300)
	if got != "" {
		t.Errorf("truncateNotes empty = %q, want empty", got)
	}
}

func fixedTime() time.Time {
	return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
}
