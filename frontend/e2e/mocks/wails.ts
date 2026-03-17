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
  activeLlm: 'free',
  claudeApiKey: '',
  openaiApiKey: '',
  ollamaEndpoint: 'http://localhost:11434',
  stravaClientId: '',
  stravaClientSecret: '',
  claudeModel: '',
  openaiModel: '',
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
        // Simulate LLM reload — if "free" with no key, silently succeed (settings still save)
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

      SendMessage: (msg) => mockAsync(window.__WAILS_MOCK_STATE__.chatResponse, 300),

      GetOllamaModels: (endpoint) => mockAsync([...window.__WAILS_MOCK_STATE__.ollamaModels]),

      GetContextPreview: () => mockAsync('# CoachLM — Running Coach\n\n## Role\nYou are CoachLM...'),

      ExportContext: (filePath) => mockAsync(null),
      ImportContext: (filePath, replaceAll) => mockAsync(null),

      ImportFITFile: (filePath) => mockAsync(null),
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
