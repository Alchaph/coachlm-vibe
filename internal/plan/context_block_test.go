package plan

import (
	"strings"
	"testing"
	"time"

	"coachlm/internal/storage"
)

func newTestPlanStorageForCtx(t *testing.T) *Storage {
	t.Helper()
	db, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("newTestDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewStorage(db)
}

func seedRaceAndPlan(t *testing.T, store *Storage, now time.Time) {
	t.Helper()
	r := &Race{
		ID:         "race_ctx",
		Name:       "Spring Marathon",
		DistanceKm: 42.195,
		RaceDate:   now.AddDate(0, 0, 60),
		Terrain:    TerrainRoad,
		Priority:   PriorityA,
	}
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}
	if err := store.SetActiveRace("race_ctx"); err != nil {
		t.Fatalf("SetActiveRace: %v", err)
	}

	mondayOffset := int(now.Weekday())
	if mondayOffset == 0 {
		mondayOffset = 7
	}
	mondayOffset--
	weekStart := now.AddDate(0, 0, -mondayOffset).Truncate(24 * time.Hour)

	p := &TrainingPlan{
		ID:          "plan_ctx",
		RaceID:      "race_ctx",
		GeneratedAt: now,
		LLMBackend:  "local",
		PromptHash:  "ctxhash",
		Weeks: []Week{
			{
				ID: "pc_w1", PlanID: "plan_ctx", WeekNumber: 1,
				WeekStart: weekStart,
				Sessions: []Session{
					{ID: "pc_w1_s1", WeekID: "pc_w1", DayOfWeek: 1, Type: SessionEasy, DurationMin: 45, DistanceKm: 8.0, HRZone: 2, Notes: "Easy run"},
					{ID: "pc_w1_s2", WeekID: "pc_w1", DayOfWeek: 3, Type: SessionTempo, DurationMin: 40, DistanceKm: 7.0, HRZone: 3, Notes: "Tempo"},
					{ID: "pc_w1_s3", WeekID: "pc_w1", DayOfWeek: 6, Type: SessionRest, DurationMin: 0, Notes: "Rest day"},
				},
			},
			{
				ID: "pc_w2", PlanID: "plan_ctx", WeekNumber: 2,
				WeekStart: weekStart.AddDate(0, 0, 7),
				Sessions: []Session{
					{ID: "pc_w2_s1", WeekID: "pc_w2", DayOfWeek: 1, Type: SessionInterval, DurationMin: 50, DistanceKm: 10.0, HRZone: 4, Notes: "Intervals"},
					{ID: "pc_w2_s2", WeekID: "pc_w2", DayOfWeek: 7, Type: SessionLongRun, DurationMin: 90, DistanceKm: 18.0, HRZone: 2, Notes: "Long run"},
				},
			},
		},
	}
	if err := store.SavePlan(p); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
}

func TestPlanBlock_WithActivePlan(t *testing.T) {
	store := newTestPlanStorageForCtx(t)
	now := time.Now()
	seedRaceAndPlan(t, store, now)

	block := PlanBlock(store, now)
	if block == "" {
		t.Fatal("expected non-empty plan block")
	}
	if !strings.Contains(block, "## Active Training Plan") {
		t.Error("missing header")
	}
	if !strings.Contains(block, "Spring Marathon") {
		t.Error("missing race name")
	}
	if !strings.Contains(block, "42.2 km") {
		t.Error("missing distance")
	}
	if !strings.Contains(block, "road") {
		t.Error("missing terrain")
	}
	if !strings.Contains(block, "Weeks remaining:") {
		t.Error("missing weeks remaining")
	}
	if !strings.Contains(block, "This Week") {
		t.Error("missing current week section")
	}
}

func TestPlanBlock_NoActiveRace(t *testing.T) {
	store := newTestPlanStorageForCtx(t)
	block := PlanBlock(store, time.Now())
	if block != "" {
		t.Errorf("expected empty block with no active race, got: %q", block)
	}
}

func TestPlanBlock_ActiveRaceNoPlan(t *testing.T) {
	store := newTestPlanStorageForCtx(t)
	r := &Race{
		ID:         "race_noplan",
		Name:       "Solo Race",
		DistanceKm: 10,
		RaceDate:   time.Now().AddDate(0, 0, 30),
		Terrain:    TerrainRoad,
		Priority:   PriorityA,
	}
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}
	if err := store.SetActiveRace("race_noplan"); err != nil {
		t.Fatalf("SetActiveRace: %v", err)
	}

	block := PlanBlock(store, time.Now())
	if block != "" {
		t.Errorf("expected empty block with no plan, got: %q", block)
	}
}

func TestPlanBlock_TokenBudget(t *testing.T) {
	store := newTestPlanStorageForCtx(t)
	now := time.Now()
	seedRaceAndPlan(t, store, now)

	block := PlanBlock(store, now)
	tokens := estimateTokens(block)
	if tokens > maxPlanBlockTokens {
		t.Errorf("plan block %d tokens exceeds budget %d", tokens, maxPlanBlockTokens)
	}
}

func TestPlanBlock_SessionFormatting(t *testing.T) {
	store := newTestPlanStorageForCtx(t)
	now := time.Now()
	seedRaceAndPlan(t, store, now)

	block := PlanBlock(store, now)
	if !strings.Contains(block, "easy") {
		t.Error("missing easy session type")
	}
	if !strings.Contains(block, "45min") {
		t.Error("missing duration")
	}
	if !strings.Contains(block, "Rest") {
		t.Error("missing rest day")
	}
}

func TestPlanBlock_CompletedSessionStatus(t *testing.T) {
	store := newTestPlanStorageForCtx(t)
	now := time.Now()
	seedRaceAndPlan(t, store, now)

	dur := 48
	if err := store.UpdateSessionStatus("pc_w1_s1", StatusCompleted, ActualMetrics{DurationMin: &dur}); err != nil {
		t.Fatalf("UpdateSessionStatus: %v", err)
	}

	block := PlanBlock(store, now)
	if !strings.Contains(block, "[completed]") {
		t.Error("expected [completed] status marker in plan block")
	}
}

func TestFindCurrentWeek_InRange(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC) // Thursday
	monday := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	weeks := []Week{
		{WeekNumber: 1, WeekStart: monday},
		{WeekNumber: 2, WeekStart: monday.AddDate(0, 0, 7)},
		{WeekNumber: 3, WeekStart: monday.AddDate(0, 0, 14)},
	}

	got := findCurrentWeek(weeks, now)
	if got != 1 {
		t.Errorf("findCurrentWeek = %d, want 1", got)
	}
}

func TestFindCurrentWeek_SecondWeek(t *testing.T) {
	monday := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	now := monday.AddDate(0, 0, 9) // Wednesday of week 2

	weeks := []Week{
		{WeekNumber: 1, WeekStart: monday},
		{WeekNumber: 2, WeekStart: monday.AddDate(0, 0, 7)},
		{WeekNumber: 3, WeekStart: monday.AddDate(0, 0, 14)},
	}

	got := findCurrentWeek(weeks, now)
	if got != 2 {
		t.Errorf("findCurrentWeek = %d, want 2", got)
	}
}

func TestFindCurrentWeek_PastAllWeeks(t *testing.T) {
	monday := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	now := monday.AddDate(0, 0, 30)

	weeks := []Week{
		{WeekNumber: 1, WeekStart: monday},
		{WeekNumber: 2, WeekStart: monday.AddDate(0, 0, 7)},
	}

	got := findCurrentWeek(weeks, now)
	if got != 2 {
		t.Errorf("findCurrentWeek past all = %d, want 2 (last week)", got)
	}
}

func TestFindCurrentWeek_BeforeFirstWeek(t *testing.T) {
	monday := time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	weeks := []Week{
		{WeekNumber: 1, WeekStart: monday},
		{WeekNumber: 2, WeekStart: monday.AddDate(0, 0, 7)},
	}

	got := findCurrentWeek(weeks, now)
	if got != 1 {
		t.Errorf("findCurrentWeek before first = %d, want 1", got)
	}
}

func TestFindCurrentWeek_Empty(t *testing.T) {
	got := findCurrentWeek(nil, time.Now())
	if got != 1 {
		t.Errorf("findCurrentWeek empty = %d, want 1", got)
	}
}

func TestWeeksToRaceDisplay_WeeksAway(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	raceDate := now.AddDate(0, 0, 42)
	got := WeeksToRaceDisplay(now, raceDate)
	if !strings.Contains(got, "weeks to race") {
		t.Errorf("expected 'weeks to race', got: %q", got)
	}
}

func TestWeeksToRaceDisplay_RaceDay(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	got := WeeksToRaceDisplay(now, now)
	if got != "Race day!" {
		t.Errorf("expected 'Race day!', got: %q", got)
	}
}

func TestWeeksToRaceDisplay_PastRace(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	raceDate := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	got := WeeksToRaceDisplay(now, raceDate)
	if got != "Race day!" {
		t.Errorf("expected 'Race day!' for past race, got: %q", got)
	}
}

func TestWeeksToRaceDisplay_DaysOnly(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	raceDate := now.AddDate(0, 0, 5)
	got := WeeksToRaceDisplay(now, raceDate)
	if !strings.Contains(got, "day(s) to race") {
		t.Errorf("expected 'day(s) to race' for <1 week, got: %q", got)
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"a", 1},
		{"abcd", 1},
		{"abcde", 2},
		{strings.Repeat("x", 100), 25},
	}
	for _, tt := range tests {
		got := estimateTokens(tt.input)
		if got != tt.want {
			t.Errorf("estimateTokens(%d chars) = %d, want %d", len(tt.input), got, tt.want)
		}
	}
}

func TestFindWeek(t *testing.T) {
	weeks := []Week{
		{WeekNumber: 1},
		{WeekNumber: 2},
		{WeekNumber: 3},
	}

	got := findWeek(weeks, 2)
	if got == nil {
		t.Fatal("expected week 2, got nil")
	}
	if got.WeekNumber != 2 {
		t.Errorf("WeekNumber = %d, want 2", got.WeekNumber)
	}

	got = findWeek(weeks, 99)
	if got != nil {
		t.Error("expected nil for nonexistent week")
	}
}

func TestWriteWeekSessions_DayNames(t *testing.T) {
	w := &Week{
		Sessions: []Session{
			{DayOfWeek: 1, Type: SessionEasy, DurationMin: 45, DistanceKm: 8.0, HRZone: 2},
			{DayOfWeek: 3, Type: SessionRest, DurationMin: 0},
			{DayOfWeek: 7, Type: SessionLongRun, DurationMin: 90, DistanceKm: 18.0},
		},
	}
	var sb strings.Builder
	writeWeekSessions(&sb, w)
	result := sb.String()

	if !strings.Contains(result, "Mon:") {
		t.Error("missing Mon day name")
	}
	if !strings.Contains(result, "Wed: Rest") {
		t.Error("missing Wed rest day")
	}
	if !strings.Contains(result, "Sun:") {
		t.Error("missing Sun day name")
	}
	if !strings.Contains(result, "Z2") {
		t.Error("missing HR zone marker")
	}
	if !strings.Contains(result, "8.0km") {
		t.Error("missing distance")
	}
}
