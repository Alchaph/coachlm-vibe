package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ZoneRange represents a single heart rate zone with min and max BPM values.
type ZoneRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// AthleteZonesResponse is the top-level response from GET /athlete/zones.
type AthleteZonesResponse struct {
	HeartRate HeartRateZones `json:"heart_rate"`
}

// HeartRateZones contains the athlete's heart rate zone configuration.
type HeartRateZones struct {
	CustomZones bool        `json:"custom_zones"`
	Zones       []ZoneRange `json:"zones"`
}

// FetchAthleteZones retrieves the authenticated athlete's heart rate zones
// from the Strava API (GET /athlete/zones). Returns the HR zones slice and
// whether they are custom-configured by the athlete.
func FetchAthleteZones(ctx context.Context, httpClient *http.Client, apiBase, accessToken string) (*HeartRateZones, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	url := fmt.Sprintf("%s/athlete/zones", apiBase)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create zones request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch zones: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava zones API error: HTTP %d", resp.StatusCode)
	}

	var zonesResp AthleteZonesResponse
	if err := json.NewDecoder(resp.Body).Decode(&zonesResp); err != nil {
		return nil, fmt.Errorf("decode zones response: %w", err)
	}

	return &zonesResp.HeartRate, nil
}
