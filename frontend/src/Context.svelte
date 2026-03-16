<script lang="ts">
  import { onMount } from 'svelte'
  import {
    GetProfileData,
    SaveProfileData,
    GetPinnedInsights,
    DeletePinnedInsight,
    GetRecentActivities
  } from '../wailsjs/go/main/App.js'

  let age = 0
  let maxHR = 0
  let thresholdPaceSecs = 0
  let weeklyMileageTarget = 0
  let raceGoals = ''
  let injuryHistory = ''
  let profileLoaded = false

  let insights: Array<{id: number, content: string, sourceSessionId: string, createdAt: string}> = []

  interface Activity {
    name: string
    activityType: string
    startDate: string
    distance: number
    durationSecs: number
    avgPaceSecs: number
    avgHR: number
  }

  let activities: Activity[] = []
  let loading = true
  let saving = false
  let feedback = ''
  let feedbackType: 'success' | 'error' = 'success'
  let feedbackTimer: ReturnType<typeof setTimeout> | null = null

  function showFeedback(msg: string, type: 'success' | 'error') {
    feedback = msg
    feedbackType = type
    if (feedbackTimer) clearTimeout(feedbackTimer)
    feedbackTimer = setTimeout(() => { feedback = '' }, 3000)
  }

  onMount(async () => {
    try {
      const [profile, insightList, activityList] = await Promise.all([
        GetProfileData(),
        GetPinnedInsights(),
        GetRecentActivities(10)
      ])

      if (profile) {
        age = profile.age || 0
        maxHR = profile.maxHR || 0
        thresholdPaceSecs = profile.thresholdPaceSecs || 0
        weeklyMileageTarget = profile.weeklyMileageTarget || 0
        raceGoals = profile.raceGoals || ''
        injuryHistory = profile.injuryHistory || ''
        profileLoaded = true
      }

      insights = insightList || []
      activities = activityList || []
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to load context data', 'error')
    } finally {
      loading = false
    }
  })

  async function saveProfile() {
    saving = true
    try {
      await SaveProfileData({
        age,
        maxHR,
        thresholdPaceSecs,
        weeklyMileageTarget,
        raceGoals,
        injuryHistory
      })
      profileLoaded = true
      showFeedback('Profile saved', 'success')
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to save profile', 'error')
    } finally {
      saving = false
    }
  }

  async function deleteInsight(id: number) {
    try {
      await DeletePinnedInsight(id)
      insights = insights.filter(i => i.id !== id)
      showFeedback('Insight removed', 'success')
    } catch (e: any) {
      showFeedback(e?.message || 'Failed to delete insight', 'error')
    }
  }

  function formatDuration(secs: number): string {
    const h = Math.floor(secs / 3600)
    const m = Math.floor((secs % 3600) / 60)
    const s = secs % 60
    const mm = String(m).padStart(2, '0')
    const ss = String(s).padStart(2, '0')
    if (h > 0) return `${h}:${mm}:${ss}`
    return `${m}:${ss}`
  }

  function formatPace(secs: number): string {
    if (secs <= 0) return '-'
    const m = Math.floor(secs / 60)
    const s = secs % 60
    return `${m}:${String(s).padStart(2, '0')}/km`
  }

  function formatDistance(km: number): string {
    return `${km.toFixed(1)} km`
  }

  function formatHR(hr: number): string {
    if (hr <= 0) return '-'
    return `${hr} bpm`
  }
</script>

<div class="context">
  {#if loading}
    <div class="state-msg">
      <div class="spinner"></div>
      <p>Loading context...</p>
    </div>
  {:else}
    {#if feedback}
      <div class="feedback" class:error={feedbackType === 'error'} class:success={feedbackType === 'success'}>
        {feedback}
      </div>
    {/if}

    <section>
      <h2>Athlete Profile</h2>
      <div class="form-grid">
        <div class="field">
          <label class="field-label">Age</label>
          <input type="number" bind:value={age} placeholder="30" min="1" max="120" />
        </div>
        <div class="field">
          <label class="field-label">Max Heart Rate</label>
          <input type="number" bind:value={maxHR} placeholder="185" min="100" max="220" />
        </div>
        <div class="field">
          <label class="field-label">Threshold Pace (sec/km)</label>
          <input type="number" bind:value={thresholdPaceSecs} placeholder="300" min="1" />
        </div>
        <div class="field">
          <label class="field-label">Weekly Mileage Target (km)</label>
          <input type="number" bind:value={weeklyMileageTarget} placeholder="50" step="0.1" min="0" />
        </div>
        <div class="field full-width">
          <label class="field-label">Race Goals</label>
          <textarea bind:value={raceGoals} placeholder="e.g. Sub-3:30 marathon in October" rows="2"></textarea>
        </div>
        <div class="field full-width">
          <label class="field-label">Injury History</label>
          <textarea bind:value={injuryHistory} placeholder="e.g. IT band issues in 2024, fully recovered" rows="2"></textarea>
        </div>
      </div>
      <button class="btn btn-primary" on:click={saveProfile} disabled={saving}>
        {saving ? 'Saving...' : 'Save Profile'}
      </button>
    </section>

    <section>
      <h2>Pinned Insights</h2>
      {#if insights.length === 0}
        <p class="empty-text">No pinned insights yet. Pin insights from chat to build your coaching context.</p>
      {:else}
        <div class="insights-list">
          {#each insights as insight}
            <div class="insight-item">
              <p class="insight-content">{insight.content}</p>
              <div class="insight-meta">
                <span class="insight-date">{new Date(insight.createdAt).toLocaleDateString()}</span>
                <button class="delete-btn" on:click={() => deleteInsight(insight.id)} title="Remove insight">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M18 6L6 18"></path>
                    <path d="M6 6l12 12"></path>
                  </svg>
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>

    <section>
      <h2>Training Summary</h2>
      {#if activities.length === 0}
        <p class="empty-text">No activities yet. Sync Strava or import a FIT file to see your training here.</p>
      {:else}
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Name</th>
                <th>Type</th>
                <th class="num">Distance</th>
                <th class="num">Duration</th>
                <th class="num">Pace</th>
                <th class="num">HR</th>
              </tr>
            </thead>
            <tbody>
              {#each activities as a}
                <tr>
                  <td class="date">{new Date(a.startDate).toLocaleDateString()}</td>
                  <td class="name">{a.name}</td>
                  <td class="type">{a.activityType}</td>
                  <td class="num">{formatDistance(a.distance)}</td>
                  <td class="num">{formatDuration(a.durationSecs)}</td>
                  <td class="num">{formatPace(a.avgPaceSecs)}</td>
                  <td class="num">{formatHR(a.avgHR)}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </section>
  {/if}
</div>

<style>
  .context {
    flex: 1;
    overflow-y: auto;
    padding: 24px 24px;
    max-width: 900px;
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

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px 20px;
    margin-bottom: 16px;
  }

  .field.full-width {
    grid-column: 1 / -1;
  }

  .field-label {
    display: block;
    font-size: 0.8rem;
    color: #94a3b8;
    margin-bottom: 6px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 600;
  }

  input[type="number"],
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
    box-sizing: border-box;
  }

  input:focus,
  textarea:focus {
    border-color: #3b82f6;
  }

  textarea {
    resize: vertical;
    min-height: 60px;
  }

  textarea::placeholder,
  input::placeholder {
    color: rgba(255, 255, 255, 0.3);
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

  .empty-text {
    color: #64748b;
    font-size: 0.9rem;
    line-height: 1.5;
  }

  .insights-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .insight-item {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    padding: 12px 14px;
  }

  .insight-content {
    margin: 0 0 8px;
    font-size: 0.9rem;
    color: #e2e8f0;
    line-height: 1.5;
    white-space: pre-wrap;
  }

  .insight-meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .insight-date {
    font-size: 0.75rem;
    color: #64748b;
  }

  .delete-btn {
    background: none;
    border: none;
    color: #64748b;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    transition: color 0.2s, background 0.2s;
  }

  .delete-btn:hover {
    color: #f87171;
    background: rgba(248, 113, 113, 0.1);
  }

  .table-wrap {
    overflow-x: auto;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }

  thead th {
    text-align: left;
    padding: 8px 10px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.15);
    color: #94a3b8;
    font-weight: 600;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
  }

  th.num, td.num {
    text-align: right;
  }

  tbody tr {
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  }

  td {
    padding: 10px 10px;
    white-space: nowrap;
    color: #e2e8f0;
  }

  td.date {
    color: #94a3b8;
    font-size: 0.85rem;
  }

  td.name {
    font-weight: 500;
    white-space: normal;
    max-width: 200px;
  }

  td.type {
    color: #94a3b8;
  }
</style>
