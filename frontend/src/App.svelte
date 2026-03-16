<script lang="ts">
  import { SendMessage, SaveInsight, IsFirstRun } from '../wailsjs/go/main/App.js'
  import { afterUpdate, onMount } from 'svelte'
  import Dashboard from './Dashboard.svelte'
  import Settings from './Settings.svelte'
  import Onboarding from './Onboarding.svelte'

  type Tab = 'chat' | 'dashboard' | 'settings'
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

  $: canSend = input.trim().length > 0 && !loading

  onMount(async () => {
    try {
      showOnboarding = await IsFirstRun()
    } catch (e) {
      showOnboarding = false
    }
    onboardingChecked = true
  })

  afterUpdate(() => {
    if (chatContainer) {
      chatContainer.scrollTop = chatContainer.scrollHeight
    }
  })

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

    try {
      const response = await SendMessage(text)
      messages = [...messages, { role: 'assistant', content: response }]
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

  function renderMarkdown(text: string): string {
    let html = text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')

    html = html.replace(/```(\w*)\n([\s\S]*?)```/g, (_m, _lang, code) => {
      return `<pre><code>${code.replace(/\n$/, '')}</code></pre>`
    })

    html = html.replace(/`([^`]+)`/g, '<code>$1</code>')

    html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')

    html = html.replace(/\*(.+?)\*/g, '<em>$1</em>')

    html = html.replace(/^[-*] (.+)$/gm, '<li>$1</li>')
    html = html.replace(/((?:<li>.*<\/li>\n?)+)/g, '<ul>$1</ul>')

    html = html.replace(/^\d+\. (.+)$/gm, '<li>$1</li>')

    html = html.replace(/\n\n/g, '</p><p>')
    html = `<p>${html}</p>`
    html = html.replace(/<p>\s*<\/p>/g, '')

    html = html.replace(/(?<!<\/li>)\n(?!<)/g, '<br>')

    return html
  }
</script>

<main class="app-shell">
  <nav class="tab-bar">
    <button
      class="tab"
      class:active={activeTab === 'chat'}
      on:click={() => activeTab = 'chat'}
    >Chat</button>
    <button
      class="tab"
      class:active={activeTab === 'dashboard'}
      on:click={() => activeTab = 'dashboard'}
    >Dashboard</button>
    <button
      class="tab"
      class:active={activeTab === 'settings'}
      on:click={() => activeTab = 'settings'}
    >Settings</button>
  </nav>

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

      <div class="input-area">
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
  {:else if activeTab === 'settings'}
    <Settings />
  {/if}
</main>

{#if onboardingChecked && showOnboarding}
  <Onboarding on:complete={() => { showOnboarding = false; activeTab = 'chat' }} />
{/if}

<style>
  .app-shell {
    display: flex;
    flex-direction: column;
    height: 100vh;
    max-width: 800px;
    margin: 0 auto;
    position: relative;
  }

  .tab-bar {
    display: flex;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    background: rgba(27, 38, 54, 0.95);
    flex-shrink: 0;
  }

  .tab {
    flex: 1;
    padding: 10px 16px;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: #94a3b8;
    font-size: 0.9rem;
    font-weight: 500;
    cursor: pointer;
    transition: color 0.2s, border-color 0.2s;
  }

  .tab:hover {
    color: #e2e8f0;
  }

  .tab.active {
    color: #3b82f6;
    border-bottom-color: #3b82f6;
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
    padding: 16px;
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
    max-width: 75%;
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
    padding: 12px 16px;
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
</style>
