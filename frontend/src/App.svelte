<script lang="ts">
  import { SendMessage, SaveInsight, IsFirstRun } from '../wailsjs/go/main/App.js'
  import { onMount } from 'svelte'
  import { marked } from 'marked'
  import Dashboard from './Dashboard.svelte'
  import Settings from './Settings.svelte'
  import Context from './Context.svelte'
  import TrainingPlan from './lib/TrainingPlan.svelte'
  import Onboarding from './Onboarding.svelte'

  type Tab = 'chat' | 'dashboard' | 'context' | 'plan' | 'settings'
  let activeTab: Tab = 'chat'

  let showOnboarding = false
  let onboardingChecked = false

  interface ChatMessage {
    role: 'user' | 'assistant'
    content: string
  }

  let messages: ChatMessage[] = []
  let input = ''
  let loading = false
  let error = ''
  let errorTimer: ReturnType<typeof setTimeout> | null = null
  let chatContainer: HTMLElement
  let pinnedIndices: Set<number> = new Set()
  let pinFeedback: Record<number, string> = {}

  let showPlanInput = false
  let planRace = ''
  let planDate = ''
  let planTime = ''
  let planError = ''

  $: canSend = input.trim().length > 0 && !loading

  onMount(async () => {
    try {
      showOnboarding = await IsFirstRun()
    } catch (e) {
      showOnboarding = false
    }
    onboardingChecked = true
  })

  function scrollToBottom() {
    if (chatContainer) {
      chatContainer.scrollTop = chatContainer.scrollHeight
    }
  }

  function showError(msg: string) {
    error = msg
    if (errorTimer) clearTimeout(errorTimer)
    errorTimer = setTimeout(() => { error = '' }, 5000)
  }

  async function send() {
    const text = input.trim()
    if (!text || loading) return

    messages = [...messages, { role: 'user', content: text }]
    input = ''
    loading = true
    scrollToBottom()

    try {
      const response = await SendMessage(text)
      messages = [...messages, { role: 'assistant', content: response }]
      scrollToBottom()
    } catch (e: any) {
      showError(e?.message || String(e) || 'Failed to get response')
    } finally {
      loading = false
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (canSend) send()
    }
  }

  async function sendPlanRequest() {
    if (!planRace.trim() && !planDate.trim() && !planTime.trim()) {
      planError = 'Enter at least a race type, date, or target time'
      return
    }
    planError = ''

    const prompt = `Generate a structured training plan for me.
Race type: ${planRace.trim() || 'Not specified'}
Target date: ${planDate.trim() || 'Not specified'}
Target time: ${planTime.trim() || 'Not specified'}

Please create a weekly breakdown with key workouts, easy days, and recovery. Use my profile data and recent training to set appropriate paces.`

    showPlanInput = false
    planRace = ''
    planDate = ''
    planTime = ''

    messages = [...messages, { role: 'user', content: prompt }]
    loading = true
    scrollToBottom()

    try {
      const response = await SendMessage(prompt)
      messages = [...messages, { role: 'assistant', content: response }]
      scrollToBottom()
    } catch (e: any) {
      showError(e?.message || String(e) || 'Failed to get response')
    } finally {
      loading = false
    }
  }

  async function pinInsight(index: number, content: string) {
    if (pinnedIndices.has(index)) {
      pinFeedback = { ...pinFeedback, [index]: 'Already pinned' }
      setTimeout(() => { const f = { ...pinFeedback }; delete f[index]; pinFeedback = f }, 2000)
      return
    }
    try {
      await SaveInsight(content)
      pinnedIndices.add(index)
      pinnedIndices = pinnedIndices
      pinFeedback = { ...pinFeedback, [index]: 'Insight saved!' }
    } catch (e: any) {
      const msg = e?.message || String(e) || 'Failed to save'
      if (msg.toLowerCase().includes('already')) {
        pinnedIndices.add(index)
        pinnedIndices = pinnedIndices
        pinFeedback = { ...pinFeedback, [index]: 'Already pinned' }
      } else {
        pinFeedback = { ...pinFeedback, [index]: 'Save failed' }
      }
    }
    setTimeout(() => { const f = { ...pinFeedback }; delete f[index]; pinFeedback = f }, 2000)
  }

  const renderer = {
    link({ href, title, tokens }: { href: string, title?: string | null, tokens: any[] }) {
      const text = (this as any).parser.parseInline(tokens)
      let out = `<a href="${href}" target="_blank" rel="noopener noreferrer"`
      if (title) {
        out += ` title="${title}"`
      }
      out += `>${text}</a>`
      return out
    }
  }

  marked.use({
    renderer,
    breaks: true,
    gfm: true
  })

  function renderMarkdown(text: string): string {
    return marked.parse(text) as string
  }
</script>

<main class="app-shell">
  <nav class="sidebar">
    <button
      class="nav-item"
      class:active={activeTab === 'chat'}
      on:click={() => activeTab = 'chat'}
      title="Chat"
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
      </svg>
      <span class="nav-label">Chat</span>
    </button>
    <button
      class="nav-item"
      class:active={activeTab === 'dashboard'}
      on:click={() => activeTab = 'dashboard'}
      title="Dashboard"
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="3" width="7" height="7"></rect>
        <rect x="14" y="3" width="7" height="7"></rect>
        <rect x="3" y="14" width="7" height="7"></rect>
        <rect x="14" y="14" width="7" height="7"></rect>
      </svg>
      <span class="nav-label">Dashboard</span>
    </button>
    <button
      class="nav-item"
      class:active={activeTab === 'context'}
      on:click={() => activeTab = 'context'}
      title="Context"
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 2L2 7l10 5 10-5-10-5z"></path>
        <path d="M2 17l10 5 10-5"></path>
        <path d="M2 12l10 5 10-5"></path>
      </svg>
      <span class="nav-label">Context</span>
    </button>
    <button
      class="nav-item"
      class:active={activeTab === 'plan'}
      on:click={() => activeTab = 'plan'}
      title="Training Plan"
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
        <line x1="16" y1="2" x2="16" y2="6"></line>
        <line x1="8" y1="2" x2="8" y2="6"></line>
        <line x1="3" y1="10" x2="21" y2="10"></line>
      </svg>
      <span class="nav-label">Plan</span>
    </button>
    <div class="nav-spacer"></div>
    <button
      class="nav-item"
      class:active={activeTab === 'settings'}
      on:click={() => activeTab = 'settings'}
      title="Settings"
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="3"></circle>
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
      </svg>
      <span class="nav-label">Settings</span>
    </button>
  </nav>

  <div class="content">
    {#if activeTab === 'chat'}
    <div class="chat-app">
      {#if error}
        <div class="error-banner" role="alert" on:click={() => error = ''} on:keydown={(e) => e.key === 'Enter' && (error = '')}>
          {error}
        </div>
      {/if}

      <div class="chat-messages" bind:this={chatContainer}>
        {#if messages.length === 0 && !loading}
          <div class="empty-state">
            <div class="empty-icon">🏃</div>
            <h2>CoachLM</h2>
            <p>Your AI running coach. Ask me anything about training, recovery, or race preparation.</p>
          </div>
        {/if}

        {#each messages as msg, i}
          <div class="message {msg.role}">
            <div class="message-bubble">
              {#if msg.role === 'assistant'}
                <div class="markdown">{@html renderMarkdown(msg.content)}</div>
                <div class="pin-row">
                  {#if pinFeedback[i]}
                    <span class="pin-feedback">{pinFeedback[i]}</span>
                  {/if}
                  <button
                    class="pin-btn"
                    class:pinned={pinnedIndices.has(i)}
                    on:click={() => pinInsight(i, msg.content)}
                    aria-label="Pin insight"
                    title={pinnedIndices.has(i) ? 'Already pinned' : 'Save as insight'}
                  >📌</button>
                </div>
              {:else}
                <div class="text">{msg.content}</div>
              {/if}
            </div>
          </div>
        {/each}

        {#if loading}
          <div class="message assistant">
            <div class="message-bubble loading-bubble">
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="dot"></span>
            </div>
          </div>
        {/if}
      </div>

      {#if showPlanInput}
      <div class="plan-input-panel">
        <div class="plan-header">
          <span>Training Plan Goal</span>
          <button class="plan-close" on:click={() => showPlanInput = false} aria-label="Close">&times;</button>
        </div>
        {#if planError}
          <div class="plan-error">{planError}</div>
        {/if}
        <div class="plan-fields">
          <label>
            <span>Race type</span>
            <input type="text" bind:value={planRace} placeholder="e.g., 5K, Half Marathon, Marathon" />
          </label>
          <label>
            <span>Target date</span>
            <input type="date" bind:value={planDate} />
          </label>
          <label>
            <span>Target time</span>
            <input type="text" bind:value={planTime} placeholder="e.g., 3:30:00, sub-20" />
          </label>
        </div>
        <button class="plan-submit" on:click={sendPlanRequest} disabled={loading}>
          Generate Plan
        </button>
      </div>
      {/if}

      <div class="input-area">
        <button
          class="plan-btn"
          on:click={() => { showPlanInput = !showPlanInput; planError = '' }}
          disabled={loading}
          aria-label="Generate Training Plan"
          title="Generate Training Plan"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
            <rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
          </svg>
        </button>
        <textarea
          bind:value={input}
          on:keydown={handleKeydown}
          placeholder="Ask your coach..."
          rows="1"
          disabled={loading}
        ></textarea>
        <button on:click={send} disabled={!canSend} class="send-btn" aria-label="Send message">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 2L11 13"></path>
            <path d="M22 2L15 22L11 13L2 9L22 2Z"></path>
          </svg>
        </button>
      </div>
    </div>
  {:else if activeTab === 'dashboard'}
    <Dashboard />
  {:else if activeTab === 'context'}
    <Context />
  {:else if activeTab === 'plan'}
    <TrainingPlan on:adjustchat={(e) => { input = e.detail; activeTab = 'chat' }} />
  {:else if activeTab === 'settings'}
    <Settings />
  {/if}
  </div>
</main>

{#if onboardingChecked && showOnboarding}
  <Onboarding on:complete={() => { showOnboarding = false; activeTab = 'chat' }} />
{/if}

<style>
  .app-shell {
    display: flex;
    flex-direction: row;
    height: 100vh;
    position: relative;
  }

  .sidebar {
    display: flex;
    flex-direction: column;
    width: 56px;
    min-width: 56px;
    background: rgba(15, 23, 36, 0.95);
    border-right: 1px solid rgba(255, 255, 255, 0.08);
    padding: 8px 0;
    flex-shrink: 0;
    overflow: hidden;
    transition: width 0.2s ease;
  }

  .sidebar:hover {
    width: 180px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 18px;
    background: none;
    border: none;
    color: #64748b;
    cursor: pointer;
    transition: color 0.15s, background 0.15s;
    white-space: nowrap;
    width: 100%;
    text-align: left;
  }

  .nav-item:hover {
    color: #e2e8f0;
    background: rgba(255, 255, 255, 0.05);
  }

  .nav-item.active {
    color: #3b82f6;
    background: rgba(59, 130, 246, 0.1);
  }

  .nav-item svg {
    flex-shrink: 0;
  }

  .nav-label {
    font-size: 0.85rem;
    font-weight: 500;
    opacity: 0;
    transition: opacity 0.15s ease;
  }

  .sidebar:hover .nav-label {
    opacity: 1;
  }

  .nav-spacer {
    flex: 1;
  }

  .content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
  }

  .chat-app {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow: hidden;
    position: relative;
  }

  .error-banner {
    position: absolute;
    top: 12px;
    left: 16px;
    right: 16px;
    background: #dc3545;
    color: white;
    padding: 10px 16px;
    border-radius: 8px;
    font-size: 0.9rem;
    z-index: 10;
    cursor: pointer;
    text-align: center;
    animation: slideDown 0.3s ease;
  }

  @keyframes slideDown {
    from { opacity: 0; transform: translateY(-10px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 16px 24px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    flex: 1;
    min-height: 60vh;
    opacity: 0.7;
    text-align: center;
  }

  .empty-icon {
    font-size: 3rem;
    margin-bottom: 12px;
  }

  .empty-state h2 {
    margin: 0 0 8px;
    font-size: 1.5rem;
    font-weight: 700;
  }

  .empty-state p {
    margin: 0;
    font-size: 0.95rem;
    max-width: 360px;
    line-height: 1.5;
  }

  .message {
    display: flex;
    width: 100%;
  }

  .message.user {
    justify-content: flex-end;
  }

  .message.assistant {
    justify-content: flex-start;
  }

  .message-bubble {
    max-width: 70%;
    padding: 10px 14px;
    border-radius: 16px;
    font-size: 0.95rem;
    line-height: 1.5;
    word-wrap: break-word;
    overflow-wrap: break-word;
  }

  .message.user .message-bubble {
    background: #3b82f6;
    color: white;
    border-bottom-right-radius: 4px;
  }

  .message.assistant .message-bubble {
    background: rgba(255, 255, 255, 0.1);
    color: #e2e8f0;
    border-bottom-left-radius: 4px;
  }

  .pin-row {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 6px;
    margin-top: 4px;
    min-height: 24px;
  }

  .pin-btn {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 0.8rem;
    padding: 2px 4px;
    border-radius: 4px;
    opacity: 0;
    transition: opacity 0.2s;
    line-height: 1;
  }

  .message-bubble:hover .pin-btn,
  .pin-btn.pinned {
    opacity: 0.7;
  }

  .pin-btn:hover {
    opacity: 1 !important;
    background: rgba(255, 255, 255, 0.1);
  }

  .pin-btn.pinned {
    opacity: 0.5;
    cursor: default;
  }

  .pin-feedback {
    font-size: 0.75rem;
    color: #94a3b8;
    animation: slideDown 0.3s ease;
  }

  .message-bubble .text {
    white-space: pre-wrap;
  }

  .message-bubble .markdown :global(p) {
    margin: 0 0 8px;
  }

  .message-bubble .markdown :global(p:last-child) {
    margin-bottom: 0;
  }

  .message-bubble .markdown :global(pre) {
    background: rgba(0, 0, 0, 0.3);
    border-radius: 6px;
    padding: 10px;
    overflow-x: auto;
    margin: 8px 0;
  }

  .message-bubble .markdown :global(code) {
    font-family: 'Courier New', monospace;
    font-size: 0.85em;
  }

  .message-bubble .markdown :global(p code),
  .message-bubble .markdown :global(li code) {
    background: rgba(0, 0, 0, 0.25);
    padding: 2px 5px;
    border-radius: 3px;
  }

  .message-bubble .markdown :global(ul) {
    margin: 4px 0;
    padding-left: 20px;
  }

  .message-bubble .markdown :global(li) {
    margin: 2px 0;
  }

  .message-bubble .markdown :global(strong) {
    font-weight: 700;
  }

  .message-bubble .markdown :global(em) {
    font-style: italic;
  }

  .message-bubble .markdown :global(h1),
  .message-bubble .markdown :global(h2),
  .message-bubble .markdown :global(h3),
  .message-bubble .markdown :global(h4) {
    margin: 12px 0 6px;
    font-weight: 700;
    line-height: 1.3;
  }

  .message-bubble .markdown :global(h1) { font-size: 1.4em; }
  .message-bubble .markdown :global(h2) { font-size: 1.25em; }
  .message-bubble .markdown :global(h3) { font-size: 1.1em; }
  .message-bubble .markdown :global(h4) { font-size: 1em; }

  .message-bubble .markdown :global(h1:first-child),
  .message-bubble .markdown :global(h2:first-child),
  .message-bubble .markdown :global(h3:first-child),
  .message-bubble .markdown :global(h4:first-child) {
    margin-top: 0;
  }

  .message-bubble .markdown :global(a) {
    color: #3b82f6;
    text-decoration: none;
  }

  .message-bubble .markdown :global(a:hover) {
    text-decoration: underline;
  }

  .message-bubble .markdown :global(blockquote) {
    border-left: 3px solid #3b82f6;
    margin: 8px 0;
    padding: 4px 12px;
    color: #94a3b8;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 0 4px 4px 0;
  }

  .message-bubble .markdown :global(blockquote p) {
    margin: 4px 0;
  }

  .message-bubble .markdown :global(ol) {
    margin: 4px 0;
    padding-left: 20px;
  }

  .message-bubble .markdown :global(hr) {
    border: none;
    border-top: 1px solid rgba(255, 255, 255, 0.15);
    margin: 12px 0;
  }

  .message-bubble .markdown :global(del) {
    text-decoration: line-through;
    opacity: 0.7;
  }

  .message-bubble .markdown :global(table) {
    border-collapse: collapse;
    width: 100%;
    margin: 8px 0;
    font-size: 0.9em;
  }

  .message-bubble .markdown :global(th),
  .message-bubble .markdown :global(td) {
    border: 1px solid rgba(255, 255, 255, 0.15);
    padding: 6px 10px;
    text-align: left;
  }

  .message-bubble .markdown :global(th) {
    background: rgba(255, 255, 255, 0.08);
    font-weight: 700;
  }

  .message-bubble .markdown :global(tr:nth-child(even)) {
    background: rgba(255, 255, 255, 0.03);
  }

  .message-bubble .markdown :global(img) {
    max-width: 100%;
    border-radius: 6px;
    margin: 8px 0;
  }

  .message-bubble .markdown :global(input[type="checkbox"]) {
    margin-right: 6px;
    pointer-events: none;
  }

  .loading-bubble {
    display: flex;
    gap: 4px;
    padding: 14px 18px;
  }

  .dot {
    width: 8px;
    height: 8px;
    background: #94a3b8;
    border-radius: 50%;
    animation: bounce 1.4s infinite ease-in-out both;
  }

  .dot:nth-child(1) { animation-delay: -0.32s; }
  .dot:nth-child(2) { animation-delay: -0.16s; }

  @keyframes bounce {
    0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; }
    40% { transform: scale(1); opacity: 1; }
  }

  .input-area {
    display: flex;
    align-items: flex-end;
    gap: 8px;
    padding: 12px 24px;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    background: rgba(27, 38, 54, 0.95);
  }

  textarea {
    flex: 1;
    resize: none;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.08);
    color: white;
    padding: 10px 14px;
    font-family: inherit;
    font-size: 0.95rem;
    line-height: 1.4;
    outline: none;
    min-height: 42px;
    max-height: 120px;
  }

  textarea::placeholder {
    color: rgba(255, 255, 255, 0.4);
  }

  textarea:focus {
    border-color: #3b82f6;
  }

  textarea:disabled {
    opacity: 0.5;
  }

  .send-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    border: none;
    border-radius: 50%;
    background: #3b82f6;
    color: white;
    cursor: pointer;
    flex-shrink: 0;
    transition: background 0.2s;
  }

  .send-btn:hover:not(:disabled) {
    background: #2563eb;
  }

  .send-btn:disabled {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.3);
    cursor: not-allowed;
  }

  .plan-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    border: none;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.08);
    color: #94a3b8;
    cursor: pointer;
    flex-shrink: 0;
    transition: background 0.2s, color 0.2s;
  }

  .plan-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.15);
    color: #e2e8f0;
  }

  .plan-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .plan-input-panel {
    background: rgba(27, 38, 54, 0.98);
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    padding: 12px 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .plan-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-weight: 600;
    color: #e2e8f0;
  }

  .plan-close {
    background: none;
    border: none;
    color: #94a3b8;
    font-size: 1.2rem;
    cursor: pointer;
    padding: 0 4px;
    line-height: 1;
  }

  .plan-close:hover {
    color: #e2e8f0;
  }

  .plan-fields {
    display: flex;
    flex-direction: row;
    gap: 12px;
    flex-wrap: wrap;
  }

  .plan-fields label {
    display: flex;
    flex-direction: column;
    gap: 4px;
    color: #94a3b8;
    font-size: 0.8rem;
    flex: 1;
    min-width: 120px;
  }

  .plan-fields input {
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.08);
    color: white;
    padding: 8px 12px;
    font-family: inherit;
    font-size: 0.95rem;
    outline: none;
  }

  .plan-fields input:focus {
    border-color: #3b82f6;
  }

  .plan-submit {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    border: none;
    border-radius: 8px;
    background: #3b82f6;
    color: white;
    cursor: pointer;
    padding: 8px;
    font-weight: 500;
    transition: background 0.2s;
  }

  .plan-submit:hover:not(:disabled) {
    background: #2563eb;
  }

  .plan-submit:disabled {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.3);
    cursor: not-allowed;
  }

  .plan-error {
    color: #f87171;
    font-size: 0.8rem;
    margin: 4px 0;
  }
</style>
