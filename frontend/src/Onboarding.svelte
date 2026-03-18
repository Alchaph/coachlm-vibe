<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte'
  import { SaveSettingsData, StartStravaAuth, GetOllamaModels, SyncStravaActivities, GetProfileData, GetPinnedInsights, GetRecentActivities, GetStravaCredentialsAvailable } from '../wailsjs/go/main/App.js'

  const dispatch = createEventDispatcher()

  let step = 1
  let connectingStrava = false
  let stravaConnected = false
  let stravaCredentialsAvailable = false
  let saving = false
  let error = ''

  onMount(async () => {
    try {
      stravaCredentialsAvailable = !!(await GetStravaCredentialsAvailable())
    } catch (_) {}
  })
  // Context readiness (step 5)
  let hasProfile = false
  let hasTrainingData = false
  let hasInsights = false

  $: if (step === 3) checkContextReadiness()

  function next() {
    if (step < 3) step++
  }

  function back() {
    if (step > 1) step--
  }

  async function connectStrava() {
    connectingStrava = true
    error = ''
    try {
      await StartStravaAuth()
      stravaConnected = true
      SyncStravaActivities().catch(() => {})
      next()
    } catch (e: any) {
      error = e?.message || 'Failed to connect Strava'
    } finally {
      connectingStrava = false
    }
  }
  async function checkContextReadiness() {
    try {
      const [profile, insights, activities] = await Promise.all([
        GetProfileData().catch(() => null),
        GetPinnedInsights().catch(() => []),
        GetRecentActivities(1).catch(() => [])
      ])
      hasProfile = !!(profile && (profile.age > 0 || profile.maxHR > 0 || profile.thresholdPaceSecs > 0 || profile.weeklyMileageTarget > 0 || profile.raceGoals || profile.injuryHistory))
      hasTrainingData = (activities || []).length > 0
      hasInsights = (insights || []).length > 0
    } catch (_) {}
  }

  async function finish() {
    saving = true
    error = ''
    try {
      await SaveSettingsData({
        ollamaEndpoint: 'http://localhost:11434',
        ollamaModel: '',
        customSystemPrompt: ''
      })
      dispatch('complete')
    } catch (e: any) {
      error = e?.message || 'Failed to save settings'
    } finally {
      saving = false
    }
  }
</script>

<div class="overlay">
  <div class="wizard">
    <div class="progress">
      {#each [1, 2, 3] as s}
        <div class="dot" class:active={s === step} class:done={s < step}></div>
      {/each}
    </div>

    {#if error}
      <div class="error-msg">{error}</div>
    {/if}

    {#if step === 1}
      <div class="step">
        <h1>Welcome to CoachLM</h1>
        <p class="subtitle">Your AI-powered running coach. Let's get you set up in a few steps.</p>
        <div class="actions">
          <button class="btn btn-primary" on:click={next}>Get Started</button>
        </div>
      </div>
    {/if}

    {#if step === 2}
      <div class="step">
        <h1>Connect Strava</h1>
        <p class="subtitle">Sync your activities automatically. You can skip this and set it up later.</p>

        <div class="actions">
          <button class="btn btn-secondary" on:click={back}>Back</button>
          <button class="btn btn-secondary" on:click={next}>Skip</button>
          {#if stravaCredentialsAvailable}
            <button
              class="btn btn-primary"
              on:click={connectStrava}
              disabled={connectingStrava}
            >
              {connectingStrava ? 'Connecting...' : 'Connect Strava'}
            </button>
          {:else}
            <p class="field-note strava-unavailable">Not available in this build</p>
          {/if}
        </div>
      </div>
    {/if}

    {#if step === 3}
      <div class="step">
        <h1>You're All Set!</h1>
        <p class="subtitle">Start chatting with your AI running coach.</p>
        {#if stravaConnected}
          <p class="connected-note">Strava connected successfully.</p>
        {/if}
        <div class="context-readiness">
          <div class="readiness-item" class:ready={hasProfile}>
            <span class="readiness-icon">{hasProfile ? '\u2713' : '\u2717'}</span>
            <span>Athlete profile</span>
            {#if !hasProfile}
              <button class="btn-link" on:click={() => dispatch('openContext')}>Set up profile</button>
            {/if}
          </div>
          <div class="readiness-item" class:ready={hasTrainingData}>
            <span class="readiness-icon">{hasTrainingData ? '\u2713' : '\u2717'}</span>
            <span>Training data</span>
          </div>
          <div class="readiness-item" class:ready={hasInsights}>
            <span class="readiness-icon">{hasInsights ? '\u2713' : '\u2717'}</span>
            <span>Pinned insights</span>
          </div>
        </div>
        <div class="actions">
          <button class="btn btn-primary" on:click={finish} disabled={saving}>
            {saving ? 'Saving...' : 'Start Chatting'}
          </button>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(27, 38, 54, 0.98);
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .wizard {
    width: 100%;
    max-width: 480px;
  }

  .progress {
    display: flex;
    justify-content: center;
    gap: 10px;
    margin-bottom: 32px;
  }

  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.15);
    transition: background 0.3s;
  }

  .dot.active {
    background: #3b82f6;
  }

  .dot.done {
    background: #22c55e;
  }

  .error-msg {
    background: rgba(220, 53, 69, 0.15);
    color: #f87171;
    border: 1px solid rgba(220, 53, 69, 0.3);
    padding: 10px 16px;
    border-radius: 8px;
    font-size: 0.9rem;
    text-align: center;
    margin-bottom: 16px;
  }

  .step {
    text-align: center;
  }

  h1 {
    font-size: 1.5rem;
    font-weight: 700;
    color: #e2e8f0;
    margin: 0 0 8px;
  }

  .subtitle {
    color: #94a3b8;
    font-size: 0.95rem;
    margin: 0 0 28px;
    line-height: 1.5;
  }

  .connected-note {
    color: #22c55e;
    font-size: 0.9rem;
    margin: 0 0 20px;
  }

  .field-note {
    font-size: 0.85rem;
    color: #22c55e;
    margin-bottom: 12px;
    font-style: italic;
  }

  .actions {
    display: flex;
    justify-content: center;
    gap: 12px;
    margin-top: 24px;
  }

  .btn {
    padding: 10px 24px;
    border: none;
    border-radius: 12px;
    font-size: 0.9rem;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
    font-family: inherit;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .btn-secondary {
    background: rgba(255, 255, 255, 0.08);
    color: #94a3b8;
    border: 1px solid rgba(255, 255, 255, 0.15);
  }

  .btn-secondary:hover:not(:disabled) {
    color: #e2e8f0;
    background: rgba(255, 255, 255, 0.12);
  }

  .context-readiness {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    margin: 20px auto 24px;
    padding: 16px 24px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    width: fit-content;
  }

  .readiness-item {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 0.9rem;
    color: #94a3b8;
  }

  .readiness-item.ready {
    color: #e2e8f0;
  }

  .readiness-icon {
    font-size: 1rem;
    width: 18px;
    text-align: center;
    color: #f87171;
  }

  .readiness-item.ready .readiness-icon {
    color: #22c55e;
  }

  .strava-unavailable {
    color: #94a3b8;
    font-style: italic;
  }

  .btn-link {
    background: none;
    border: none;
    color: #3b82f6;
    text-decoration: underline;
    cursor: pointer;
    padding: 0;
    font-size: 0.9rem;
    font-family: inherit;
    margin-left: 8px;
  }
  .btn-link:hover {
    color: #2563eb;
  }
</style>
