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
      ImportContext: (filePath, replaceAll) => mockAsync(null),

      ImportFITFile: (filePath) => mockAsync(null),
      
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
  DialogSaveFile: (opts) =>
    mockAsync('/tmp/mock-export.coachctx'),
  DialogOpenFile: (opts) =>
    mockAsync('/tmp/mock-import.coachctx'),
  EventsOn: (event, callback) => () => {},
  EventsOff: (event) => {},
  BrowserOpenURL: (url) => {},
}
