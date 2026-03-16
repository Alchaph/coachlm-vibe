<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { GetRecentActivities, GetStravaAuthStatus, SyncStravaActivities } from '../wailsjs/go/main/App.js'
  import { EventsOn } from '../wailsjs/runtime/runtime.js'

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
  let error = ''
  let stravaConnected = false
  let syncing = false
  let syncCurrent = 0
  let syncTotal = 0
  let syncResult = ''
  let syncResultTimer: ReturnType<typeof setTimeout> | null = null

  let unsubStart: (() => void) | null = null
  let unsubProgress: (() => void) | null = null
  let unsubComplete: (() => void) | null = null
  let unsubError: (() => void) | null = null

  async function loadActivities() {
    activities = await GetRecentActivities(20)
  }

  onMount(async () => {
    try {
      const [, status] = await Promise.all([
        loadActivities().catch(() => {}),
        GetStravaAuthStatus().catch(() => null)
      ])
      if (status) stravaConnected = !!status.connected
    } catch (e: any) {
      error = e?.message || String(e) || 'Failed to load activities'
    } finally {
      loading = false
    }

    unsubStart = EventsOn("strava:sync:start", () => {
      syncing = true
      syncCurrent = 0
      syncTotal = 0
      syncResult = ''
    })

    unsubProgress = EventsOn("strava:sync:progress", (data: any) => {
      syncCurrent = data?.current || 0
      syncTotal = data?.total || 0
    })

    unsubComplete = EventsOn("strava:sync:complete", async (data: any) => {
      syncing = false
      const saved = data?.saved || 0
      const total = data?.total || 0
      syncResult = saved > 0 ? `Synced ${saved} new activities (${total} total)` : `Up to date (${total} activities)`
      if (syncResultTimer) clearTimeout(syncResultTimer)
      syncResultTimer = setTimeout(() => { syncResult = '' }, 5000)
      try { await loadActivities() } catch (_) {}
    })

    unsubError = EventsOn("strava:sync:error", (msg: any) => {
      syncing = false
      syncResult = `Sync failed: ${msg}`
      if (syncResultTimer) clearTimeout(syncResultTimer)
      syncResultTimer = setTimeout(() => { syncResult = '' }, 5000)
    })
  })

  onDestroy(() => {
    if (unsubStart) unsubStart()
    if (unsubProgress) unsubProgress()
    if (unsubComplete) unsubComplete()
    if (unsubError) unsubError()
    if (syncResultTimer) clearTimeout(syncResultTimer)
  })

  async function startSync() {
    if (syncing) return
    syncing = true
    syncResult = ''
    try {
      await SyncStravaActivities()
    } catch (e: any) {
      syncing = false
      syncResult = `Sync failed: ${e?.message || String(e)}`
      if (syncResultTimer) clearTimeout(syncResultTimer)
      syncResultTimer = setTimeout(() => { syncResult = '' }, 5000)
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

<div class="dashboard">
  {#if stravaConnected}
    <div class="sync-bar">
      <button class="btn btn-sync" on:click={startSync} disabled={syncing}>
        {#if syncing && syncTotal > 0}
          Syncing {syncCurrent}/{syncTotal}...
        {:else if syncing}
          Starting sync...
        {:else}
          Sync Activities
        {/if}
      </button>
      {#if syncing && syncTotal > 0}
        <div class="progress-bar">
          <div class="progress-fill" style="width: {Math.round((syncCurrent / syncTotal) * 100)}%"></div>
        </div>
      {/if}
      {#if syncResult}
        <span class="sync-result">{syncResult}</span>
      {/if}
    </div>
  {/if}

  {#if loading}
    <div class="state-msg">
      <div class="spinner"></div>
      <p>Loading activities...</p>
    </div>
  {:else if error}
    <div class="state-msg error">
      <p>{error}</p>
    </div>
  {:else if activities.length === 0}
    <div class="state-msg">
      <div class="empty-icon">📋</div>
      <h2>No activities yet</h2>
      <p>Sync your Strava account to see your training here.</p>
    </div>
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
              <td class="date">{a.startDate}</td>
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
</div>

<style>
  .dashboard {
    flex: 1;
    overflow-y: auto;
    padding: 24px 16px;
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

  .state-msg.error {
    color: #f87171;
    opacity: 1;
  }

  .state-msg h2 {
    margin: 0;
    font-size: 1.3rem;
  }

  .state-msg p {
    margin: 0;
    font-size: 0.95rem;
    max-width: 360px;
  }

  .empty-icon {
    font-size: 2.5rem;
    margin-bottom: 4px;
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
    transition: background 0.15s;
  }

  tbody tr:hover {
    background: rgba(255, 255, 255, 0.04);
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

  .sync-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 0;
    margin-bottom: 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    flex-wrap: wrap;
  }

  .btn.btn-sync {
    padding: 8px 18px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 10px;
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
    font-family: inherit;
    white-space: nowrap;
  }

  .btn.btn-sync:hover:not(:disabled) {
    background: #2563eb;
  }

  .btn.btn-sync:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .progress-bar {
    flex: 1;
    min-width: 120px;
    height: 6px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: #3b82f6;
    border-radius: 3px;
    transition: width 0.3s ease;
  }

  .sync-result {
    font-size: 0.8rem;
    color: #94a3b8;
  }
</style>
