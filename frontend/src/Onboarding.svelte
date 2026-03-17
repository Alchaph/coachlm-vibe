<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { SaveSettingsData, StartStravaAuth, GetOllamaModels, SaveProfileData, SyncStravaActivities, GetProfileData, GetPinnedInsights, GetRecentActivities } from '../wailsjs/go/main/App.js'

  const dispatch = createEventDispatcher()

  let step = 1
  let activeLlm = 'free'
  let claudeApiKey = ''
  let openaiApiKey = ''
  let ollamaEndpoint = 'http://localhost:11434'
  let stravaClientId = ''
  let stravaClientSecret = ''
  let claudeModel = ''
  let openaiModel = ''
  let ollamaModel = ''
  let connectingStrava = false
  let stravaConnected = false
  let saving = false
  let error = ''

  let showApiKey = false
  let ollamaModels: string[] = []
  let fetchingModels = false
  let modelFetchError = ''

  // Profile fields (step 4)
  let profileAge = 0
  let profileMaxHR = 0
  let profileThresholdMins = 0
  let profileThresholdSecs = 0
  let profileWeeklyMileage = 0
  let profileRaceGoals = ''
  let profileInjuryHistory = ''
  let profileExperienceLevel = ''
  let profileTrainingDaysPerWeek = 0
  let profileRestingHR = 0
  let profilePreferredTerrain = ''
  let savingProfile = false

  // Context readiness (step 5)
  let hasProfile = false
  let hasTrainingData = false
  let hasInsights = false

  async function fetchOllamaModels() {
    fetchingModels = true
    modelFetchError = ''
    ollamaModels = []
    try {
      ollamaModels = await GetOllamaModels(ollamaEndpoint) || []
      if (ollamaModels.length === 0) {
        modelFetchError = 'No models installed. Run: ollama pull llama3'
      }
    } catch (e: any) {
      modelFetchError = e?.message || 'Cannot reach Ollama'
    } finally {
      fetchingModels = false
    }
  }

  function next() {
    if (step < 5) step++
  }

  function back() {
    if (step > 1) step--
  }

  async function connectStrava() {
    if (!stravaClientId || !stravaClientSecret) return
    connectingStrava = true
    error = ''
    try {
      await SaveSettingsData({
        claudeApiKey,
        openaiApiKey,
        activeLlm,
        ollamaEndpoint,
        stravaClientId,
        stravaClientSecret,
        claudeModel,
        openaiModel,
        ollamaModel
      })
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

  async function saveProfile() {
    savingProfile = true
    error = ''
    try {
      const totalSecs = (profileThresholdMins * 60) + profileThresholdSecs
      await SaveProfileData({
        age: profileAge,
        maxHR: profileMaxHR,
        thresholdPaceSecs: totalSecs,
        weeklyMileageTarget: profileWeeklyMileage,
        raceGoals: profileRaceGoals,
        injuryHistory: profileInjuryHistory,
        experienceLevel: profileExperienceLevel,
        trainingDaysPerWeek: profileTrainingDaysPerWeek,
        restingHR: profileRestingHR,
        preferredTerrain: profilePreferredTerrain
      })
      hasProfile = profileAge > 0 || profileMaxHR > 0 || totalSecs > 0 || profileWeeklyMileage > 0 || profileRaceGoals !== '' || profileInjuryHistory !== '' || profileExperienceLevel !== '' || profileTrainingDaysPerWeek > 0 || profileRestingHR > 0 || profilePreferredTerrain !== ''
    } catch (e: any) {
      error = e?.message || 'Failed to save profile'
    } finally {
      savingProfile = false
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
        claudeApiKey,
        openaiApiKey,
        activeLlm,
        ollamaEndpoint,
        stravaClientId,
        stravaClientSecret,
        claudeModel,
        openaiModel,
        ollamaModel
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
      {#each [1, 2, 3, 4, 5] as s}
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
        <h1>Choose Your AI Backend</h1>
        <p class="subtitle">Select which LLM will power your coaching conversations.</p>

        <div class="form">
          <label class="field-label" for="onboarding-backend">Backend</label>
          <select id="onboarding-backend" bind:value={activeLlm}>
            <option value="free">Free (Gemini Flash)</option>
            <option value="claude">Claude</option>
            <option value="openai">OpenAI</option>
            <option value="local">Local (Ollama)</option>
          </select>

          {#if activeLlm === 'free'}
            <p class="field-note">No setup required - using built-in free tier API.</p>
          {/if}

          {#if activeLlm === 'claude'}
            <label class="field-label" for="onboarding-claude-api-key">Claude API Key</label>
            <div class="input-row">
              {#if showApiKey}
                <input id="onboarding-claude-api-key" type="text" bind:value={claudeApiKey} placeholder="sk-ant-..." />
              {:else}
                <input id="onboarding-claude-api-key" type="password" bind:value={claudeApiKey} placeholder="sk-ant-..." />
              {/if}
              <button class="toggle-btn" on:click={() => showApiKey = !showApiKey}>
                {showApiKey ? 'Hide' : 'Show'}
              </button>
            </div>
            <label class="field-label" for="onboarding-claude-model">Model</label>
            <input id="onboarding-claude-model" type="text" bind:value={claudeModel} placeholder="claude-sonnet-4-20250514" />
          {/if}

          {#if activeLlm === 'openai'}
            <label class="field-label" for="onboarding-openai-api-key">OpenAI API Key</label>
            <div class="input-row">
              {#if showApiKey}
                <input id="onboarding-openai-api-key" type="text" bind:value={openaiApiKey} placeholder="sk-..." />
              {:else}
                <input id="onboarding-openai-api-key" type="password" bind:value={openaiApiKey} placeholder="sk-..." />
              {/if}
              <button class="toggle-btn" on:click={() => showApiKey = !showApiKey}>
                {showApiKey ? 'Hide' : 'Show'}
              </button>
            </div>
            <label class="field-label" for="onboarding-openai-model">Model</label>
            <input id="onboarding-openai-model" type="text" bind:value={openaiModel} placeholder="gpt-4o" />
          {/if}

          {#if activeLlm === 'local'}
            <label class="field-label" for="onboarding-ollama-endpoint">Ollama Endpoint</label>
            <input id="onboarding-ollama-endpoint" type="text" bind:value={ollamaEndpoint} placeholder="http://localhost:11434" />
            <label class="field-label" for="onboarding-ollama-model">Model</label>
            <div class="input-row">
              <input id="onboarding-ollama-model" type="text" bind:value={ollamaModel} placeholder="llama3" />
              <button class="toggle-btn" on:click={fetchOllamaModels} disabled={fetchingModels}>
                {fetchingModels ? '...' : 'Fetch'}
              </button>
            </div>
            {#if modelFetchError}
              <p class="model-fetch-error">{modelFetchError}</p>
            {/if}
            {#if ollamaModels.length > 0}
              <div class="model-chips">
                {#each ollamaModels as model}
                  <button
                    class="model-chip"
                    class:selected={ollamaModel === model}
                    on:click={() => ollamaModel = model}
                  >
                    {model}
                  </button>
                {/each}
              </div>
            {/if}
          {/if}
        </div>

        <div class="actions">
          <button class="btn btn-secondary" on:click={back}>Back</button>
          <button class="btn btn-primary" on:click={next}>Next</button>
        </div>
      </div>
    {/if}

    {#if step === 3}
      <div class="step">
        <h1>Connect Strava</h1>
        <p class="subtitle">Sync your activities automatically. You can skip this and set it up later.</p>

        <div class="form">
          <label class="field-label" for="onboarding-strava-client-id">Client ID</label>
          <input id="onboarding-strava-client-id" type="text" bind:value={stravaClientId} placeholder="Your Strava Client ID" />

          <label class="field-label" for="onboarding-strava-client-secret">Client Secret</label>
          <input id="onboarding-strava-client-secret" type="password" bind:value={stravaClientSecret} placeholder="Your Strava Client Secret" />
        </div>

        <div class="actions">
          <button class="btn btn-secondary" on:click={back}>Back</button>
          <button class="btn btn-secondary" on:click={next}>Skip</button>
          <button
            class="btn btn-primary"
            on:click={connectStrava}
            disabled={connectingStrava || !stravaClientId || !stravaClientSecret}
          >
            {connectingStrava ? 'Connecting...' : 'Connect'}
          </button>
        </div>
      </div>
    {/if}

    {#if step === 4}
      <div class="step">
        <h1>Athlete Profile</h1>
        <p class="subtitle">Help your coach understand you. All fields are optional — you can update them later.</p>

        <div class="profile-form">
          <div class="form-row">
            <div class="field">
              <label class="field-label" for="onboarding-age">Age</label>
              <input id="onboarding-age" type="number" bind:value={profileAge} placeholder="30" min="1" max="120" />
            </div>
            <div class="field">
              <label class="field-label" for="onboarding-max-hr">Max Heart Rate</label>
              <input id="onboarding-max-hr" type="number" bind:value={profileMaxHR} placeholder="185" min="100" max="220" />
            </div>
          </div>
          <div class="form-row">
            <div class="field">
              <label class="field-label" for="onboarding-threshold-mins">Threshold Pace (/km)</label>
              <div class="pace-input">
                <input id="onboarding-threshold-mins" type="number" bind:value={profileThresholdMins} placeholder="5" min="0" max="15" />
                <span class="pace-sep">:</span>
                <input id="onboarding-threshold-secs" type="number" bind:value={profileThresholdSecs} placeholder="00" min="0" max="59" />
              </div>
            </div>
            <div class="field">
              <label class="field-label" for="onboarding-weekly-mileage">Weekly Mileage Target (km)</label>
              <input id="onboarding-weekly-mileage" type="number" bind:value={profileWeeklyMileage} placeholder="50" step="0.1" min="0" />
            </div>
          </div>
          <div class="field">
            <label class="field-label" for="onboarding-race-goals">Race Goals</label>
            <textarea id="onboarding-race-goals" bind:value={profileRaceGoals} placeholder="e.g. Sub-3:30 marathon in October" rows="2"></textarea>
          </div>
          <div class="field">
            <label class="field-label" for="onboarding-injury-history">Injury History</label>
            <textarea id="onboarding-injury-history" bind:value={profileInjuryHistory} placeholder="e.g. IT band issues in 2024, fully recovered" rows="2"></textarea>
          </div>
          <div class="form-row">
            <div class="field">
              <label class="field-label" for="onboarding-experience">Experience Level</label>
              <select id="onboarding-experience" bind:value={profileExperienceLevel}>
                <option value=""></option>
                <option value="beginner">Beginner</option>
                <option value="intermediate">Intermediate</option>
                <option value="advanced">Advanced</option>
                <option value="elite">Elite</option>
              </select>
            </div>
            <div class="field">
              <label class="field-label" for="onboarding-training-days">Training Days/Week</label>
              <input id="onboarding-training-days" type="number" bind:value={profileTrainingDaysPerWeek} placeholder="4" min="1" max="7" />
            </div>
          </div>
          <div class="form-row">
            <div class="field">
              <label class="field-label" for="onboarding-resting-hr">Resting HR</label>
              <input id="onboarding-resting-hr" type="number" bind:value={profileRestingHR} placeholder="50" min="30" max="120" />
            </div>
            <div class="field">
              <label class="field-label" for="onboarding-terrain">Preferred Terrain</label>
              <select id="onboarding-terrain" bind:value={profilePreferredTerrain}>
                <option value=""></option>
                <option value="road">Road</option>
                <option value="trail">Trail</option>
                <option value="track">Track</option>
                <option value="mixed">Mixed</option>
              </select>
            </div>
          </div>
        </div>

        <div class="actions">
          <button class="btn btn-secondary" on:click={back}>Back</button>
          <button class="btn btn-secondary" on:click={() => { checkContextReadiness(); next() }}>Skip</button>
          <button class="btn btn-primary" on:click={async () => { await saveProfile(); checkContextReadiness(); next() }} disabled={savingProfile}>
            {savingProfile ? 'Saving...' : 'Next'}
          </button>
        </div>
      </div>
    {/if}

    {#if step === 5}
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

  .form {
    text-align: left;
    margin-bottom: 24px;
  }

  .field-label {
    display: block;
    font-size: 0.8rem;
    color: #94a3b8;
    margin-bottom: 6px;
    margin-top: 12px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 600;
  }

  .field-note {
    font-size: 0.85rem;
    color: #22c55e;
    margin-bottom: 12px;
    font-style: italic;
  }

  select,
  input[type="text"],
  input[type="password"] {
    width: 100%;
    padding: 10px 14px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 12px;
    color: white;
    font-family: inherit;
    font-size: 0.95rem;
    outline: none;
    transition: border-color 0.2s;
  }

  select:focus,
  input:focus {
    border-color: #3b82f6;
  }

  select {
    appearance: none;
    cursor: pointer;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%2394a3b8' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 14px center;
    padding-right: 36px;
  }

  select option {
    background: #1b2636;
    color: white;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px 20px;
  }

  .form-row .field {
    display: flex;
    flex-direction: column;
  }

  .form-row select {
    width: 100%;
  }

  .form-row input {
    width: 100%;
  }

  .input-row {
    display: flex;
    gap: 8px;
  }

  .input-row input {
    flex: 1;
  }

  .toggle-btn {
    padding: 8px 14px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 12px;
    color: #94a3b8;
    font-size: 0.85rem;
    cursor: pointer;
    transition: color 0.2s, background 0.2s;
    white-space: nowrap;
  }

  .toggle-btn:hover {
    color: #e2e8f0;
    background: rgba(255, 255, 255, 0.12);
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

  .model-fetch-error {
    color: #f87171;
    font-size: 0.8rem;
    margin: 6px 0 0;
  }

  .model-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 10px;
  }

  .model-chip {
    padding: 6px 14px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 20px;
    color: #94a3b8;
    font-size: 0.85rem;
    cursor: pointer;
    transition: all 0.2s;
    font-family: inherit;
  }

  .model-chip:hover {
    color: #e2e8f0;
    background: rgba(255, 255, 255, 0.1);
  }

  .model-chip.selected {
    background: rgba(59, 130, 246, 0.2);
    border-color: #3b82f6;
    color: #3b82f6;
  }

  .profile-form {
    text-align: left;
    margin-bottom: 24px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px 20px;
  }

  .profile-form .field {
    display: flex;
    flex-direction: column;
  }

  .profile-form input[type="number"],
  .profile-form textarea {
    width: 100%;
    padding: 10px 14px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 12px;
    color: white;
    font-family: inherit;
    font-size: 0.95rem;
    outline: none;
    transition: border-color 0.2s;
    box-sizing: border-box;
  }

  .profile-form input:focus,
  .profile-form textarea:focus {
    border-color: #3b82f6;
  }

  .profile-form textarea {
    resize: vertical;
    min-height: 60px;
  }

  .profile-form textarea::placeholder,
  .profile-form input::placeholder {
    color: rgba(255, 255, 255, 0.3);
  }

  .pace-input {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .pace-input input {
    flex: 1;
    text-align: center;
  }

  .pace-sep {
    color: #94a3b8;
    font-size: 1.1rem;
    font-weight: 600;
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
</style>
