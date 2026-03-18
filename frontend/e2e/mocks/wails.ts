/**
 * Wails backend mock for Playwright e2e tests.
 *
 * This script is injected into every page via page.addInitScript() before
 * the Svelte app loads. It stubs window.go.main.App.* and window.runtime.*
 * so tests run without a live Go backend.
 *
 * Individual tests can override specific functions by calling
 * page.addInitScript() again after this one, since scripts run in order.
 */

// ---------------------------------------------------------------------------
// Default mock state — can be overridden per-test
// ---------------------------------------------------------------------------
const DEFAULT_SETTINGS = {
  ollamaEndpoint: 'http://localhost:11434',
  ollamaModel: '',
  customSystemPrompt: '',
}

const DEFAULT_PROFILE = {
  age: 32,
  maxHR: 185,
  thresholdPaceSecs: 300,
  weeklyMileageTarget: 60,
  raceGoals: 'Sub-3:30 marathon',
  injuryHistory: 'None',
  experienceLevel: 'intermediate',
  trainingDaysPerWeek: 5,
  restingHR: 48,
  preferredTerrain: 'road',
  heartRateZones: '[{"min":0,"max":115},{"min":115,"max":152},{"min":152,"max":171},{"min":171,"max":190},{"min":190,"max":-1}]',
}

const DEFAULT_ACTIVITIES = [
  {
    name: 'Morning Run',
    activityType: 'Run',
    startDate: '2026-03-10T07:00:00Z',
    distance: 10.5,
    durationSecs: 3150,
    avgPaceSecs: 300,
    avgHR: 155,
  },
  {
    name: 'Long Run',
    activityType: 'Run',
    startDate: '2026-03-08T08:00:00Z',
    distance: 21.0,
    durationSecs: 6720,
    avgPaceSecs: 320,
    avgHR: 162,
  },
]

const DEFAULT_INSIGHTS = [
  {
    id: 1,
    content: 'Focus on easy aerobic base building for the next 4 weeks.',
    sourceSessionId: 'sess-001',
    createdAt: '2026-03-01T10:00:00Z',
  },
]

const DEFAULT_STATS = {
  totalCount: 42,
  totalDistanceKm: 380.5,
  earliestDate: '2025-10-01',
  latestDate: '2026-03-10',
}

const DEFAULT_RACES = [
  {
    id: 'race_1',
    name: 'Berlin Marathon',
    distanceKm: 42.195,
    raceDate: '2026-10-15T00:00:00Z',
    terrain: 'road',
    elevationM: null,
    goalTimeSec: 12600,
    priority: 'A',
    isActive: true,
    createdAt: '2026-03-01T00:00:00Z',
  },
]

const DEFAULT_ACTIVE_PLAN = {
  id: 'plan_1',
  raceId: 'race_1',
  generatedAt: '2026-03-15T00:00:00Z',
  llmBackend: 'ollama',
  promptHash: 'abc123',
}

const DEFAULT_PLAN_WEEKS = [
  {
    id: 'plan_1_w1',
    planId: 'plan_1',
    weekNumber: 1,
    weekStart: '2026-03-16T00:00:00Z',
    sessions: [
      { id: 'plan_1_w1_s0', weekId: 'plan_1_w1', dayOfWeek: 1, type: 'easy', durationMin: 45, distanceKm: 8, hrZone: 2, paceMinLow: null, paceMinHigh: null, notes: 'Easy aerobic run', status: 'planned', actualDurationMin: null, actualDistanceKm: null, completedAt: null },
      { id: 'plan_1_w1_s1', weekId: 'plan_1_w1', dayOfWeek: 3, type: 'tempo', durationMin: 50, distanceKm: 10, hrZone: 3, paceMinLow: null, paceMinHigh: null, notes: 'Tempo at threshold', status: 'planned', actualDurationMin: null, actualDistanceKm: null, completedAt: null },
      { id: 'plan_1_w1_s2', weekId: 'plan_1_w1', dayOfWeek: 6, type: 'long_run', durationMin: 90, distanceKm: 18, hrZone: 2, paceMinLow: null, paceMinHigh: null, notes: 'Long slow distance', status: 'planned', actualDurationMin: null, actualDistanceKm: null, completedAt: null },
      { id: 'plan_1_w1_s3', weekId: 'plan_1_w1', dayOfWeek: 7, type: 'rest', durationMin: 0, distanceKm: null, hrZone: null, paceMinLow: null, paceMinHigh: null, notes: '', status: 'planned', actualDurationMin: null, actualDistanceKm: null, completedAt: null },
    ],
  },
  {
    id: 'plan_1_w2',
    planId: 'plan_1',
    weekNumber: 2,
    weekStart: '2026-03-23T00:00:00Z',
    sessions: [
      { id: 'plan_1_w2_s0', weekId: 'plan_1_w2', dayOfWeek: 1, type: 'easy', durationMin: 40, distanceKm: 7, hrZone: 2, paceMinLow: null, paceMinHigh: null, notes: 'Recovery run', status: 'planned', actualDurationMin: null, actualDistanceKm: null, completedAt: null },
      { id: 'plan_1_w2_s1', weekId: 'plan_1_w2', dayOfWeek: 4, type: 'intervals', durationMin: 55, distanceKm: 12, hrZone: 4, paceMinLow: null, paceMinHigh: null, notes: '6x800m at 5K pace', status: 'planned', actualDurationMin: null, actualDistanceKm: null, completedAt: null },
      { id: 'plan_1_w2_s2', weekId: 'plan_1_w2', dayOfWeek: 6, type: 'long_run', durationMin: 100, distanceKm: 20, hrZone: 2, paceMinLow: null, paceMinHigh: null, notes: 'Progressive long run', status: 'planned', actualDurationMin: null, actualDistanceKm: null, completedAt: null },
    ],
  },
]

// ---------------------------------------------------------------------------
// Install mocks on window before Svelte app boots
// ---------------------------------------------------------------------------
window.__WAILS_MOCK_STATE__ = {
  isFirstRun: false,
  settings: { ...DEFAULT_SETTINGS },
  profile: { ...DEFAULT_PROFILE },
  activities: [...DEFAULT_ACTIVITIES],
  insights: [...DEFAULT_INSIGHTS],
  stats: { ...DEFAULT_STATS },
  races: DEFAULT_RACES.map((r) => ({ ...r })),
  activePlan: { ...DEFAULT_ACTIVE_PLAN },
  planWeeks: DEFAULT_PLAN_WEEKS.map((w) => ({ ...w, sessions: w.sessions.map((s) => ({ ...s })) })),
  syncStatus: { enabled: false, provider: '', lastSyncedAt: '', lastChatSyncAt: '', syncing: false, lastError: '' },
  stravaConnected: false,
  ollamaModels: [],
  chatResponse: 'Great question! Based on your recent training data, I recommend a steady-state run at 5:10/km for 45 minutes tomorrow.',
}

// Helper: simulate async delay
function mockAsync(value, delayMs = 50) {
  return new Promise((resolve) => setTimeout(() => resolve(value), delayMs))
}

// Install the window.go namespace that Wails generates
window.go = {
  main: {
    App: {
      IsFirstRun: () => mockAsync(window.__WAILS_MOCK_STATE__.isFirstRun),

      GetSettingsData: () => mockAsync({ ...window.__WAILS_MOCK_STATE__.settings }),
      SaveSettingsData: (data) => {
        window.__WAILS_MOCK_STATE__.settings = { ...window.__WAILS_MOCK_STATE__.settings, ...data }
        return mockAsync(null)
      },

      GetProfileData: () => mockAsync({ ...window.__WAILS_MOCK_STATE__.profile }),
      SaveProfileData: (data) => {
        window.__WAILS_MOCK_STATE__.profile = { ...window.__WAILS_MOCK_STATE__.profile, ...data }
        return mockAsync(null)
      },

      GetRecentActivities: (limit) => mockAsync(window.__WAILS_MOCK_STATE__.activities.slice(0, limit)),
      GetActivityStats: () => mockAsync({ ...window.__WAILS_MOCK_STATE__.stats }),

      GetPinnedInsights: () => mockAsync([...window.__WAILS_MOCK_STATE__.insights]),
      SaveInsight: (content) => {
        const exists = window.__WAILS_MOCK_STATE__.insights.some((i) => i.content === content)
        if (!exists) {
          window.__WAILS_MOCK_STATE__.insights.push({
            id: Date.now(),
            content,
            sourceSessionId: 'mock-session',
            createdAt: new Date().toISOString(),
          })
        }
        return mockAsync(null)
      },
      DeletePinnedInsight: (id) => {
        window.__WAILS_MOCK_STATE__.insights = window.__WAILS_MOCK_STATE__.insights.filter((i) => i.id !== id)
        return mockAsync(null)
      },

      GetStravaAuthStatus: () => mockAsync({ connected: window.__WAILS_MOCK_STATE__.stravaConnected }),
      StartStravaAuth: () => mockAsync(null, 200),
      DisconnectStrava: () => {
        window.__WAILS_MOCK_STATE__.stravaConnected = false
        return mockAsync(null)
      },
      SyncStravaActivities: () => mockAsync(null, 300),

      GetStravaCredentialsAvailable: () => mockAsync(true),

      SendMessage: (msg) => mockAsync(window.__WAILS_MOCK_STATE__.chatResponse, 300),

      GetOllamaModels: (endpoint) => mockAsync([...window.__WAILS_MOCK_STATE__.ollamaModels]),

      GetContextPreview: () => mockAsync('# CoachLM — Running Coach\n\n## Role\nYou are CoachLM...'),

      ExportContext: (filePath) => mockAsync(null),
      ExportContextWithDialog: () => mockAsync(null),
      ImportContext: (filePath, replaceAll) => mockAsync(null),
      ImportContextWithDialog: (replaceAll) => mockAsync(null),

      ImportFITFile: (filePath) => mockAsync(null),

      // Plan bindings
      CreateRace: (race) => {
        race.id = 'race_' + Date.now()
        race.createdAt = new Date().toISOString()
        race.isActive = false
        window.__WAILS_MOCK_STATE__.races.push(race)
        return mockAsync(race)
      },
      UpdateRace: (race) => {
        const idx = window.__WAILS_MOCK_STATE__.races.findIndex((r) => r.id === race.id)
        if (idx >= 0) window.__WAILS_MOCK_STATE__.races[idx] = { ...window.__WAILS_MOCK_STATE__.races[idx], ...race }
        return mockAsync(null)
      },
      DeleteRace: (id) => {
        window.__WAILS_MOCK_STATE__.races = window.__WAILS_MOCK_STATE__.races.filter((r) => r.id !== id)
        if (window.__WAILS_MOCK_STATE__.activePlan && window.__WAILS_MOCK_STATE__.activePlan.raceId === id) {
          window.__WAILS_MOCK_STATE__.activePlan = null
          window.__WAILS_MOCK_STATE__.planWeeks = []
        }
        return mockAsync(null)
      },
      ListRaces: () => mockAsync(window.__WAILS_MOCK_STATE__.races.map((r) => ({ ...r }))),
      SetActiveRace: (id) => {
        window.__WAILS_MOCK_STATE__.races.forEach((r) => (r.isActive = r.id === id))
        return mockAsync(null)
      },
      GeneratePlan: (raceId) => {
        const plan = { id: 'plan_' + Date.now(), raceId, generatedAt: new Date().toISOString(), llmBackend: 'ollama', promptHash: 'mock' }
        window.__WAILS_MOCK_STATE__.activePlan = plan
        return mockAsync(plan)
      },
      GetActivePlan: () => {
        const plan = window.__WAILS_MOCK_STATE__.activePlan
        return mockAsync(plan ? { ...plan } : null)
      },
      GetPlanWeeks: (planId) => {
        return mockAsync(window.__WAILS_MOCK_STATE__.planWeeks.map((w) => ({ ...w, sessions: (w.sessions || []).map((s) => ({ ...s })) })))
      },
      UpdateSessionStatus: (sessionId, status, actual) => {
        for (const w of window.__WAILS_MOCK_STATE__.planWeeks) {
          const s = (w.sessions || []).find((s) => s.id === sessionId)
          if (s) {
            s.status = status
            if (actual) {
              if (actual.durationMin != null) s.actualDurationMin = actual.durationMin
              if (actual.distanceKm != null) s.actualDistanceKm = actual.distanceKm
            }
            if (status === 'completed') s.completedAt = new Date().toISOString()
            break
          }
        }
        return mockAsync(null)
      },
      
      ConnectS3: (endpoint, bucket, accessKey, secretKey) => {
        window.__WAILS_MOCK_STATE__.syncStatus.enabled = true
        window.__WAILS_MOCK_STATE__.syncStatus.provider = 'S3'
        return mockAsync(null)
      },
      ConnectGoogleDrive: () => {
        window.__WAILS_MOCK_STATE__.syncStatus.enabled = true
        window.__WAILS_MOCK_STATE__.syncStatus.provider = 'Google Drive'
        return mockAsync(null)
      },
      DisconnectCloud: () => {
        window.__WAILS_MOCK_STATE__.syncStatus.enabled = false
        window.__WAILS_MOCK_STATE__.syncStatus.provider = ''
        return mockAsync(null)
      },
      SyncNow: () => mockAsync(null),
      GetSyncStatus: () => mockAsync({ ...window.__WAILS_MOCK_STATE__.syncStatus }),
      ExportChatSessions: () => mockAsync(new Uint8Array()),
      ImportChatSessions: (data, replaceAll) => mockAsync(null),
    },
  },
}

// Install window.runtime (Wails dialog functions)
window.runtime = {
  EventsOn: (event, callback) => () => {},
  EventsOff: (event) => {},
  BrowserOpenURL: (url) => {},
}
