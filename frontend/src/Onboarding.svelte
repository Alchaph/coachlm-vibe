<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { SaveSettingsData, StartStravaAuth } from '../wailsjs/go/main/App.js'

  const dispatch = createEventDispatcher()

  let step = 1
  let activeLlm = 'local'
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

  function next() {
    if (step < 4) step++
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
      next()
    } catch (e: any) {
      error = e?.message || 'Failed to connect Strava'
    } finally {
      connectingStrava = false
    }
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
      {#each [1, 2, 3, 4] as s}
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
          <label class="field-label">Backend</label>
          <select bind:value={activeLlm}>
            <option value="claude">Claude</option>
            <option value="openai">OpenAI</option>
            <option value="local">Local (Ollama)</option>
          </select>

          {#if activeLlm === 'claude'}
            <label class="field-label">Claude API Key</label>
            <div class="input-row">
              {#if showApiKey}
                <input type="text" bind:value={claudeApiKey} placeholder="sk-ant-..." />
              {:else}
                <input type="password" bind:value={claudeApiKey} placeholder="sk-ant-..." />
              {/if}
              <button class="toggle-btn" on:click={() => showApiKey = !showApiKey}>
                {showApiKey ? 'Hide' : 'Show'}
              </button>
            </div>
            <label class="field-label">Model</label>
            <input type="text" bind:value={claudeModel} placeholder="claude-sonnet-4-20250514" />
          {/if}

          {#if activeLlm === 'openai'}
            <label class="field-label">OpenAI API Key</label>
            <div class="input-row">
              {#if showApiKey}
                <input type="text" bind:value={openaiApiKey} placeholder="sk-..." />
              {:else}
                <input type="password" bind:value={openaiApiKey} placeholder="sk-..." />
              {/if}
              <button class="toggle-btn" on:click={() => showApiKey = !showApiKey}>
                {showApiKey ? 'Hide' : 'Show'}
              </button>
            </div>
            <label class="field-label">Model</label>
            <input type="text" bind:value={openaiModel} placeholder="gpt-4o" />
          {/if}

          {#if activeLlm === 'local'}
            <label class="field-label">Ollama Endpoint</label>
            <input type="text" bind:value={ollamaEndpoint} placeholder="http://localhost:11434" />
            <label class="field-label">Model</label>
            <input type="text" bind:value={ollamaModel} placeholder="llama3" />
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
          <label class="field-label">Client ID</label>
          <input type="text" bind:value={stravaClientId} placeholder="Your Strava Client ID" />

          <label class="field-label">Client Secret</label>
          <input type="password" bind:value={stravaClientSecret} placeholder="Your Strava Client Secret" />
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
        <h1>You're All Set!</h1>
        <p class="subtitle">Start chatting with your AI running coach.</p>
        {#if stravaConnected}
          <p class="connected-note">Strava connected successfully.</p>
        {/if}
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
</style>
