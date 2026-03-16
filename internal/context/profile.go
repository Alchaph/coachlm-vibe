package context

import (
	"fmt"
	"strings"

	"coachlm/internal/storage"
)

// FormatPace converts seconds-per-km to "M:SS/km" format.
// For example, 270 → "4:30/km", 300 → "5:00/km", 195 → "3:15/km".
func FormatPace(secs int) string {
	minutes := secs / 60
	remainder := secs % 60
	return fmt.Sprintf("%d:%02d/km", minutes, remainder)
}

// FormatProfileBlock takes an AthleteProfile and returns a structured text
// block suitable for inclusion in LLM context. Fields with zero/empty values
// are omitted. Output is deterministic: same input always produces same output.
func FormatProfileBlock(profile *storage.AthleteProfile) string {
	if profile == nil {
		return "No profile configured."
	}

	var lines []string

	if profile.Age > 0 {
		lines = append(lines, fmt.Sprintf("Age: %d", profile.Age))
	}
	if profile.MaxHR > 0 {
		lines = append(lines, fmt.Sprintf("Max Heart Rate: %d bpm", profile.MaxHR))
	}
	if profile.ThresholdPaceSecs > 0 {
		lines = append(lines, fmt.Sprintf("Threshold Pace: %s", FormatPace(profile.ThresholdPaceSecs)))
	}
	if profile.WeeklyMileageTarget > 0 {
		lines = append(lines, fmt.Sprintf("Weekly Mileage Target: %.1f km", profile.WeeklyMileageTarget))
	}
	if profile.RaceGoals != "" {
		lines = append(lines, fmt.Sprintf("Race Goals: %s", profile.RaceGoals))
	}
	if profile.InjuryHistory != "" {
		lines = append(lines, fmt.Sprintf("Injury History: %s", profile.InjuryHistory))
	}

	if len(lines) == 0 {
		return "No profile configured."
	}

	return strings.Join(lines, "\n")
}
