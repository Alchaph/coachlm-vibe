<script lang="ts">
  import { onMount } from 'svelte'
  import {
    GetSettingsData,
    SaveSettingsData,
    GetStravaAuthStatus,
    StartStravaAuth,
    DisconnectStrava,
    GetOllamaModels,
    GetStravaCredentialsAvailable,
    ConnectS3,
    ConnectGoogleDrive,
    DisconnectCloud,
    SyncNow,
    GetSyncStatus,
    ResetApp
  } from '../wailsjs/go/main/App.js'
  import { cloudsync } from '../wailsjs/go/models'

  let ollamaEndpoint = 'http://localhost:11434'
  let ollamaModel = ''

  let stravaConnected = false
  let stravaCredentialsAvailable = false
  let loading = true
  let saving = false
  let connectingStrava = false
  let feedback = ''
  let feedbackType: 'success' | 'error' = 'success'
  let feedbackTimer: ReturnType<typeof setTimeout> | null = null

  let ollamaModels: string[] = []
  let fetchingModels = false
  let modelFetchError = ''

  let syncStatus: cloudsync.SyncStatus | null = null
  let cloudProvider = 'Google Drive'
  let s3Endpoint = ''
  let s3Bucket = ''
  let s3AccessKey = ''
  let s3SecretKey = ''
  let connectingCloud = false
  let syncingCloud = false

  let resetting = false

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

  async function loadSettings() {
    try {
      const [settings, status, credsAvailable, sync] = await Promise.all([
        GetSettingsData(),
        GetStravaAuthStatus(),
        GetStravaCredentialsAvailable(),
        GetSyncStatus()
      ])

      if (settings) {
        ollamaEndpoint = settings.ollamaEndpoint || 'http://localhost:11434'
        ollamaModel = settings.ollamaModel || ''
      }

      if (status) {
        stravaConnected = !!status.connected
      }

      stravaCredentialsAvailable = !!credsAvailable
      syncStatus = sync
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to load settings', 'error')
    }
  }

  onMount(async () => {
    try {
      const [settings, status, credsAvailable, sync] = await Promise.all([
        GetSettingsData(),
        GetStravaAuthStatus(),
        GetStravaCredentialsAvailable(),
        GetSyncStatus()
      ])

      if (settings) {
        ollamaEndpoint = settings.ollamaEndpoint || 'http://localhost:11434'
        ollamaModel = settings.ollamaModel || ''
      }

      if (status) {
        stravaConnected = !!status.connected
      }

      stravaCredentialsAvailable = !!credsAvailable
      syncStatus = sync
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to load settings', 'error')
    } finally {
      loading = false
    }
  })

  async function save() {
    saving = true
    try {
      const currentSettings = await GetSettingsData()
      await SaveSettingsData({
        ollamaEndpoint,
        ollamaModel,
        customSystemPrompt: currentSettings?.customSystemPrompt || ''
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
      const currentSettings = await GetSettingsData()
      await SaveSettingsData({
        ollamaEndpoint,
        ollamaModel,
        customSystemPrompt: currentSettings?.customSystemPrompt || ''
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

  async function connectS3() {
    if (!s3Endpoint || !s3Bucket || !s3AccessKey || !s3SecretKey) {
      showFeedback('Please fill in all S3 fields', 'error')
      return
    }
    connectingCloud = true
    try {
      await ConnectS3(s3Endpoint, s3Bucket, s3AccessKey, s3SecretKey)
      syncStatus = await GetSyncStatus()
      showFeedback('Connected to S3 successfully', 'success')
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to connect to S3', 'error')
    } finally {
      connectingCloud = false
    }
  }

  async function connectGoogleDrive() {
    connectingCloud = true
    try {
      await ConnectGoogleDrive()
      syncStatus = await GetSyncStatus()
      showFeedback('Connected to Google Drive successfully', 'success')
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to connect to Google Drive', 'error')
    } finally {
      connectingCloud = false
    }
  }

  async function disconnectCloud() {
    if (!confirm('Disconnect Cloud Sync? Your local data will remain.')) return
    try {
      await DisconnectCloud()
      syncStatus = await GetSyncStatus()
      showFeedback('Cloud Sync disconnected', 'success')
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to disconnect Cloud Sync', 'error')
    }
  }

  async function syncNow() {
    syncingCloud = true
    try {
      await SyncNow()
      syncStatus = await GetSyncStatus()
      showFeedback('Sync completed successfully', 'success')
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to sync', 'error')
    } finally {
      syncingCloud = false
    }
  }

  async function resetApp() {
    if (!confirm('Reset CoachLM? This will erase ALL data (settings, activities, chat history, Strava connection) and return to the setup wizard. This cannot be undone.')) return
    resetting = true
    try {
      await ResetApp()
      window.location.reload()
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to reset', 'error')
      resetting = false
    }
  }
</script>

<div class="settings">
  <div class="settings-inner">
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
      <h2>AI Model</h2>
      <p class="ollama-label">Powered by Ollama (local)</p>

      <label class="field-label" for="ollama-endpoint">Ollama Endpoint</label>
      <input id="ollama-endpoint" type="text" bind:value={ollamaEndpoint} placeholder="http://localhost:11434" />
      <label class="field-label" for="ollama-model">Model</label>
      <div class="input-row">
        <input id="ollama-model" type="text" bind:value={ollamaModel} placeholder="llama3" />
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
    </section>

    <section>
      <h2>Strava Connection</h2>

      <div class="status-row">
        <span class="status-badge" class:connected={stravaConnected}>
          {stravaConnected ? 'Connected' : 'Not Connected'}
        </span>
      </div>

      <div class="strava-actions">
        {#if stravaConnected}
          <button class="btn btn-danger" on:click={disconnectStrava}>Disconnect</button>
        {:else if stravaCredentialsAvailable}
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
    </section>

    <section>
      <h2>Cloud Sync</h2>
      
      <div class="status-row">
        <span class="status-badge" class:connected={syncStatus?.enabled}>
          {syncStatus?.enabled ? 'Connected' : 'Not Connected'}
        </span>
      </div>

      {#if syncStatus?.enabled}
        <div class="cloud-connected-info">
          <p><strong>Provider:</strong> {syncStatus.provider}</p>
          <p><strong>Last Synced:</strong> {syncStatus.lastSyncedAt ? new Date(syncStatus.lastSyncedAt).toLocaleString() : 'Never'}</p>
          {#if syncStatus.lastError}
            <p class="cloud-error"><strong>Error:</strong> {syncStatus.lastError}</p>
          {/if}
        </div>
        <div class="cloud-actions">
          <button class="btn btn-primary" on:click={syncNow} disabled={syncingCloud || syncStatus.syncing}>
            {syncingCloud || syncStatus.syncing ? 'Syncing...' : 'Sync Now'}
          </button>
          <button class="btn btn-danger" on:click={disconnectCloud}>Disconnect</button>
        </div>
      {:else}
        <label class="field-label" for="cloud-provider">Provider</label>
        <select id="cloud-provider" bind:value={cloudProvider}>
          <option value="Google Drive">Google Drive</option>
          <option value="S3-Compatible">S3-Compatible</option>
        </select>

        {#if cloudProvider === 'S3-Compatible'}
          <div class="s3-fields">
            <label class="field-label" for="s3-endpoint">Endpoint URL</label>
            <input id="s3-endpoint" type="text" bind:value={s3Endpoint} placeholder="https://s3.us-east-1.amazonaws.com" />
            
            <label class="field-label" for="s3-bucket">Bucket Name</label>
            <input id="s3-bucket" type="text" bind:value={s3Bucket} placeholder="my-coachlm-bucket" />
            
            <label class="field-label" for="s3-access-key">Access Key</label>
            <input id="s3-access-key" type="text" bind:value={s3AccessKey} placeholder="AKIA..." />
            
            <label class="field-label" for="s3-secret-key">Secret Key</label>
            <input id="s3-secret-key" type="password" bind:value={s3SecretKey} placeholder="Secret Key" />
            
            <div class="cloud-actions">
              <button class="btn btn-primary" on:click={connectS3} disabled={connectingCloud}>
                {connectingCloud ? 'Connecting...' : 'Connect S3'}
              </button>
            </div>
          </div>
        {:else if cloudProvider === 'Google Drive'}
          <div class="cloud-actions">
            <button class="btn btn-primary" on:click={connectGoogleDrive} disabled={connectingCloud}>
              {connectingCloud ? 'Connecting...' : 'Connect Google Drive'}
            </button>
          </div>
        {/if}
      {/if}
    </section>

    <div class="save-area">
      <button class="btn btn-primary save-btn" on:click={save} disabled={saving}>
        {saving ? 'Saving...' : 'Save Settings'}
      </button>
    </div>

    <section class="danger-zone">
      <h2>Danger Zone</h2>
      <p class="danger-desc">Erase all data and return to the setup wizard.</p>
      <button class="btn btn-danger" on:click={resetApp} disabled={resetting}>
        {resetting ? 'Resetting...' : 'Reset App'}
      </button>
    </section>
  {/if}
  </div>
</div>

<style>
  .settings {
    flex: 1;
    overflow-y: auto;
  }

  .settings-inner {
    max-width: 700px;
    padding: 24px 24px;
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

  .ollama-label {
    color: #22c55e;
    font-size: 1rem;
    margin-bottom: 16px;
    font-weight: 500;
  }

  select,
  input[type="text"],
  input[type="password"],
  textarea {
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
  input:focus,
  textarea:focus {
    border-color: #3b82f6;
  }

  textarea {
    resize: vertical;
    min-height: 80px;
    box-sizing: border-box;
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

  .strava-unavailable {
    color: #94a3b8;
    font-style: italic;
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

  .cloud-connected-info {
    margin-bottom: 16px;
    font-size: 0.9rem;
    color: #e2e8f0;
  }

  .cloud-connected-info p {
    margin: 4px 0;
  }

  .cloud-error {
    color: #f87171;
  }

  .cloud-actions {
    margin-top: 16px;
    display: flex;
    gap: 12px;
  }

  .s3-fields {
    margin-top: 12px;
  }

  .danger-zone {
    margin-top: 32px;
    padding-top: 24px;
    border-top: 1px solid rgba(220, 53, 69, 0.3);
    border-bottom: none;
  }

  .danger-zone h2 {
    color: #f87171;
  }

  .danger-desc {
    font-size: 0.85rem;
    color: #94a3b8;
    margin-bottom: 12px;
  }
</style>
