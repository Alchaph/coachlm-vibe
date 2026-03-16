package context

import (
	"strings"
	"testing"
	"time"

	"coachlm/internal/storage"
)

func makeActivity(name, actType string, daysAgo int, now time.Time, distMeters float64, durSecs, paceSecs, avgHR int) storage.Activity {
	return storage.Activity{
		Name:         name,
		ActivityType: actType,
		StartDate:    now.AddDate(0, 0, -daysAgo),
		Distance:     distMeters,
		DurationSecs: durSecs,
		AvgPaceSecs:  paceSecs,
		AvgHR:        avgHR,
	}
}

func TestFormatTrainingSummary_NoActivities(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)
	got := FormatTrainingSummary(nil, now)
	if got != "No recent training data." {
		t.Errorf("expected 'No recent training data.', got %q", got)
	}

	got = FormatTrainingSummary([]storage.Activity{}, now)
	if got != "No recent training data." {
		t.Errorf("expected 'No recent training data.' for empty slice, got %q", got)
	}
}

func TestFormatTrainingSummary_AllActivitiesOlderThan28Days(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)
	activities := []storage.Activity{
		makeActivity("Old Run", "Run", 30, now, 10000, 3000, 300, 150),
		makeActivity("Ancient Run", "Run", 60, now, 5000, 1500, 300, 140),
	}
	got := FormatTrainingSummary(activities, now)
	if got != "No recent training data." {
		t.Errorf("expected 'No recent training data.' for old activities, got %q", got)
	}
}

func TestFormatTrainingSummary_Full4Weeks(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC) // Monday

	activities := []storage.Activity{
		// Week 0 (this week): 0-6 days ago
		makeActivity("Morning Run", "Run", 0, now, 10000, 3000, 300, 145),
		makeActivity("Tempo Run", "Run", 1, now, 8000, 2160, 270, 165),

		// Week 1 (last week): 7-13 days ago
		makeActivity("Easy Run", "Run", 7, now, 5000, 1500, 300, 130),
		makeActivity("Interval Run", "Run", 7, now, 10000, 3000, 300, 160),
		makeActivity("Recovery Run", "Run", 9, now, 6000, 2100, 350, 125),

		// Week 2 (2 weeks ago): 14-20 days ago
		makeActivity("Sunday Long Run", "Run", 14, now, 21000, 6300, 300, 150),
		makeActivity("Wednesday Intervals", "Run", 17, now, 8000, 2040, 255, 170),
		makeActivity("Easy Jog", "Run", 18, now, 5000, 1800, 360, 125),
		makeActivity("Tempo", "Run", 19, now, 10000, 2700, 270, 160),
		makeActivity("Shakeout", "Run", 20, now, 3000, 1080, 360, 120),

		// Week 3 (3 weeks ago): 21-27 days ago
		makeActivity("Long Run", "Run", 21, now, 18000, 5400, 300, 148),
		makeActivity("Tempo Run", "Run", 23, now, 10000, 2700, 270, 162),
		makeActivity("Easy Run", "Run", 24, now, 6000, 2100, 350, 128),
		makeActivity("Intervals", "Run", 25, now, 8000, 2160, 270, 168),
		makeActivity("Recovery", "Run", 26, now, 5000, 1800, 360, 122),
		makeActivity("Park Run", "Run", 27, now, 5000, 1200, 240, 175),
	}

	got := FormatTrainingSummary(activities, now)

	if !strings.Contains(got, "## Training Summary (Last 4 Weeks)") {
		t.Error("missing header")
	}

	if !strings.Contains(got, "### This Week (detailed)") {
		t.Error("missing week 0 section")
	}
	if !strings.Contains(got, "Morning Run") {
		t.Error("missing detailed activity name")
	}
	if !strings.Contains(got, "10.0 km") {
		t.Error("missing distance in km")
	}
	if !strings.Contains(got, "5:00/km") {
		t.Error("missing pace")
	}
	if !strings.Contains(got, "HR 145") {
		t.Error("missing HR")
	}

	if !strings.Contains(got, "### Last Week (daily summary)") {
		t.Error("missing week 1 section")
	}
	if !strings.Contains(got, "2 runs") {
		t.Error("missing aggregated run count for day with 2 runs")
	}

	if !strings.Contains(got, "### 2 Weeks Ago (key sessions)") {
		t.Error("missing week 2 section")
	}
	if !strings.Contains(got, "Longest: Sunday Long Run") {
		t.Error("missing longest run")
	}
	if !strings.Contains(got, "Fastest: Wednesday Intervals") {
		t.Error("missing fastest run")
	}
	if !strings.Contains(got, "Highest mileage day") {
		t.Error("missing highest mileage day")
	}

	if !strings.Contains(got, "### 3 Weeks Ago (totals)") {
		t.Error("missing week 3 section")
	}
	if !strings.Contains(got, "52.0 km") {
		t.Errorf("missing total distance for week 3, got:\n%s", got)
	}
	if !strings.Contains(got, "6 runs") {
		t.Errorf("missing total runs for week 3, got:\n%s", got)
	}
	if !strings.Contains(got, "avg") {
		t.Error("missing avg pace in week 3 totals")
	}
}

func TestFormatTrainingSummary_FewerThan4Weeks(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("Today Run", "Run", 0, now, 10000, 3000, 300, 145),
		makeActivity("Last Week Run", "Run", 8, now, 8000, 2400, 300, 150),
	}

	got := FormatTrainingSummary(activities, now)

	if !strings.Contains(got, "### This Week (detailed)") {
		t.Error("missing week 0")
	}
	if !strings.Contains(got, "### Last Week (daily summary)") {
		t.Error("missing week 1")
	}
	if strings.Contains(got, "### 2 Weeks Ago") {
		t.Error("week 2 should not appear when no activities exist for it")
	}
	if strings.Contains(got, "### 3 Weeks Ago") {
		t.Error("week 3 should not appear when no activities exist for it")
	}
}

func TestFormatTrainingSummary_100PlusActivitiesInOneWeek(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	var activities []storage.Activity
	for i := 0; i < 105; i++ {
		daysAgo := i % 7
		activities = append(activities, makeActivity(
			"Run", "Run", daysAgo, now, 5000, 1500, 300, 140,
		))
	}

	got := FormatTrainingSummary(activities, now)

	if !strings.Contains(got, "### This Week (detailed)") {
		t.Error("missing week 0 section with 100+ activities")
	}

	count := strings.Count(got, "- ")
	if count < 105 {
		t.Errorf("expected at least 105 detail lines, got %d", count)
	}
}

func TestFormatTrainingSummary_MissingMetrics(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("No HR Run", "Run", 1, now, 10000, 3000, 300, 0),
		makeActivity("No Pace Run", "Run", 2, now, 8000, 2400, 0, 150),
		{
			Name:         "Nothing Run",
			ActivityType: "Run",
			StartDate:    now.AddDate(0, 0, -3),
			Distance:     5000,
			DurationSecs: 1500,
			AvgPaceSecs:  0,
			AvgHR:        0,
			AvgCadence:   0,
		},
	}

	got := FormatTrainingSummary(activities, now)

	if strings.Contains(got, "HR 0") {
		t.Error("should not display HR 0")
	}

	noHRLine := ""
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "No HR Run") {
			noHRLine = line
			break
		}
	}
	if noHRLine != "" && strings.Contains(noHRLine, "HR 0") {
		t.Errorf("No HR Run line should not contain 'HR 0', got: %s", noHRLine)
	}

	noPaceLine := ""
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "No Pace Run") {
			noPaceLine = line
			break
		}
	}
	if noPaceLine != "" && strings.Contains(noPaceLine, "/km") {
		t.Errorf("No Pace Run line should not contain pace, got: %s", noPaceLine)
	}
}

func TestFormatTrainingSummary_NonRunningActivities(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("Morning Run", "Run", 1, now, 10000, 3000, 300, 145),
		makeActivity("Bike Ride", "Ride", 2, now, 30000, 3600, 0, 135),
		makeActivity("Pool Session", "Swim", 3, now, 2000, 1800, 0, 120),
	}

	got := FormatTrainingSummary(activities, now)

	if !strings.Contains(got, "Bike Ride [Ride]") {
		t.Errorf("expected non-run activity labeled with type, got:\n%s", got)
	}
	if !strings.Contains(got, "Pool Session [Swim]") {
		t.Errorf("expected swim labeled with type, got:\n%s", got)
	}
	if strings.Contains(got, "Morning Run [Run]") {
		t.Error("Run activities should NOT have type label")
	}
}

func TestFormatTrainingSummary_BoundaryActivities(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("Day 0", "Run", 0, now, 5000, 1500, 300, 140),
		makeActivity("Day 6", "Run", 6, now, 5000, 1500, 300, 140),
		makeActivity("Day 7", "Run", 7, now, 5000, 1500, 300, 140),
		makeActivity("Day 13", "Run", 13, now, 5000, 1500, 300, 140),
		makeActivity("Day 14", "Run", 14, now, 5000, 1500, 300, 140),
		makeActivity("Day 20", "Run", 20, now, 5000, 1500, 300, 140),
		makeActivity("Day 21", "Run", 21, now, 5000, 1500, 300, 140),
		makeActivity("Day 27", "Run", 27, now, 5000, 1500, 300, 140),
		makeActivity("Day 28", "Run", 28, now, 5000, 1500, 300, 140),
	}

	got := FormatTrainingSummary(activities, now)

	// Day 0 and Day 6 should be in week 0 (detailed)
	if !strings.Contains(got, "Day 0") || !strings.Contains(got, "Day 6") {
		t.Error("Days 0 and 6 should be in This Week")
	}

	// Day 7 and Day 13 should be in week 1 (daily summary)
	if !strings.Contains(got, "### Last Week (daily summary)") {
		t.Error("missing week 1 section")
	}

	// Day 14 and Day 20 should be in week 2 (key sessions)
	if !strings.Contains(got, "### 2 Weeks Ago (key sessions)") {
		t.Error("missing week 2 section")
	}

	// Day 21 and Day 27 should be in week 3 (totals)
	if !strings.Contains(got, "### 3 Weeks Ago (totals)") {
		t.Error("missing week 3 section")
	}

	// Day 28 should be excluded
	if strings.Contains(got, "Day 28") {
		t.Error("Day 28 should be excluded (>= 28 days is out of range)")
	}
}

func TestFormatTrainingSummary_PartialCurrentWeek(t *testing.T) {
	now := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC) // Wednesday

	activities := []storage.Activity{
		makeActivity("Mon Run", "Run", 2, now, 10000, 3000, 300, 145),
		makeActivity("Wed Run", "Run", 0, now, 8000, 2400, 300, 150),
	}

	got := FormatTrainingSummary(activities, now)

	if !strings.Contains(got, "### This Week (detailed)") {
		t.Error("missing this week section for partial week")
	}
	if !strings.Contains(got, "Mon Run") {
		t.Error("missing Monday run")
	}
	if !strings.Contains(got, "Wed Run") {
		t.Error("missing Wednesday run")
	}
}

func TestFormatTrainingSummary_DailySummaryAggregation(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("AM Run", "Run", 8, now, 10000, 3000, 300, 145),
		makeActivity("PM Run", "Run", 8, now, 5000, 1500, 300, 140),
	}

	got := FormatTrainingSummary(activities, now)

	if !strings.Contains(got, "2 runs") {
		t.Errorf("expected '2 runs' for same-day aggregate, got:\n%s", got)
	}
	if !strings.Contains(got, "15.0 km") {
		t.Errorf("expected aggregated distance '15.0 km', got:\n%s", got)
	}
}

func TestFormatTrainingSummary_KeySessionsAllZeroPace(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("Walk A", "Run", 15, now, 5000, 3000, 0, 100),
		makeActivity("Walk B", "Run", 16, now, 3000, 2000, 0, 95),
	}

	got := FormatTrainingSummary(activities, now)

	if !strings.Contains(got, "Longest: Walk A") {
		t.Errorf("expected Walk A as longest, got:\n%s", got)
	}
	if strings.Contains(got, "Fastest:") {
		t.Error("should not show Fastest when all paces are 0")
	}
}

func TestFormatTrainingSummary_WeeklyTotalsAvgPace(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("Run A", "Run", 22, now, 10000, 3000, 300, 145),
		makeActivity("Run B", "Run", 24, now, 10000, 2700, 270, 160),
	}

	got := FormatTrainingSummary(activities, now)

	if !strings.Contains(got, "### 3 Weeks Ago (totals)") {
		t.Error("missing week 3")
	}
	if !strings.Contains(got, "20.0 km") {
		t.Errorf("expected 20.0 km total, got:\n%s", got)
	}
	if !strings.Contains(got, "2 runs") {
		t.Errorf("expected 2 runs, got:\n%s", got)
	}
	// avg pace of 300 and 270 = 285 = 4:45/km
	if !strings.Contains(got, "avg 4:45/km") {
		t.Errorf("expected avg 4:45/km, got:\n%s", got)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		secs int
		want string
	}{
		{0, "0:00"},
		{59, "0:59"},
		{60, "1:00"},
		{3599, "59:59"},
		{3600, "1:00:00"},
		{3661, "1:01:01"},
		{7200, "2:00:00"},
		{-5, "0:00"},
	}
	for _, tt := range tests {
		got := FormatDuration(tt.secs)
		if got != tt.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", tt.secs, got, tt.want)
		}
	}
}

func TestFormatTrainingSummary_FutureActivitiesExcluded(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("Today", "Run", 0, now, 5000, 1500, 300, 140),
		{
			Name:         "Tomorrow",
			ActivityType: "Run",
			StartDate:    now.AddDate(0, 0, 1),
			Distance:     5000,
			DurationSecs: 1500,
			AvgPaceSecs:  300,
			AvgHR:        140,
		},
	}

	got := FormatTrainingSummary(activities, now)

	if strings.Contains(got, "Tomorrow") {
		t.Error("future activities should be excluded")
	}
	if !strings.Contains(got, "Today") {
		t.Error("today's activities should be included")
	}
}

func TestFormatTrainingSummary_Deterministic(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("Run A", "Run", 1, now, 10000, 3000, 300, 145),
		makeActivity("Run B", "Run", 2, now, 8000, 2400, 300, 150),
		makeActivity("Run C", "Run", 8, now, 12000, 3600, 300, 155),
	}

	first := FormatTrainingSummary(activities, now)
	second := FormatTrainingSummary(activities, now)
	if first != second {
		t.Errorf("non-deterministic output:\nfirst:  %q\nsecond: %q", first, second)
	}
}

func TestFormatTrainingSummary_DailySummarySingleRun(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("Solo Run", "Run", 10, now, 8000, 2400, 300, 150),
	}

	got := FormatTrainingSummary(activities, now)

	if !strings.Contains(got, "1 run,") {
		t.Errorf("expected '1 run,' (singular), got:\n%s", got)
	}
}

func TestWeekBucket(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		daysAgo  int
		wantWeek int
	}{
		{"today", 0, 0},
		{"6 days ago", 6, 0},
		{"7 days ago", 7, 1},
		{"13 days ago", 13, 1},
		{"14 days ago", 14, 2},
		{"20 days ago", 20, 2},
		{"21 days ago", 21, 3},
		{"27 days ago", 27, 3},
		{"28 days ago", 28, -1},
		{"100 days ago", 100, -1},
		{"future", -1, -1},
	}

	for _, tt := range tests {
		actTime := now.AddDate(0, 0, -tt.daysAgo)
		got := weekBucket(actTime, now)
		if got != tt.wantWeek {
			t.Errorf("weekBucket(%s, daysAgo=%d) = %d, want %d", tt.name, tt.daysAgo, got, tt.wantWeek)
		}
	}
}

func TestFormatTrainingSummary_OnlyWeek3(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("Old Run", "Run", 22, now, 15000, 4500, 300, 145),
	}

	got := FormatTrainingSummary(activities, now)

	if strings.Contains(got, "### This Week") {
		t.Error("should not have This Week section")
	}
	if strings.Contains(got, "### Last Week") {
		t.Error("should not have Last Week section")
	}
	if strings.Contains(got, "### 2 Weeks Ago") {
		t.Error("should not have 2 Weeks Ago section")
	}
	if !strings.Contains(got, "### 3 Weeks Ago (totals)") {
		t.Error("missing week 3 section")
	}
	if !strings.Contains(got, "15.0 km") {
		t.Error("missing total distance")
	}
}

func TestFormatTrainingSummary_NonRunInKeySessions(t *testing.T) {
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)

	activities := []storage.Activity{
		makeActivity("Long Ride", "Ride", 15, now, 50000, 7200, 0, 135),
		makeActivity("Short Run", "Run", 16, now, 5000, 1500, 300, 145),
	}

	got := FormatTrainingSummary(activities, now)

	if !strings.Contains(got, "Longest: Long Ride") {
		t.Errorf("expected Long Ride as longest, got:\n%s", got)
	}
	if !strings.Contains(got, "Fastest: Short Run") {
		t.Errorf("expected Short Run as fastest, got:\n%s", got)
	}
}
