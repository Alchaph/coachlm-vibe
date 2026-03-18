package context

import (
	"encoding/json"
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
	if profile.ExperienceLevel != "" {
		lines = append(lines, fmt.Sprintf("Experience Level: %s", profile.ExperienceLevel))
	}
	if profile.TrainingDaysPerWeek > 0 {
		lines = append(lines, fmt.Sprintf("Training Days Per Week: %d", profile.TrainingDaysPerWeek))
	}
	if profile.RestingHR > 0 {
		lines = append(lines, fmt.Sprintf("Resting Heart Rate: %d bpm", profile.RestingHR))
	}
	if profile.PreferredTerrain != "" {
		lines = append(lines, fmt.Sprintf("Preferred Terrain: %s", profile.PreferredTerrain))
	}
	if profile.HeartRateZones != "" {
		lines = append(lines, formatHeartRateZones(profile.HeartRateZones)...)
	}

	if len(lines) == 0 {
		return "No profile configured."
	}

	return strings.Join(lines, "\n")
}

var zoneLabels = []string{"Recovery", "Endurance", "Tempo", "Threshold", "VO2 Max"}

func formatHeartRateZones(zonesJSON string) []string {
	var zones []struct {
		Min int `json:"min"`
		Max int `json:"max"`
	}
	if err := json.Unmarshal([]byte(zonesJSON), &zones); err != nil || len(zones) == 0 {
		return nil
	}

	lines := []string{"Heart Rate Zones:"}
	for i, z := range zones {
		label := "Zone"
		if i < len(zoneLabels) {
			label = zoneLabels[i]
		}
		if z.Max == -1 || z.Max == 0 {
			lines = append(lines, fmt.Sprintf("- Zone %d: %d+ bpm (%s)", i+1, z.Min, label))
		} else {
			lines = append(lines, fmt.Sprintf("- Zone %d: %d-%d bpm (%s)", i+1, z.Min, z.Max, label))
		}
	}
	return lines
}
