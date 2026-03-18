package plan

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"coachlm/internal/storage"
)

func newTestPlanStorage(t *testing.T) (*Storage, *storage.DB) {
	t.Helper()
	db, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("newTestDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewStorage(db), db
}

func futureDate(daysFromNow int) time.Time {
	return time.Now().AddDate(0, 0, daysFromNow).Truncate(24 * time.Hour)
}

func testRace(id string) *Race {
	return &Race{
		ID:         id,
		Name:       "Test Marathon",
		DistanceKm: 42.195,
		RaceDate:   futureDate(90),
		Terrain:    TerrainRoad,
		Priority:   PriorityA,
	}
}

func TestCreateRace_Valid(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	r := testRace("race_1")
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}

	got, err := store.GetRace("race_1")
	if err != nil {
		t.Fatalf("GetRace: %v", err)
	}
	if got == nil {
		t.Fatal("expected race, got nil")
	}
	if got.Name != "Test Marathon" {
		t.Errorf("Name = %q, want %q", got.Name, "Test Marathon")
	}
	if got.DistanceKm != 42.195 {
		t.Errorf("DistanceKm = %f, want 42.195", got.DistanceKm)
	}
	if got.Terrain != TerrainRoad {
		t.Errorf("Terrain = %q, want %q", got.Terrain, TerrainRoad)
	}
}

func TestCreateRace_PastDate(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	r := &Race{
		ID:         "race_past",
		Name:       "Past Race",
		DistanceKm: 10,
		RaceDate:   time.Now().AddDate(0, 0, -10).Truncate(24 * time.Hour),
		Terrain:    TerrainRoad,
		Priority:   PriorityA,
	}
	err := store.CreateRace(r)
	if err == nil {
		t.Error("expected error for past race date")
	}
	if !strings.Contains(err.Error(), "past") {
		t.Errorf("error should mention past, got: %v", err)
	}
}

func TestCreateRace_ValidationError(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	r := &Race{ID: "race_bad"}
	err := store.CreateRace(r)
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestCreateRace_OptionalFields(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	elev := 1200.0
	goal := 14400
	r := &Race{
		ID:          "race_opts",
		Name:        "Mountain Ultra",
		DistanceKm:  50,
		RaceDate:    futureDate(60),
		Terrain:     TerrainTrail,
		ElevationM:  &elev,
		GoalTimeSec: &goal,
		Priority:    PriorityB,
	}
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}

	got, err := store.GetRace("race_opts")
	if err != nil {
		t.Fatalf("GetRace: %v", err)
	}
	if got.ElevationM == nil || *got.ElevationM != 1200.0 {
		t.Errorf("ElevationM = %v, want 1200.0", got.ElevationM)
	}
	if got.GoalTimeSec == nil || *got.GoalTimeSec != 14400 {
		t.Errorf("GoalTimeSec = %v, want 14400", got.GoalTimeSec)
	}
}

func TestUpdateRace_Valid(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	r := testRace("race_upd")
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}

	r.Name = "Updated Marathon"
	r.DistanceKm = 21.1
	if err := store.UpdateRace(r); err != nil {
		t.Fatalf("UpdateRace: %v", err)
	}

	got, err := store.GetRace("race_upd")
	if err != nil {
		t.Fatalf("GetRace: %v", err)
	}
	if got.Name != "Updated Marathon" {
		t.Errorf("Name = %q, want %q", got.Name, "Updated Marathon")
	}
	if got.DistanceKm != 21.1 {
		t.Errorf("DistanceKm = %f, want 21.1", got.DistanceKm)
	}
}

func TestUpdateRace_NotFound(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	r := testRace("nonexistent")
	err := store.UpdateRace(r)
	if err == nil {
		t.Error("expected error for nonexistent race")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestUpdateRace_EmptyID(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	r := testRace("")
	err := store.UpdateRace(r)
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestDeleteRace_Valid(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	r := testRace("race_del")
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}

	if err := store.DeleteRace("race_del"); err != nil {
		t.Fatalf("DeleteRace: %v", err)
	}

	got, err := store.GetRace("race_del")
	if err != nil {
		t.Fatalf("GetRace: %v", err)
	}
	if got != nil {
		t.Error("expected nil after deletion")
	}
}

func TestDeleteRace_NotFound(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	err := store.DeleteRace("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent race")
	}
}

func TestDeleteRace_EmptyID(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	err := store.DeleteRace("")
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestListRaces_Empty(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	races, err := store.ListRaces()
	if err != nil {
		t.Fatalf("ListRaces: %v", err)
	}
	if races != nil {
		t.Errorf("expected nil for empty list, got %d races", len(races))
	}
}

func TestListRaces_OrderByDate(t *testing.T) {
	store, _ := newTestPlanStorage(t)

	dates := []int{90, 30, 60}
	for i, d := range dates {
		r := &Race{
			ID:         fmt.Sprintf("race_%d", i),
			Name:       fmt.Sprintf("Race %d", i),
			DistanceKm: 10,
			RaceDate:   futureDate(d),
			Terrain:    TerrainRoad,
			Priority:   PriorityA,
		}
		if err := store.CreateRace(r); err != nil {
			t.Fatalf("CreateRace %d: %v", i, err)
		}
	}

	races, err := store.ListRaces()
	if err != nil {
		t.Fatalf("ListRaces: %v", err)
	}
	if len(races) != 3 {
		t.Fatalf("expected 3 races, got %d", len(races))
	}
	if races[0].Name != "Race 1" {
		t.Errorf("first race = %q, want %q (30 days out = earliest)", races[0].Name, "Race 1")
	}
	if races[2].Name != "Race 0" {
		t.Errorf("last race = %q, want %q (90 days out = latest)", races[2].Name, "Race 0")
	}
}

func TestGetRace_NotFound(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	got, err := store.GetRace("nonexistent")
	if err != nil {
		t.Fatalf("GetRace: %v", err)
	}
	if got != nil {
		t.Error("expected nil for nonexistent race")
	}
}

func TestSetActiveRace_Basic(t *testing.T) {
	store, _ := newTestPlanStorage(t)

	for i := 0; i < 3; i++ {
		r := &Race{
			ID:         fmt.Sprintf("race_%d", i),
			Name:       fmt.Sprintf("Race %d", i),
			DistanceKm: 10,
			RaceDate:   futureDate(30 + i*30),
			Terrain:    TerrainRoad,
			Priority:   PriorityA,
		}
		if err := store.CreateRace(r); err != nil {
			t.Fatalf("CreateRace: %v", err)
		}
	}

	if err := store.SetActiveRace("race_1"); err != nil {
		t.Fatalf("SetActiveRace: %v", err)
	}

	active, err := store.GetActiveRace()
	if err != nil {
		t.Fatalf("GetActiveRace: %v", err)
	}
	if active == nil {
		t.Fatal("expected active race")
	}
	if active.ID != "race_1" {
		t.Errorf("active race ID = %q, want %q", active.ID, "race_1")
	}

	if err := store.SetActiveRace("race_2"); err != nil {
		t.Fatalf("SetActiveRace second: %v", err)
	}
	active2, err := store.GetActiveRace()
	if err != nil {
		t.Fatalf("GetActiveRace: %v", err)
	}
	if active2.ID != "race_2" {
		t.Errorf("active race after switch = %q, want %q", active2.ID, "race_2")
	}

	r0, _ := store.GetRace("race_0")
	r1, _ := store.GetRace("race_1")
	if r0.IsActive {
		t.Error("race_0 should not be active")
	}
	if r1.IsActive {
		t.Error("race_1 should not be active after switch")
	}
}

func TestSetActiveRace_NotFound(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	err := store.SetActiveRace("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent race")
	}
}

func TestSetActiveRace_EmptyID(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	err := store.SetActiveRace("")
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestGetActiveRace_None(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	got, err := store.GetActiveRace()
	if err != nil {
		t.Fatalf("GetActiveRace: %v", err)
	}
	if got != nil {
		t.Error("expected nil when no active race")
	}
}

func TestSavePlan_AndGetActivePlan(t *testing.T) {
	store, _ := newTestPlanStorage(t)

	r := testRace("race_plan")
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}
	if err := store.SetActiveRace("race_plan"); err != nil {
		t.Fatalf("SetActiveRace: %v", err)
	}

	p := &TrainingPlan{
		ID:          "plan_1",
		RaceID:      "race_plan",
		GeneratedAt: time.Now(),
		LLMBackend:  "local",
		PromptHash:  "abc123",
		Weeks: []Week{
			{
				ID:         "plan_1_w1",
				PlanID:     "plan_1",
				WeekNumber: 1,
				WeekStart:  time.Now().Truncate(24 * time.Hour),
				Sessions: []Session{
					{
						ID:          "plan_1_w1_s1",
						WeekID:      "plan_1_w1",
						DayOfWeek:   1,
						Type:        SessionEasy,
						DurationMin: 45,
						DistanceKm:  8.0,
						HRZone:      2,
						Notes:       "Easy run",
						Status:      StatusPlanned,
					},
					{
						ID:          "plan_1_w1_s2",
						WeekID:      "plan_1_w1",
						DayOfWeek:   3,
						Type:        SessionTempo,
						DurationMin: 40,
						DistanceKm:  7.0,
						HRZone:      3,
						Notes:       "Tempo run",
						Status:      StatusPlanned,
					},
				},
			},
			{
				ID:         "plan_1_w2",
				PlanID:     "plan_1",
				WeekNumber: 2,
				WeekStart:  time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour),
				Sessions: []Session{
					{
						ID:          "plan_1_w2_s1",
						WeekID:      "plan_1_w2",
						DayOfWeek:   1,
						Type:        SessionInterval,
						DurationMin: 50,
						Notes:       "Interval session",
						Status:      StatusPlanned,
					},
				},
			},
		},
	}

	if err := store.SavePlan(p); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	active, err := store.GetActivePlan()
	if err != nil {
		t.Fatalf("GetActivePlan: %v", err)
	}
	if active == nil {
		t.Fatal("expected active plan")
	}
	if active.ID != "plan_1" {
		t.Errorf("plan ID = %q, want %q", active.ID, "plan_1")
	}
	if active.LLMBackend != "local" {
		t.Errorf("LLMBackend = %q, want %q", active.LLMBackend, "local")
	}
}

func TestSavePlan_Nil(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	err := store.SavePlan(nil)
	if err == nil {
		t.Error("expected error for nil plan")
	}
}

func TestSavePlan_EmptyRaceID(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	err := store.SavePlan(&TrainingPlan{ID: "plan_x"})
	if err == nil {
		t.Error("expected error for empty race_id")
	}
}

func TestGetActivePlan_NoActiveRace(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	got, err := store.GetActivePlan()
	if err != nil {
		t.Fatalf("GetActivePlan: %v", err)
	}
	if got != nil {
		t.Error("expected nil when no active race")
	}
}

func TestGetActivePlan_NoPlan(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	r := testRace("race_noplan")
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}
	if err := store.SetActiveRace("race_noplan"); err != nil {
		t.Fatalf("SetActiveRace: %v", err)
	}

	got, err := store.GetActivePlan()
	if err != nil {
		t.Fatalf("GetActivePlan: %v", err)
	}
	if got != nil {
		t.Error("expected nil when no plan exists")
	}
}

func TestGetPlanWeeks_WithSessions(t *testing.T) {
	store, _ := newTestPlanStorage(t)

	r := testRace("race_weeks")
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}

	p := &TrainingPlan{
		ID:          "plan_weeks",
		RaceID:      "race_weeks",
		GeneratedAt: time.Now(),
		LLMBackend:  "local",
		PromptHash:  "hash",
		Weeks: []Week{
			{
				ID:         "plan_weeks_w1",
				PlanID:     "plan_weeks",
				WeekNumber: 1,
				WeekStart:  time.Now().Truncate(24 * time.Hour),
				Sessions: []Session{
					{ID: "pw1_s1", WeekID: "plan_weeks_w1", DayOfWeek: 1, Type: SessionEasy, DurationMin: 45, Notes: "Easy"},
					{ID: "pw1_s2", WeekID: "plan_weeks_w1", DayOfWeek: 3, Type: SessionTempo, DurationMin: 40, Notes: "Tempo"},
					{ID: "pw1_s3", WeekID: "plan_weeks_w1", DayOfWeek: 7, Type: SessionLongRun, DurationMin: 90, Notes: "Long"},
				},
			},
			{
				ID:         "plan_weeks_w2",
				PlanID:     "plan_weeks",
				WeekNumber: 2,
				WeekStart:  time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour),
				Sessions: []Session{
					{ID: "pw2_s1", WeekID: "plan_weeks_w2", DayOfWeek: 2, Type: SessionRest, DurationMin: 0, Notes: "Rest"},
				},
			},
		},
	}
	if err := store.SavePlan(p); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	weeks, err := store.GetPlanWeeks("plan_weeks")
	if err != nil {
		t.Fatalf("GetPlanWeeks: %v", err)
	}
	if len(weeks) != 2 {
		t.Fatalf("expected 2 weeks, got %d", len(weeks))
	}
	if len(weeks[0].Sessions) != 3 {
		t.Errorf("week 1 sessions = %d, want 3", len(weeks[0].Sessions))
	}
	if len(weeks[1].Sessions) != 1 {
		t.Errorf("week 2 sessions = %d, want 1", len(weeks[1].Sessions))
	}
	if weeks[0].Sessions[0].Type != SessionEasy {
		t.Errorf("first session type = %q, want %q", weeks[0].Sessions[0].Type, SessionEasy)
	}
	if weeks[0].Sessions[0].Status != StatusPlanned {
		t.Errorf("first session status = %q, want %q", weeks[0].Sessions[0].Status, StatusPlanned)
	}
}

func TestGetPlanWeeks_EmptyPlanID(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	_, err := store.GetPlanWeeks("")
	if err == nil {
		t.Error("expected error for empty plan ID")
	}
}

func TestGetPlanWeeks_NoPlan(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	weeks, err := store.GetPlanWeeks("nonexistent")
	if err != nil {
		t.Fatalf("GetPlanWeeks: %v", err)
	}
	if weeks != nil {
		t.Errorf("expected nil for nonexistent plan, got %d weeks", len(weeks))
	}
}

func TestUpdateSessionStatus_Complete(t *testing.T) {
	store, _ := newTestPlanStorage(t)

	r := testRace("race_status")
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}

	p := &TrainingPlan{
		ID:          "plan_status",
		RaceID:      "race_status",
		GeneratedAt: time.Now(),
		LLMBackend:  "local",
		PromptHash:  "h",
		Weeks: []Week{{
			ID: "ps_w1", PlanID: "plan_status", WeekNumber: 1,
			WeekStart: time.Now().Truncate(24 * time.Hour),
			Sessions: []Session{{
				ID: "ps_w1_s1", WeekID: "ps_w1", DayOfWeek: 1,
				Type: SessionEasy, DurationMin: 45, Notes: "test",
			}},
		}},
	}
	if err := store.SavePlan(p); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	dur := 50
	dist := 9.5
	actual := ActualMetrics{DurationMin: &dur, DistanceKm: &dist}
	if err := store.UpdateSessionStatus("ps_w1_s1", StatusCompleted, actual); err != nil {
		t.Fatalf("UpdateSessionStatus: %v", err)
	}

	weeks, err := store.GetPlanWeeks("plan_status")
	if err != nil {
		t.Fatalf("GetPlanWeeks: %v", err)
	}
	sess := weeks[0].Sessions[0]
	if sess.Status != StatusCompleted {
		t.Errorf("status = %q, want %q", sess.Status, StatusCompleted)
	}
	if sess.ActualDurationMin == nil || *sess.ActualDurationMin != 50 {
		t.Errorf("actual_duration_min = %v, want 50", sess.ActualDurationMin)
	}
	if sess.ActualDistanceKm == nil || *sess.ActualDistanceKm != 9.5 {
		t.Errorf("actual_distance_km = %v, want 9.5", sess.ActualDistanceKm)
	}
	if sess.CompletedAt == nil {
		t.Error("completed_at should be set for completed status")
	}
}

func TestUpdateSessionStatus_Skip(t *testing.T) {
	store, _ := newTestPlanStorage(t)

	r := testRace("race_skip")
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}

	p := &TrainingPlan{
		ID:          "plan_skip",
		RaceID:      "race_skip",
		GeneratedAt: time.Now(),
		LLMBackend:  "local",
		PromptHash:  "h",
		Weeks: []Week{{
			ID: "psk_w1", PlanID: "plan_skip", WeekNumber: 1,
			WeekStart: time.Now().Truncate(24 * time.Hour),
			Sessions: []Session{{
				ID: "psk_w1_s1", WeekID: "psk_w1", DayOfWeek: 1,
				Type: SessionEasy, DurationMin: 45, Notes: "test",
			}},
		}},
	}
	if err := store.SavePlan(p); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	if err := store.UpdateSessionStatus("psk_w1_s1", StatusSkipped, ActualMetrics{}); err != nil {
		t.Fatalf("UpdateSessionStatus: %v", err)
	}

	weeks, _ := store.GetPlanWeeks("plan_skip")
	sess := weeks[0].Sessions[0]
	if sess.Status != StatusSkipped {
		t.Errorf("status = %q, want %q", sess.Status, StatusSkipped)
	}
	if sess.CompletedAt != nil {
		t.Error("completed_at should be nil for skipped status")
	}
}

func TestUpdateSessionStatus_NotFound(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	err := store.UpdateSessionStatus("nonexistent", StatusCompleted, ActualMetrics{})
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestUpdateSessionStatus_EmptyID(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	err := store.UpdateSessionStatus("", StatusCompleted, ActualMetrics{})
	if err == nil {
		t.Error("expected error for empty session ID")
	}
}

func TestUpdateSessionStatus_InvalidStatus(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	err := store.UpdateSessionStatus("some_id", "invalid", ActualMetrics{})
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestGetLatestPlanForRace(t *testing.T) {
	store, _ := newTestPlanStorage(t)

	r := testRace("race_latest")
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}

	p1 := &TrainingPlan{
		ID: "plan_old", RaceID: "race_latest",
		GeneratedAt: time.Now().Add(-time.Hour), LLMBackend: "local", PromptHash: "h1",
	}
	p2 := &TrainingPlan{
		ID: "plan_new", RaceID: "race_latest",
		GeneratedAt: time.Now(), LLMBackend: "local", PromptHash: "h2",
	}

	if err := store.SavePlan(p1); err != nil {
		t.Fatalf("SavePlan 1: %v", err)
	}
	if err := store.SavePlan(p2); err != nil {
		t.Fatalf("SavePlan 2: %v", err)
	}

	latest, err := store.GetLatestPlanForRace("race_latest")
	if err != nil {
		t.Fatalf("GetLatestPlanForRace: %v", err)
	}
	if latest == nil {
		t.Fatal("expected plan, got nil")
	}
	if latest.ID != "plan_new" {
		t.Errorf("latest plan = %q, want %q", latest.ID, "plan_new")
	}
}

func TestGetLatestPlanForRace_NotFound(t *testing.T) {
	store, _ := newTestPlanStorage(t)
	got, err := store.GetLatestPlanForRace("nonexistent")
	if err != nil {
		t.Fatalf("GetLatestPlanForRace: %v", err)
	}
	if got != nil {
		t.Error("expected nil for nonexistent race")
	}
}

func TestDeleteRace_CascadeDeletesPlan(t *testing.T) {
	store, _ := newTestPlanStorage(t)

	r := testRace("race_cascade")
	if err := store.CreateRace(r); err != nil {
		t.Fatalf("CreateRace: %v", err)
	}

	p := &TrainingPlan{
		ID: "plan_cascade", RaceID: "race_cascade",
		GeneratedAt: time.Now(), LLMBackend: "local", PromptHash: "h",
		Weeks: []Week{{
			ID: "pc_w1", PlanID: "plan_cascade", WeekNumber: 1,
			WeekStart: time.Now().Truncate(24 * time.Hour),
			Sessions: []Session{{
				ID: "pc_w1_s1", WeekID: "pc_w1", DayOfWeek: 1,
				Type: SessionEasy, DurationMin: 30, Notes: "test",
			}},
		}},
	}
	if err := store.SavePlan(p); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	weeks, _ := store.GetPlanWeeks("plan_cascade")
	if len(weeks) == 0 {
		t.Fatal("expected weeks before delete")
	}

	if err := store.DeleteRace("race_cascade"); err != nil {
		t.Fatalf("DeleteRace: %v", err)
	}

	latest, _ := store.GetLatestPlanForRace("race_cascade")
	if latest != nil {
		t.Error("plan should be cascade-deleted with race")
	}

	weeks2, _ := store.GetPlanWeeks("plan_cascade")
	if len(weeks2) != 0 {
		t.Errorf("weeks should be cascade-deleted, got %d", len(weeks2))
	}
}
