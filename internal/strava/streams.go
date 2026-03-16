package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type StreamSet struct {
	HeartRate []int     // bpm per second; nil if unavailable
	Pace      []float64 // sec/km per second (from velocity_smooth m/s); nil if unavailable
	Cadence   []int     // spm per second; nil if unavailable
}

type stravaStream struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// FetchStreams retrieves HR, velocity_smooth, and cadence streams for an activity.
// Missing streams (common for non-running types) result in nil slices, not errors.
func (wh *WebhookHandler) FetchStreams(ctx context.Context, accessToken string, activityID int64) (*StreamSet, error) {
	url := fmt.Sprintf("%s/activities/%d/streams?keys=heartrate,velocity_smooth,cadence&key_type=time", wh.stravaAPIBase, activityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create streams request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := wh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch streams: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava streams API error: HTTP %d", resp.StatusCode)
	}

	var rawStreams []stravaStream
	if err := json.NewDecoder(resp.Body).Decode(&rawStreams); err != nil {
		return nil, fmt.Errorf("decode streams response: %w", err)
	}

	return parseStreams(rawStreams)
}

func parseStreams(rawStreams []stravaStream) (*StreamSet, error) {
	ss := &StreamSet{}

	for _, s := range rawStreams {
		switch s.Type {
		case "heartrate":
			var data []int
			if err := json.Unmarshal(s.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal heartrate stream: %w", err)
			}
			ss.HeartRate = data

		case "velocity_smooth":
			var velocities []float64
			if err := json.Unmarshal(s.Data, &velocities); err != nil {
				return nil, fmt.Errorf("unmarshal velocity_smooth stream: %w", err)
			}
			ss.Pace = velocityToPace(velocities)

		case "cadence":
			var data []int
			if err := json.Unmarshal(s.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal cadence stream: %w", err)
			}
			ss.Cadence = data
		}
	}

	return ss, nil
}

// velocityToPace converts m/s → sec/km. Zero/negative velocity → 0.
func velocityToPace(velocities []float64) []float64 {
	paces := make([]float64, len(velocities))
	for i, v := range velocities {
		if v > 0 {
			paces[i] = 1000.0 / v
		}
	}
	return paces
}
