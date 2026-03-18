package context

import (
	"strings"
	"testing"

	"coachlm/internal/storage"
)

func TestFormatStatsBlock_Nil(t *testing.T) {
	result := FormatStatsBlock(nil)
	if result != "" {
		t.Errorf("expected empty string for nil stats, got %q", result)
	}
}

func TestFormatStatsBlock_AllZero(t *testing.T) {
	result := FormatStatsBlock(&storage.AthleteStats{})
	if result != "" {
		t.Errorf("expected empty string for zero stats, got %q", result)
	}
}

func TestFormatStatsBlock_FullStats(t *testing.T) {
	stats := &storage.AthleteStats{
		RecentRunCount:      5,
		RecentRunDistance:   45000.0,
		RecentRunMovingTime: 14400,
		RecentRunElevation:  300.0,
		YTDRunCount:         42,
		YTDRunDistance:      400000.0,
		YTDRunMovingTime:    120000,
		YTDRunElevation:     2500.0,
		AllRunCount:         500,
		AllRunDistance:      5000000.0,
		AllRunMovingTime:    1500000,
		AllRunElevation:     30000.0,
	}

	result := FormatStatsBlock(stats)

	if !strings.HasPrefix(result, "## Training Load Statistics") {
		t.Error("missing heading")
	}
	if !strings.Contains(result, "Recent (4 weeks): 5 runs, 45.0 km") {
		t.Errorf("missing recent stats, got:\n%s", result)
	}
	if !strings.Contains(result, "4h,") {
		t.Errorf("missing recent duration, got:\n%s", result)
	}
	if !strings.Contains(result, "Year to date: 42 runs, 400.0 km") {
		t.Errorf("missing YTD stats, got:\n%s", result)
	}
	if !strings.Contains(result, "All time: 500 runs, 5000.0 km") {
		t.Errorf("missing all-time stats, got:\n%s", result)
	}
	if !strings.Contains(result, "300 m elevation") {
		t.Errorf("missing elevation, got:\n%s", result)
	}
}

func TestFormatStatsBlock_PartialStats(t *testing.T) {
	stats := &storage.AthleteStats{
		RecentRunCount:      3,
		RecentRunDistance:   25000.0,
		RecentRunMovingTime: 7200,
		RecentRunElevation:  100.0,
	}

	result := FormatStatsBlock(stats)

	if !strings.Contains(result, "Recent (4 weeks)") {
		t.Error("missing recent section")
	}
	if strings.Contains(result, "Year to date") {
		t.Error("YTD should not appear when count is 0")
	}
	if strings.Contains(result, "All time") {
		t.Error("all-time should not appear when count is 0")
	}
}

func TestFormatStatsDuration(t *testing.T) {
	tests := []struct {
		secs int
		want string
	}{
		{0, "0h"},
		{-1, "0h"},
		{300, "5m"},
		{3600, "1h"},
		{3660, "1h01m"},
		{14400, "4h"},
		{120000, "33h20m"},
	}
	for _, tt := range tests {
		got := formatStatsDuration(tt.secs)
		if got != tt.want {
			t.Errorf("formatStatsDuration(%d) = %q, want %q", tt.secs, got, tt.want)
		}
	}
}
