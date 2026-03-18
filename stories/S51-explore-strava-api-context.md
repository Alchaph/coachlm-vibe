---
id: S51
title: Explore and integrate Strava API data into context framework
status: done
created: 2026-03-18
updated: 2026-03-18
---

# S51 — Explore and integrate Strava API data into context framework

## User story

As a runner using CoachLM, I want the LLM to have richer training data from Strava beyond basic activity summaries, so that coaching advice is more informed and personalized using detailed metrics like heart rate zones, gear mileage, training load statistics, and route information.

## Acceptance criteria

- [ ] Document exploration results: list all viable Strava API endpoints with their value for coaching context
- [ ] Implement fetching and storing athlete training zones (HR zones from `/athlete/zones`)
- [ ] Implement fetching and storing athlete statistics (training load from `/athletes/:id/stats`)
- [ ] Implement fetching and storing gear details (shoe mileage from `/gear/:id`)
- [ ] Integrate fetched Strava data into context assembly (assembler.go)
- [ ] Add rate limiting and retry logic for Strava API calls
- [ ] Wire activity streams fetching into webhook processing (currently implemented but not called)
- [ ] Ensure new data sources respect token budget (training summary still lowest priority, pinned insights sacred)
- [ ] Add unit tests for new Strava endpoint clients
- [ ] Add unit tests for context assembly with new data sources
- [ ] Update existing tests to account for new context data
- [ ] All tests pass: `go test ./...`

## Technical notes

### Phase 1: Exploration & Documentation

**Strava API v3 endpoints to evaluate** (based on official docs):

**High Priority** (core training insights):
- `GET /api/v3/athlete/zones` — HR and power zones for intensity analysis
- `GET /api/v3/athletes/:id/stats` — recent 4 weeks, YTD, all-time totals for training load tracking
- `GET /api/v3/gear/:id` — shoe/bike mileage for equipment tracking
- `GET /api/v3/activities/:id/streams` — raw time-series data (HR, pace, cadence, power)

**Medium Priority** (enrichment):
- `GET /api/v3/segments/starred` — favorite routes for benchmarking
- `GET /api/v3/segments/:id` — route metadata (grade, distance, elevation)
- `GET /api/v3/athletes/:id/koms` — achievements for strengths analysis

**Lower Priority** (contextual):
- `GET /api/v3/routes/*` — route planning and metadata
- `GET /api/v3/segments/:id/leaderboard` — peer comparison

**Rate limits** (must respect):
- Non-upload limit: 100 requests / 15 min, 1000 requests / day
- Overall limit: 200 requests / 15 min, 2000 requests / day
- Track via `X-RateLimit-Usage` headers
- Implement exponential backoff and Retry-After handling for 429 responses

**Avoid**:
- Running races endpoints (deprecated Oct 2021)
- Activity comments (social, not training data)

### Phase 2: Implementation - New Strava Endpoints

**Backend files to create/modify**:

`internal/strava/stats.go` (new file):
- `FetchAthleteStats(accessToken string) (*AthleteStats, error)`
- `FetchAthleteZones(accessToken string) (*AthleteZones, error)`
- `FetchGear(accessToken string, gearID string) (*Gear, error)`

Data structures (add to `internal/strava/` or `internal/storage/`):
```go
type AthleteStats struct {
    RecentRunTotals *Totals `json:"recent_run_totals"`
    YTDRunTotals   *Totals `json:"ytd_run_totals"`
    AllRunTotals   *Totals `json:"all_run_totals"`
}

type Totals struct {
    Count          int     `json:"count"`
    Distance       float64 `json:"distance"`       // meters
    MovingTime     int     `json:"moving_time"`    // seconds
    ElevationGain  float64 `json:"elevation_gain"` // meters
    AchievementCount int   `json:"achievement_count"`
}

type AthleteZones struct {
    HeartRate *HeartRateZones `json:"heart_rate"`
    Power     *PowerZones     `json:"power"`
}

type HeartRateZones struct {
    CustomZones bool           `json:"custom_zones"`
    Zones       []ZoneRange    `json:"zones"`
}

type ZoneRange struct {
    Min int `json:"min"`
    Max int `json:"max"`
}

type Gear struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Distance    float64 `json:"distance"`    // meters
    BrandName   string  `json:"brand_name"`
    ModelName   string  `json:"model_name"`
    Description string  `json:"description"`
}
```

`internal/storage/stats.go` (new file):
- DB table: `athlete_stats` (singleton, updated periodically)
- `SaveAthleteStats(stats *storage.AthleteStats) error`
- `GetAthleteStats() (*storage.AthleteStats, error)`

`internal/storage/zones.go` (new file):
- DB table: `athlete_zones` (singleton, cached)
- `SaveAthleteZones(zones *storage.AthleteZones) error`
- `GetAthleteZones() (*storage.AthleteZones, error)`

`internal/storage/gear.go` (new file):
- DB table: `gear` (multiple rows, gear_id unique)
- `SaveGear(gear *storage.Gear) error`
- `GetGear(gearID string) (*storage.Gear, error)`
- `ListGear() ([]*storage.Gear, error)`

### Phase 3: Integration with Context Engine

**Modify `internal/context/assembler.go`**:

Add to `PromptInput` struct:
```go
type PromptInput struct {
    Profile      *storage.AthleteProfile
    Activities   []storage.Activity
    Insights     []storage.PinnedInsight
    CustomPrompt string
    Stats        *storage.AthleteStats      // NEW
    Zones        *storage.AthleteZones      // NEW
    Gear         []*storage.Gear            // NEW
    Now          time.Time
}
```

Add new formatters (in `internal/context/`):
- `internal/context/stats.go` (new): `FormatStatsBlock(stats *storage.AthleteStats) string`
- `internal/context/zones.go` (new): `FormatZonesBlock(zones *storage.AthleteZones) string`
- `internal/context/gear.go` (new): `FormatGearBlock(gear []*storage.Gear) string`

Update `AssemblePrompt()` to include new sections in priority order:
1. Sacred: preamble + custom prompt + pinned insights
2. Profile block
3. Training load stats block (NEW - medium priority)
4. Zones block (NEW - medium priority)
5. Gear block (NEW - medium priority)
6. Training summary (lowest priority, truncated first)

**Modify `app.go`**:

In `SendMessage()` and `GetContextPreview()`:
```go
stats := a.db.GetAthleteStats()
zones := a.db.GetAthleteZones()
gear := a.db.ListGear()

input := coachctx.PromptInput{
    Profile:    profile,
    Activities: activities,
    Insights:   insights,
    CustomPrompt: settings.CustomSystemPrompt,
    Stats:      stats,      // NEW
    Zones:      zones,      // NEW
    Gear:       gear,       // NEW
    Now:        time.Now(),
}
```

### Phase 4: Sync and Rate Limiting

**Implement retry wrapper** (`internal/strava/http.go` - new file):
```go
type RateLimitedClient struct {
    client    *http.Client
    rateLimit *rateLimitTracker
}

func (c *RateLimitedClient) Do(req *http.Request) (*http.Response, error) {
    // Check rate limits from previous responses
    // Wait if approaching limit
    // Retry on 429 with Retry-After header
    // Exponential backoff on 5xx errors
}
```

Use this client in:
- `oauth.go` (Exchange, Refresh)
- `sync.go` (FetchAthleteActivities)
- `stats.go` (FetchAthleteStats, FetchAthleteZones, FetchGear)
- `webhook.go` (FetchActivity, FetchStreams)
- `streams.go` (FetchStreams)

**Sync strategy for new endpoints**:
- Fetch zones and stats on Strava sync completion (after activities list fetch)
- Fetch gear for each activity with a non-empty `gear_id` field
- Cache zones and stats in DB (update daily or weekly, not on every activity)
- Gear mileage updates automatically when activities are fetched (Strava returns updated distance)

### Phase 5: Activity Streams Integration

**Current state**: `FetchStreams` and `parseStreams` implemented but not called by webhook.

**Modify `internal/strava/webhook.go`**:

In `processEvent()`, after `FetchActivity()`:
```go
streams, err := wh.FetchStreams(accessToken, activityID)
if err == nil && streams != nil {
    // Persist each stream type
    if streams.HeartRate != nil {
        a.db.SaveActivityStream(activityID, "heartrate", streams.HeartRate)
    }
    if streams.Pace != nil {
        a.db.SaveActivityStream(activityID, "pace", streams.Pace)
    }
    if streams.Cadence != nil {
        a.db.SaveActivityStream(activityID, "cadence", streams.Cadence)
    }
}
```

**Context integration** (option - may be separate story):
- Stream summaries (HR zones per activity, max HR, HR variability) are valuable but large
- Implement stream summarization function to derive compact metrics
- Include in training summary or separate block with careful token budgeting

### Files to modify

**New files**:
- `internal/strava/stats.go`
- `internal/strava/http.go` (rate limiting wrapper)
- `internal/context/stats.go`
- `internal/context/zones.go`
- `internal/context/gear.go`
- `internal/storage/stats.go`
- `internal/storage/zones.go`
- `internal/storage/gear.go`

**Modify existing files**:
- `internal/strava/webhook.go` (wire FetchStreams, call new endpoints)
- `internal/strava/sync.go` (add rate limiting client)
- `internal/strava/oauth.go` (add rate limiting client)
- `internal/strava/streams.go` (no changes, already implemented)
- `internal/context/assembler.go` (add fields to PromptInput, update AssemblePrompt)
- `app.go` (fetch new data, pass to AssemblePrompt)
- `internal/storage/migrations.go` (add new tables: athlete_stats, athlete_zones, gear)

### Testing requirements

**Unit tests**:
- `internal/strava/stats_test.go`: test FetchAthleteStats, FetchAthleteZones, FetchGear
- `internal/strava/http_test.go`: test rate limiting, retry logic, 429 handling
- `internal/context/assembler_test.go`: test new blocks (stats, zones, gear) in assembly
- `internal/context/stats_test.go`: test FormatStatsBlock
- `internal/context/zones_test.go`: test FormatZonesBlock
- `internal/context/gear_test.go`: test FormatGearBlock
- Test token budget behavior: new blocks truncated before training summary
- Test pinned insights still sacred (never truncated)

**Integration tests**:
- Test full sync flow: OAuth → activities → stats → zones → gear
- Test webhook flow: event → activity → streams → stats update
- Test rate limiting: throttle requests correctly on 429

**E2e tests** (optional):
- Verify Context UI displays new data (stats, zones, gear)
- Verify new data persists across app restarts

## Out of scope

- Strava route planning endpoints (separate feature, not core coaching)
- Segment leaderboards (peer comparison is nice-to-have, not essential)
- Raw stream data in prompts (too large, implement summarization separately)
- Modifying existing LLM prompts (context changes automatically, prompt structure unchanged)
- UI changes for displaying new data (may be separate story, focus on backend integration here)

---

## Status history

| Date | Status | Notes |
|------|--------|-------|
| 2026-03-18 | draft | Created story based on Strava API research and context framework analysis |
| 2026-03-18 | in-progress | Implementation started: rate limiter, stats/gear endpoints, storage, context engine, app wiring |
| 2026-03-18 | done | All layers implemented and tested. go test ./... passes (9 packages). 122 Playwright e2e tests pass. |

<!-- Agent: add a Blocker section here if status is set to failed -->
