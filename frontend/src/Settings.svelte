<script lang="ts">
  import { onMount } from 'svelte'
  import {
    GetSettingsData,
    SaveSettingsData,
    GetStravaAuthStatus,
    StartStravaAuth,
    DisconnectStrava,
    GetOllamaModels
  } from '../wailsjs/go/main/App.js'

  let activeLlm = 'free'
  let claudeApiKey = ''
  let openaiApiKey = ''
  let ollamaEndpoint = 'http://localhost:11434'
  let stravaClientId = ''
  let stravaClientSecret = ''
  let claudeModel = ''
  let openaiModel = ''
  let ollamaModel = ''

  let stravaConnected = false
  let loading = true
  let saving = false
  let connectingStrava = false
  let feedback = ''
  let feedbackType: 'success' | 'error' = 'success'
  let feedbackTimer: ReturnType<typeof setTimeout> | null = null

  let showClaudeKey = false
  let showOpenaiKey = false
  let showStravaSecret = false

  let ollamaModels: string[] = []
  let fetchingModels = false
  let modelFetchError = ''

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

  function showFeedback(msg: string, type: 'success' | 'error') {
    feedback = msg
    feedbackType = type
    if (feedbackTimer) clearTimeout(feedbackTimer)
    feedbackTimer = setTimeout(() => { feedback = '' }, 3000)
  }

  onMount(async () => {
    try {
      const [settings, status] = await Promise.all([
        GetSettingsData(),
        GetStravaAuthStatus()
      ])

      if (settings) {
        activeLlm = settings.activeLlm || 'local'
        claudeApiKey = settings.claudeApiKey || ''
        openaiApiKey = settings.openaiApiKey || ''
        ollamaEndpoint = settings.ollamaEndpoint || 'http://localhost:11434'
        stravaClientId = settings.stravaClientId || ''
        stravaClientSecret = settings.stravaClientSecret || ''
        claudeModel = settings.claudeModel || ''
        openaiModel = settings.openaiModel || ''
        ollamaModel = settings.ollamaModel || ''
      }

      if (status) {
        stravaConnected = !!status.connected
      }
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to load settings', 'error')
    } finally {
      loading = false
    }
  })

  async function save() {
    saving = true
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
      showFeedback('Settings saved!', 'success')
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to save settings', 'error')
    } finally {
      saving = false
    }
  }

  async function connectStrava() {
    connectingStrava = true
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
      const status = await GetStravaAuthStatus()
      stravaConnected = !!status?.connected
      showFeedback('Strava connected!', 'success')
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to connect Strava', 'error')
    } finally {
      connectingStrava = false
    }
  }

  async function disconnectStrava() {
    if (!confirm('Disconnect Strava? Your synced activities will remain.')) return
    try {
      await DisconnectStrava()
      stravaConnected = false
      showFeedback('Strava disconnected', 'success')
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to disconnect', 'error')
    }
  }
</script>

<div class="settings">
  {#if loading}
    <div class="state-msg">
      <div class="spinner"></div>
      <p>Loading settings...</p>
    </div>
  {:else}
    {#if feedback}
      <div class="feedback" class:error={feedbackType === 'error'} class:success={feedbackType === 'success'}>
        {feedback}
      </div>
    {/if}

    <section>
      <h2>LLM Backend</h2>

      <label class="field-label">Active Backend</label>
      <select bind:value={activeLlm}>
        <option value="free">Free (Gemini Flash)</option>
        <option value="claude">Claude</option>
        <option value="openai">OpenAI</option>
        <option value="local">Local (Ollama)</option>
      </select>

      {#if activeLlm === 'free'}
        <p class="field-note">No setup required - using built-in free tier API.</p>
      {/if}

      {#if activeLlm === 'claude'}
        <label class="field-label">Claude API Key</label>
        <div class="input-row">
          {#if showClaudeKey}
            <input type="text" bind:value={claudeApiKey} placeholder="sk-ant-..." />
          {:else}
            <input type="password" bind:value={claudeApiKey} placeholder="sk-ant-..." />
          {/if}
          <button class="toggle-btn" on:click={() => showClaudeKey = !showClaudeKey}>
            {showClaudeKey ? 'Hide' : 'Show'}
          </button>
        </div>
        <label class="field-label">Model</label>
        <input type="text" bind:value={claudeModel} placeholder="claude-sonnet-4-20250514" />
      {/if}

      {#if activeLlm === 'openai'}
        <label class="field-label">OpenAI API Key</label>
        <div class="input-row">
          {#if showOpenaiKey}
            <input type="text" bind:value={openaiApiKey} placeholder="sk-..." />
          {:else}
            <input type="password" bind:value={openaiApiKey} placeholder="sk-..." />
          {/if}
          <button class="toggle-btn" on:click={() => showOpenaiKey = !showOpenaiKey}>
            {showOpenaiKey ? 'Hide' : 'Show'}
          </button>
        </div>
        <label class="field-label">Model</label>
        <input type="text" bind:value={openaiModel} placeholder="gpt-4o" />
      {/if}

      {#if activeLlm === 'local'}
        <label class="field-label">Ollama Endpoint</label>
        <input type="text" bind:value={ollamaEndpoint} placeholder="http://localhost:11434" />
        <label class="field-label">Model</label>
        <div class="input-row">
          <input type="text" bind:value={ollamaModel} placeholder="llama3" />
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
    </section>

    <section>
      <h2>Strava Connection</h2>

      <div class="status-row">
        <span class="status-badge" class:connected={stravaConnected}>
          {stravaConnected ? 'Connected' : 'Not Connected'}
        </span>
      </div>

      <label class="field-label">Client ID</label>
      <input type="text" bind:value={stravaClientId} placeholder="Your Strava Client ID" />

      <label class="field-label">Client Secret</label>
      <div class="input-row">
        {#if showStravaSecret}
          <input type="text" bind:value={stravaClientSecret} placeholder="Your Strava Client Secret" />
        {:else}
          <input type="password" bind:value={stravaClientSecret} placeholder="Your Strava Client Secret" />
        {/if}
        <button class="toggle-btn" on:click={() => showStravaSecret = !showStravaSecret}>
          {showStravaSecret ? 'Hide' : 'Show'}
        </button>
      </div>

      <div class="strava-actions">
        {#if stravaConnected}
          <button class="btn btn-danger" on:click={disconnectStrava}>Disconnect</button>
        {:else}
          <button
            class="btn btn-primary"
            on:click={connectStrava}
            disabled={connectingStrava || !stravaClientId || !stravaClientSecret}
          >
            {connectingStrava ? 'Connecting...' : 'Connect Strava'}
          </button>
        {/if}
      </div>
    </section>

    <div class="save-area">
      <button class="btn btn-primary save-btn" on:click={save} disabled={saving}>
        {saving ? 'Saving...' : 'Save Settings'}
      </button>
    </div>
  {/if}
</div>

<style>
  .settings {
    flex: 1;
    overflow-y: auto;
    padding: 24px 24px;
    max-width: 700px;
  }

  .state-msg {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 40vh;
    opacity: 0.7;
    text-align: center;
    gap: 8px;
  }

  .spinner {
    width: 28px;
    height: 28px;
    border: 3px solid rgba(255, 255, 255, 0.15);
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .feedback {
    padding: 10px 16px;
    border-radius: 8px;
    font-size: 0.9rem;
    text-align: center;
    margin-bottom: 16px;
    animation: slideDown 0.3s ease;
  }

  .feedback.success {
    background: rgba(34, 197, 94, 0.15);
    color: #22c55e;
    border: 1px solid rgba(34, 197, 94, 0.3);
  }

  .feedback.error {
    background: rgba(220, 53, 69, 0.15);
    color: #f87171;
    border: 1px solid rgba(220, 53, 69, 0.3);
  }

  @keyframes slideDown {
    from { opacity: 0; transform: translateY(-10px); }
    to { opacity: 1; transform: translateY(0); }
  }

  section {
    margin-bottom: 28px;
    padding-bottom: 24px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  h2 {
    font-size: 1.1rem;
    font-weight: 600;
    color: #e2e8f0;
    margin: 0 0 16px;
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

  .status-row {
    margin-bottom: 12px;
  }

  .status-badge {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 20px;
    font-size: 0.8rem;
    font-weight: 600;
    background: rgba(148, 163, 184, 0.15);
    color: #94a3b8;
    border: 1px solid rgba(148, 163, 184, 0.3);
  }

  .status-badge.connected {
    background: rgba(34, 197, 94, 0.15);
    color: #22c55e;
    border-color: rgba(34, 197, 94, 0.3);
  }

  .strava-actions {
    margin-top: 16px;
  }

  .btn {
    padding: 10px 20px;
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

  .btn-danger {
    background: rgba(220, 53, 69, 0.15);
    color: #f87171;
    border: 1px solid rgba(220, 53, 69, 0.3);
  }

  .btn-danger:hover:not(:disabled) {
    background: rgba(220, 53, 69, 0.25);
  }

  .save-area {
    padding-top: 8px;
  }

  .save-btn {
    width: 100%;
    padding: 12px;
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
</style>
