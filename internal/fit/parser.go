package fit

import (
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/tormoder/fit"
)

type ParsedActivity struct {
	Name         string
	ActivityType string // "Run", "Ride", "Swim", etc.
	StartDate    time.Time
	Distance     float64 // meters
	DurationSecs int
	AvgPaceSecs  int // seconds per km, 0 if not calculable
	AvgHR        int
	MaxHR        int
	AvgCadence   float64

	HeartRate []int     // per-second HR, nil if unavailable
	Pace      []int     // per-second pace in sec/km, nil if unavailable
	Cadence   []float64 // per-second cadence in spm, nil if unavailable
}

func ParseFITFile(filePath string) (*ParsedActivity, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path is empty")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open FIT file: %w", err)
	}
	defer f.Close()

	fitFile, err := fit.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode FIT file: %w", err)
	}

	return ParseFITData(fitFile)
}

// ParseFITData is exported for testing with synthetic FIT data structures.
func ParseFITData(fitFile *fit.File) (*ParsedActivity, error) {
	if fitFile == nil {
		return nil, fmt.Errorf("FIT file data is nil")
	}

	activity, err := fitFile.Activity()
	if err != nil {
		return nil, fmt.Errorf("not an activity file: %w", err)
	}

	if len(activity.Sessions) == 0 {
		return nil, fmt.Errorf("no sessions found in FIT file")
	}

	session := activity.Sessions[0]
	parsed := &ParsedActivity{
		Name:         buildActivityName(session),
		ActivityType: mapSport(session.Sport),
		StartDate:    session.StartTime,
	}

	dist := session.GetTotalDistanceScaled()
	if !math.IsNaN(dist) {
		parsed.Distance = dist
	}

	dur := session.GetTotalTimerTimeScaled()
	if !math.IsNaN(dur) {
		parsed.DurationSecs = int(dur)
	}

	if parsed.Distance > 0 && parsed.DurationSecs > 0 {
		parsed.AvgPaceSecs = int(float64(parsed.DurationSecs) / (parsed.Distance / 1000.0))
	}

	if session.AvgHeartRate != 0xFF {
		parsed.AvgHR = int(session.AvgHeartRate)
	}
	if session.MaxHeartRate != 0xFF {
		parsed.MaxHR = int(session.MaxHeartRate)
	}

	if session.AvgCadence != 0xFF {
		parsed.AvgCadence = float64(session.AvgCadence)
		// FIT running cadence is steps-per-leg; double for steps-per-minute
		if session.Sport == fit.SportRunning {
			parsed.AvgCadence *= 2
		}
	}

	extractStreams(parsed, activity.Records, session.Sport)

	return parsed, nil
}

func extractStreams(parsed *ParsedActivity, records []*fit.RecordMsg, sport fit.Sport) {
	if len(records) == 0 {
		return
	}

	var heartRates []int
	var paces []int
	var cadences []float64

	hasHR := false
	hasPace := false
	hasCadence := false

	for _, rec := range records {
		hr := int(rec.HeartRate)
		if rec.HeartRate != 0xFF {
			hasHR = true
		} else {
			hr = 0
		}
		heartRates = append(heartRates, hr)

		speed := rec.GetEnhancedSpeedScaled()
		if math.IsNaN(speed) {
			speed = rec.GetSpeedScaled()
		}
		pace := 0
		if !math.IsNaN(speed) && speed > 0 {
			hasPace = true
			pace = int(1000.0 / speed) // m/s → sec/km
		}
		paces = append(paces, pace)

		cad := float64(0)
		if rec.Cadence != 0xFF {
			hasCadence = true
			cad = float64(rec.Cadence)
			// FIT running cadence is steps-per-leg; double for steps-per-minute
			if sport == fit.SportRunning {
				cad *= 2
			}
		}
		cadences = append(cadences, cad)
	}

	if hasHR {
		parsed.HeartRate = heartRates
	}
	if hasPace {
		parsed.Pace = paces
	}
	if hasCadence {
		parsed.Cadence = cadences
	}
}

func mapSport(sport fit.Sport) string {
	switch sport {
	case fit.SportRunning:
		return "Run"
	case fit.SportCycling, fit.SportEBiking:
		return "Ride"
	case fit.SportSwimming:
		return "Swim"
	case fit.SportWalking:
		return "Walk"
	case fit.SportHiking:
		return "Hike"
	case fit.SportRowing:
		return "Rowing"
	case fit.SportTraining:
		return "Workout"
	case fit.SportCrossCountrySkiing:
		return "NordicSki"
	case fit.SportAlpineSkiing:
		return "AlpineSki"
	case fit.SportSnowboarding:
		return "Snowboard"
	default:
		return sport.String()
	}
}

func buildActivityName(session *fit.SessionMsg) string {
	sportName := mapSport(session.Sport)
	hour := session.StartTime.Hour()

	var timeOfDay string
	switch {
	case hour < 6:
		timeOfDay = "Early Morning"
	case hour < 12:
		timeOfDay = "Morning"
	case hour < 17:
		timeOfDay = "Afternoon"
	case hour < 21:
		timeOfDay = "Evening"
	default:
		timeOfDay = "Night"
	}

	return fmt.Sprintf("%s %s", timeOfDay, sportName)
}

func DeduplicationHash(a *ParsedActivity) string {
	if a == nil {
		return ""
	}
	data := fmt.Sprintf("%d|%d|%.2f",
		a.StartDate.Unix(),
		a.DurationSecs,
		a.Distance,
	)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
