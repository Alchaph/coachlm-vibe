package context

import (
	"fmt"
	"strings"

	"coachlm/internal/storage"
)

// FormatStatsBlock formats athlete statistics (recent, YTD, all-time) for LLM context.
// Returns empty string if stats is nil or all counts are zero.
func FormatStatsBlock(stats *storage.AthleteStats) string {
	if stats == nil {
		return ""
	}
	if stats.RecentRunCount == 0 && stats.YTDRunCount == 0 && stats.AllRunCount == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "## Training Load Statistics")

	if stats.RecentRunCount > 0 {
		lines = append(lines, fmt.Sprintf("Recent (4 weeks): %d runs, %.1f km, %s, %.0f m elevation",
			stats.RecentRunCount,
			stats.RecentRunDistance/1000.0,
			formatStatsDuration(stats.RecentRunMovingTime),
			stats.RecentRunElevation,
		))
	}

	if stats.YTDRunCount > 0 {
		lines = append(lines, fmt.Sprintf("Year to date: %d runs, %.1f km, %s, %.0f m elevation",
			stats.YTDRunCount,
			stats.YTDRunDistance/1000.0,
			formatStatsDuration(stats.YTDRunMovingTime),
			stats.YTDRunElevation,
		))
	}

	if stats.AllRunCount > 0 {
		lines = append(lines, fmt.Sprintf("All time: %d runs, %.1f km, %s, %.0f m elevation",
			stats.AllRunCount,
			stats.AllRunDistance/1000.0,
			formatStatsDuration(stats.AllRunMovingTime),
			stats.AllRunElevation,
		))
	}

	return strings.Join(lines, "\n")
}

// formatStatsDuration formats total seconds into a human-readable duration like "42h15m".
func formatStatsDuration(secs int) string {
	if secs <= 0 {
		return "0h"
	}
	h := secs / 3600
	m := (secs % 3600) / 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%02dm", h, m)
}
