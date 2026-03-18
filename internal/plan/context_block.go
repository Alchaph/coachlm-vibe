package plan

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// maxPlanBlockTokens is the hard cap for the plan context block (approx 400 tokens).
const maxPlanBlockTokens = 400

// estimateTokens returns an estimated token count: 4 characters ~ 1 token.
func estimateTokens(text string) int {
	return (len(text) + 3) / 4
}

// PlanBlock formats the active training plan as a context block for the LLM prompt.
// Returns empty string if no active plan exists.
func PlanBlock(store *Storage, now time.Time) string {
	if store == nil {
		return ""
	}
	race, err := store.GetActiveRace()
	if err != nil || race == nil {
		return ""
	}

	plan, err := store.GetActivePlan()
	if err != nil || plan == nil {
		return ""
	}

	weeks, err := store.GetPlanWeeks(plan.ID)
	if err != nil || len(weeks) == 0 {
		return ""
	}

	weeksRemaining := weeksUntil(now, race.RaceDate)
	currentWeekNum := findCurrentWeek(weeks, now)

	var sb strings.Builder
	sb.WriteString("## Active Training Plan\n")
	sb.WriteString(fmt.Sprintf("Race: %s (%.1f km %s) on %s\n",
		race.Name, race.DistanceKm, race.Terrain, race.RaceDate.Format("Jan 02, 2006")))
	sb.WriteString(fmt.Sprintf("Weeks remaining: %d\n\n", weeksRemaining))

	// Current week
	if cw := findWeek(weeks, currentWeekNum); cw != nil {
		sb.WriteString(fmt.Sprintf("### This Week (Week %d)\n", currentWeekNum))
		writeWeekSessions(&sb, cw)
	}

	// Check token budget before adding next week.
	if estimateTokens(sb.String()) < maxPlanBlockTokens {
		nextWeekNum := currentWeekNum + 1
		if nw := findWeek(weeks, nextWeekNum); nw != nil {
			nextSection := formatNextWeek(nw, nextWeekNum)
			if estimateTokens(sb.String()+nextSection) <= maxPlanBlockTokens {
				sb.WriteString(nextSection)
			}
		}
	}

	result := sb.String()
	// Hard cap enforcement.
	if estimateTokens(result) > maxPlanBlockTokens {
		maxChars := maxPlanBlockTokens * 4
		if maxChars < len(result) {
			result = result[:maxChars]
		}
	}

	return result
}

// findCurrentWeek determines which week number we're currently in.
func findCurrentWeek(weeks []Week, now time.Time) int {
	today := now.Truncate(24 * time.Hour)
	for _, w := range weeks {
		weekEnd := w.WeekStart.AddDate(0, 0, 7)
		if !today.Before(w.WeekStart) && today.Before(weekEnd) {
			return w.WeekNumber
		}
	}
	// Default: if we're past all weeks, return the last week.
	if len(weeks) > 0 {
		last := weeks[len(weeks)-1]
		if today.After(last.WeekStart) || today.Equal(last.WeekStart) {
			return last.WeekNumber
		}
		// Before first week — return week 1.
		return weeks[0].WeekNumber
	}
	return 1
}

// findWeek returns the week with the given number, or nil.
func findWeek(weeks []Week, num int) *Week {
	for i := range weeks {
		if weeks[i].WeekNumber == num {
			return &weeks[i]
		}
	}
	return nil
}

var dayNames = [8]string{"", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

// writeWeekSessions appends session info for a week.
func writeWeekSessions(sb *strings.Builder, w *Week) {
	for _, sess := range w.Sessions {
		dayName := "?"
		if sess.DayOfWeek >= 1 && sess.DayOfWeek <= 7 {
			dayName = dayNames[sess.DayOfWeek]
		}
		status := ""
		if sess.Status != StatusPlanned {
			status = fmt.Sprintf(" [%s]", sess.Status)
		}
		if sess.Type == SessionRest {
			sb.WriteString(fmt.Sprintf("- %s: Rest%s\n", dayName, status))
		} else {
			line := fmt.Sprintf("- %s: %s %dmin", dayName, sess.Type, sess.DurationMin)
			if sess.DistanceKm > 0 {
				line += fmt.Sprintf(" / %.1fkm", sess.DistanceKm)
			}
			if sess.HRZone > 0 {
				line += fmt.Sprintf(" Z%d", sess.HRZone)
			}
			line += status
			sb.WriteString(line + "\n")
		}
	}
	sb.WriteString("\n")
}

// formatNextWeek formats the next week section.
func formatNextWeek(w *Week, weekNum int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### Next Week (Week %d)\n", weekNum))
	writeWeekSessions(&sb, w)
	return sb.String()
}

// WeeksToRaceDisplay returns a user-friendly string for weeks remaining.
func WeeksToRaceDisplay(now, raceDate time.Time) string {
	days := int(math.Ceil(raceDate.Sub(now).Hours() / 24))
	if days <= 0 {
		return "Race day!"
	}
	weeks := (days + 6) / 7
	if weeks == 1 {
		return fmt.Sprintf("%d day(s) to race", days)
	}
	return fmt.Sprintf("~%d weeks to race", weeks)
}
