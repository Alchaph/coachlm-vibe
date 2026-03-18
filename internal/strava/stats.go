package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AthleteStats struct {
	RecentRunTotals ActivityTotals `json:"recent_run_totals"`
	YTDRunTotals    ActivityTotals `json:"ytd_run_totals"`
	AllRunTotals    ActivityTotals `json:"all_run_totals"`
}

type ActivityTotals struct {
	Count         int     `json:"count"`
	Distance      float64 `json:"distance"`
	MovingTime    int     `json:"moving_time"`
	ElevationGain float64 `json:"elevation_gain"`
}

type GearDetail struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Distance    float64 `json:"distance"`
	BrandName   string  `json:"brand_name"`
	ModelName   string  `json:"model_name"`
	Description string  `json:"description"`
	Primary     bool    `json:"primary"`
	Retired     bool    `json:"retired"`
}

func FetchAuthenticatedAthleteID(ctx context.Context, httpClient *http.Client, apiBase, accessToken string) (int64, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	url := fmt.Sprintf("%s/athlete", apiBase)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("create athlete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetch athlete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("strava athlete API error: HTTP %d", resp.StatusCode)
	}

	var athlete struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&athlete); err != nil {
		return 0, fmt.Errorf("decode athlete response: %w", err)
	}

	return athlete.ID, nil
}

func FetchAthleteStats(ctx context.Context, httpClient *http.Client, apiBase, accessToken string, athleteID int64) (*AthleteStats, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	url := fmt.Sprintf("%s/athletes/%d/stats", apiBase, athleteID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create stats request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch athlete stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava stats API error: HTTP %d", resp.StatusCode)
	}

	var stats AthleteStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("decode stats response: %w", err)
	}

	return &stats, nil
}

func FetchGear(ctx context.Context, httpClient *http.Client, apiBase, accessToken string, gearID string) (*GearDetail, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	url := fmt.Sprintf("%s/gear/%s", apiBase, gearID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create gear request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch gear: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava gear API error: HTTP %d", resp.StatusCode)
	}

	var gear GearDetail
	if err := json.NewDecoder(resp.Body).Decode(&gear); err != nil {
		return nil, fmt.Errorf("decode gear response: %w", err)
	}

	return &gear, nil
}
