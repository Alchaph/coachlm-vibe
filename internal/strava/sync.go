package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"coachlm/internal/storage"
)

const defaultPerPage = 30

// StravaActivitySummary represents the summary fields returned by the
// GET /athlete/activities list endpoint. This is a subset of the detail
// endpoint response; the list endpoint omits some fields but includes
// enough data for basic activity storage.
type StravaActivitySummary struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Type             string  `json:"type"`
	StartDate        string  `json:"start_date"`
	Distance         float64 `json:"distance"`
	MovingTime       int     `json:"moving_time"`
	AverageSpeed     float64 `json:"average_speed"`
	AverageHeartrate float64 `json:"average_heartrate"`
	MaxHeartrate     float64 `json:"max_heartrate"`
	AverageCadence   float64 `json:"average_cadence"`
	GearID           string  `json:"gear_id"`
}

// toActivity converts a StravaActivitySummary to a storage.Activity.
func (s *StravaActivitySummary) toActivity() *storage.Activity {
	startDate, _ := time.Parse(time.RFC3339, s.StartDate)

	var avgPaceSecs int
	if s.AverageSpeed > 0 {
		avgPaceSecs = int(1000.0 / s.AverageSpeed)
	}

	return &storage.Activity{
		StravaID:     s.ID,
		Name:         s.Name,
		ActivityType: s.Type,
		StartDate:    startDate,
		Distance:     s.Distance,
		DurationSecs: s.MovingTime,
		AvgPaceSecs:  avgPaceSecs,
		AvgHR:        int(s.AverageHeartrate),
		MaxHR:        int(s.MaxHeartrate),
		AvgCadence:   s.AverageCadence,
		Source:       "strava",
	}
}

// FetchAthleteActivities pages through the Strava athlete activities list
// endpoint and returns all activities as storage-ready structs along with
// unique gear IDs found across all activities. It stops when Strava returns
// an empty page.
func FetchAthleteActivities(ctx context.Context, httpClient *http.Client, apiBase, accessToken string) ([]*storage.Activity, []string, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	var all []*storage.Activity
	gearSet := make(map[string]struct{})
	page := 1

	for {
		url := fmt.Sprintf("%s/athlete/activities?page=%d&per_page=%d", apiBase, page, defaultPerPage)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("create request page %d: %w", page, err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("fetch page %d: %w", page, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, nil, fmt.Errorf("strava API error on page %d: HTTP %d", page, resp.StatusCode)
		}

		var summaries []StravaActivitySummary
		if err := json.NewDecoder(resp.Body).Decode(&summaries); err != nil {
			resp.Body.Close()
			return nil, nil, fmt.Errorf("decode page %d: %w", page, err)
		}
		resp.Body.Close()

		if len(summaries) == 0 {
			break
		}

		for i := range summaries {
			all = append(all, summaries[i].toActivity())
			if summaries[i].GearID != "" {
				gearSet[summaries[i].GearID] = struct{}{}
			}
		}

		if len(summaries) < defaultPerPage {
			break
		}

		page++
	}

	gearIDs := make([]string, 0, len(gearSet))
	for id := range gearSet {
		gearIDs = append(gearIDs, id)
	}

	return all, gearIDs, nil
}
