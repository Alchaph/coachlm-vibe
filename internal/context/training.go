package context

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"coachlm/internal/storage"
)

// FormatDuration formats seconds into HH:MM:SS or MM:SS.
func FormatDuration(secs int) string {
	if secs < 0 {
		secs = 0
	}
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func weekBucket(activityTime, now time.Time) int {
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	actDate := time.Date(activityTime.Year(), activityTime.Month(), activityTime.Day(), 0, 0, 0, 0, now.Location())

	days := int(nowDate.Sub(actDate).Hours() / 24)
	if days < 0 {
		return -1
	}
	week := days / 7
	if week > 3 {
		return -1
	}
	return week
}

func dayKey(t time.Time) string {
	return t.Weekday().String()[:3]
}

// FormatTrainingSummary generates a rolling training summary with graduated
// compression levels for the last 4 weeks of activities.
//
// Compression levels (older data compressed more):
//   - Week 0 (most recent, 0-7 days): Per-run detail
//   - Week 1 (8-14 days ago): Daily aggregates
//   - Week 2 (15-21 days ago): Key sessions only
//   - Week 3 (22-28 days ago): Weekly totals only
func FormatTrainingSummary(activities []storage.Activity, now time.Time) string {
	weeks := [4][]storage.Activity{}
	for i := range activities {
		bucket := weekBucket(activities[i].StartDate, now)
		if bucket >= 0 && bucket <= 3 {
			weeks[bucket] = append(weeks[bucket], activities[i])
		}
	}

	total := 0
	for _, w := range weeks {
		total += len(w)
	}
	if total == 0 {
		return "No recent training data."
	}

	for i := range weeks {
		sort.Slice(weeks[i], func(a, b int) bool {
			return weeks[i][a].StartDate.Before(weeks[i][b].StartDate)
		})
	}

	var sb strings.Builder
	sb.WriteString("## Training Summary (Last 4 Weeks)")

	if len(weeks[0]) > 0 {
		sb.WriteString("\n\n### This Week (detailed)")
		for _, a := range weeks[0] {
			sb.WriteString("\n")
			sb.WriteString(formatDetailedActivity(a))
		}
	}

	if len(weeks[1]) > 0 {
		sb.WriteString("\n\n### Last Week (daily summary)")
		sb.WriteString(formatDailySummary(weeks[1]))
	}

	if len(weeks[2]) > 0 {
		sb.WriteString("\n\n### 2 Weeks Ago (key sessions)")
		sb.WriteString(formatKeySessions(weeks[2]))
	}

	if len(weeks[3]) > 0 {
		sb.WriteString("\n\n### 3 Weeks Ago (totals)")
		sb.WriteString(formatWeeklyTotals(weeks[3]))
	}

	return sb.String()
}

func formatDetailedActivity(a storage.Activity) string {
	day := dayKey(a.StartDate)
	distKm := a.Distance / 1000.0

	parts := []string{
		fmt.Sprintf("%.1f km", distKm),
		FormatDuration(a.DurationSecs),
	}

	if a.AvgPaceSecs > 0 {
		parts = append(parts, FormatPace(a.AvgPaceSecs))
	}
	if a.AvgHR > 0 {
		parts = append(parts, fmt.Sprintf("HR %d", a.AvgHR))
	}

	name := a.Name
	if a.ActivityType != "" && a.ActivityType != "Run" {
		name = fmt.Sprintf("%s [%s]", a.Name, a.ActivityType)
	}

	return fmt.Sprintf("- %s: %s — %s", day, name, strings.Join(parts, ", "))
}

type dayAggregate struct {
	date         time.Time
	dayName      string
	runCount     int
	totalDist    float64
	totalDurSecs int
}

func formatDailySummary(activities []storage.Activity) string {
	dayMap := make(map[string]*dayAggregate)
	var dayOrder []string

	for _, a := range activities {
		dateKey := a.StartDate.Format("2006-01-02")
		agg, exists := dayMap[dateKey]
		if !exists {
			agg = &dayAggregate{
				date:    a.StartDate,
				dayName: dayKey(a.StartDate),
			}
			dayMap[dateKey] = agg
			dayOrder = append(dayOrder, dateKey)
		}
		agg.runCount++
		agg.totalDist += a.Distance
		agg.totalDurSecs += a.DurationSecs
	}

	sort.Strings(dayOrder)

	var sb strings.Builder
	for _, key := range dayOrder {
		agg := dayMap[key]
		distKm := agg.totalDist / 1000.0
		runLabel := "run"
		if agg.runCount != 1 {
			runLabel = "runs"
		}
		sb.WriteString(fmt.Sprintf("\n- %s: %d %s, %.1f km, %s",
			agg.dayName, agg.runCount, runLabel, distKm, FormatDuration(agg.totalDurSecs)))
	}
	return sb.String()
}

func formatKeySessions(activities []storage.Activity) string {
	var longest *storage.Activity
	var fastest *storage.Activity
	dayMileage := make(map[string]float64)
	dayDates := make(map[string]time.Time)

	for i := range activities {
		a := &activities[i]

		if longest == nil || a.Distance > longest.Distance {
			longest = a
		}

		if a.AvgPaceSecs > 0 && (fastest == nil || a.AvgPaceSecs < fastest.AvgPaceSecs) {
			fastest = a
		}

		dateKey := a.StartDate.Format("2006-01-02")
		dayMileage[dateKey] += a.Distance
		if _, exists := dayDates[dateKey]; !exists {
			dayDates[dateKey] = a.StartDate
		}
	}

	var sb strings.Builder

	if longest != nil {
		sb.WriteString(fmt.Sprintf("\n- Longest: %s — %.1f km", longest.Name, longest.Distance/1000.0))
	}
	if fastest != nil {
		sb.WriteString(fmt.Sprintf("\n- Fastest: %s — %s", fastest.Name, FormatPace(fastest.AvgPaceSecs)))
	}

	var highestDay string
	var highestMileage float64
	for dateKey, mileage := range dayMileage {
		if mileage > highestMileage {
			highestMileage = mileage
			highestDay = dateKey
		}
	}
	if highestDay != "" {
		dayName := dayKey(dayDates[highestDay])
		sb.WriteString(fmt.Sprintf("\n- Highest mileage day: %s — %.1f km", dayName, highestMileage/1000.0))
	}

	var totalDist float64
	for _, a := range activities {
		totalDist += a.Distance
	}
	sb.WriteString(fmt.Sprintf("\nTotal: %.1f km, %d runs", totalDist/1000.0, len(activities)))

	return sb.String()
}

func formatWeeklyTotals(activities []storage.Activity) string {
	var totalDist float64
	var totalDurSecs int
	var paceSum int
	var paceCount int

	for _, a := range activities {
		totalDist += a.Distance
		totalDurSecs += a.DurationSecs
		if a.AvgPaceSecs > 0 {
			paceSum += a.AvgPaceSecs
			paceCount++
		}
	}

	distKm := totalDist / 1000.0
	parts := []string{
		fmt.Sprintf("%.1f km", distKm),
		fmt.Sprintf("%d runs", len(activities)),
		FormatDuration(totalDurSecs),
	}

	if paceCount > 0 {
		avgPace := paceSum / paceCount
		parts = append(parts, fmt.Sprintf("avg %s", FormatPace(avgPace)))
	}

	return fmt.Sprintf("\nTotal: %s", strings.Join(parts, ", "))
}
