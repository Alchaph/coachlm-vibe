<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte'
  import {
    CreateRace, UpdateRace, DeleteRace, ListRaces, SetActiveRace,
    GeneratePlan, GetActivePlan, GetPlanWeeks, UpdateSessionStatus
  } from '../../wailsjs/go/main/App.js'

  const dispatch = createEventDispatcher()

  let races: any[] = []
  let activePlan: any = null
  let weeks: any[] = []
  let loading = true
  let generating = false
  let error = ''
  let errorTimer: ReturnType<typeof setTimeout> | null = null

  let showRaceForm = false
  let editingRace: any = null
  let showPlanView = false
  let selectedSession: any = null
  let completionDuration = ''
  let completionDistance = ''

  let raceName = ''
  let raceDistance = ''
  let raceDate = ''
  let raceTerrain = 'road'
  let raceElevation = ''
  let raceGoalTime = ''
  let racePriority = 'A'

  const DAY_NAMES = ['', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

  const SESSION_COLORS: Record<string, string> = {
    easy: '#4ade80',
    long_run: '#3b82f6',
    tempo: '#fb923c',
    intervals: '#ef4444',
    strength: '#a855f7',
    rest: '#e5e7eb',
    race: '#eab308',
  }

  function showError(msg: string) {
    error = msg
    if (errorTimer) clearTimeout(errorTimer)
    errorTimer = setTimeout(() => { error = '' }, 5000)
  }

  async function loadData() {
    try {
      races = await ListRaces() || []
      try {
        activePlan = await GetActivePlan()
        if (activePlan && activePlan.id) {
          weeks = await GetPlanWeeks(activePlan.id) || []
        } else {
          activePlan = null
          weeks = []
        }
      } catch {
        activePlan = null
        weeks = []
      }
    } catch (e: any) {
      showError(e?.message || String(e) || 'Failed to load data')
    }
  }

  onMount(async () => {
    await loadData()
    loading = false
  })

  function openNewRace() {
    editingRace = null
    raceName = ''
    raceDistance = ''
    raceDate = ''
    raceTerrain = 'road'
    raceElevation = ''
    raceGoalTime = ''
    racePriority = 'A'
    showRaceForm = true
  }

  function openEditRace(race: any) {
    editingRace = race
    raceName = race.name || ''
    raceDistance = String(race.distanceKm || '')
    raceDate = race.raceDate ? race.raceDate.substring(0, 10) : ''
    raceTerrain = race.terrain || 'road'
    raceElevation = race.elevationM != null ? String(race.elevationM) : ''
    raceGoalTime = race.goalTimeSec ? formatGoalTime(race.goalTimeSec) : ''
    racePriority = race.priority || 'A'
    showRaceForm = true
  }

  function parseGoalTime(input: string): number | undefined {
    if (!input.trim()) return undefined
    const parts = input.trim().split(':')
    if (parts.length === 3) {
      return parseInt(parts[0]) * 3600 + parseInt(parts[1]) * 60 + parseInt(parts[2])
    }
    if (parts.length === 2) {
      return parseInt(parts[0]) * 60 + parseInt(parts[1])
    }
    const n = parseInt(input)
    return isNaN(n) ? undefined : n
  }

  function formatGoalTime(secs: number): string {
    const h = Math.floor(secs / 3600)
    const m = Math.floor((secs % 3600) / 60)
    const s = secs % 60
    return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  }

  async function saveRace() {
    if (!raceName.trim()) { showError('Race name is required'); return }
    if (!raceDate) { showError('Race date is required'); return }
    const dist = parseFloat(raceDistance)
    if (isNaN(dist) || dist <= 0) { showError('Valid distance is required'); return }

    const dateVal = new Date(raceDate + 'T00:00:00Z')
    if (isNaN(dateVal.getTime())) { showError('Invalid date'); return }

    const elev = raceElevation.trim() ? parseFloat(raceElevation) : undefined
    const goalSec = parseGoalTime(raceGoalTime)

    const raceObj: any = {
      name: raceName.trim(),
      distanceKm: dist,
      raceDate: dateVal.toISOString(),
      terrain: raceTerrain,
      priority: racePriority,
    }
    if (elev != null && !isNaN(elev)) raceObj.elevationM = elev
    if (goalSec != null) raceObj.goalTimeSec = goalSec

    try {
      if (editingRace) {
        raceObj.id = editingRace.id
        await UpdateRace(raceObj)
      } else {
        await CreateRace(raceObj)
      }
      showRaceForm = false
      await loadData()
    } catch (e: any) {
      showError(e?.message || String(e) || 'Failed to save race')
    }
  }

  async function deleteRace(id: string) {
    if (!confirm('Delete this race and all associated plans?')) return
    try {
      await DeleteRace(id)
      await loadData()
      showPlanView = false
    } catch (e: any) {
      showError(e?.message || String(e) || 'Failed to delete race')
    }
  }

  async function setActive(id: string) {
    try {
      await SetActiveRace(id)
      await loadData()
    } catch (e: any) {
      showError(e?.message || String(e) || 'Failed to set active race')
    }
  }

  async function generatePlan() {
    const activeRace = races.find((r: any) => r.isActive)
    if (!activeRace) { showError('No active race selected'); return }
    generating = true
    try {
      await GeneratePlan(activeRace.id)
      await loadData()
      showPlanView = true
    } catch (e: any) {
      showError(e?.message || String(e) || 'Failed to generate plan')
    } finally {
      generating = false
    }
  }

  async function updateSession(sessionId: string, status: string) {
    const actual: any = {}
    if (status === 'completed') {
      if (completionDuration.trim()) {
        const d = parseInt(completionDuration)
        if (!isNaN(d)) actual.durationMin = d
      }
      if (completionDistance.trim()) {
        const d = parseFloat(completionDistance)
        if (!isNaN(d)) actual.distanceKm = d
      }
    }
    try {
      await UpdateSessionStatus(sessionId, status, actual)
      await loadData()
      if (selectedSession && selectedSession.id === sessionId) {
        for (const w of weeks) {
          const found = (w.sessions || []).find((s: any) => s.id === sessionId)
          if (found) { selectedSession = found; break }
        }
      }
      completionDuration = ''
      completionDistance = ''
    } catch (e: any) {
      showError(e?.message || String(e) || 'Failed to update session')
    }
  }

  function openSession(session: any) {
    selectedSession = session
    completionDuration = ''
    completionDistance = ''
  }

  function closeSession() {
    selectedSession = null
  }

  function openAdjustChat(session: any) {
    const day = DAY_NAMES[session.dayOfWeek] || 'Day'
    const type = (session.type || '').replace('_', ' ')
    let msg = `I want to adjust this session: ${day} - ${type}, ${session.durationMin}min`
    if (session.distanceKm) msg += `, ${session.distanceKm}km`
    if (session.notes) msg += `. Notes: ${session.notes}`
    dispatch('adjustchat', msg)
  }

  function weeksUntilRace(dateStr: string): number {
    const now = new Date()
    const race = new Date(dateStr)
    const diff = race.getTime() - now.getTime()
    return Math.max(0, Math.ceil(diff / (7 * 24 * 60 * 60 * 1000)))
  }

  function isPastWeek(weekStart: string): boolean {
    const now = new Date()
    const start = new Date(weekStart)
    const end = new Date(start.getTime() + 7 * 24 * 60 * 60 * 1000)
    return end < now
  }

  function getSessionColor(type: string): string {
    return SESSION_COLORS[type] || '#94a3b8'
  }

  function isRestType(type: string): boolean {
    return type === 'rest'
  }

  function formatType(type: string): string {
    return (type || '').replace('_', ' ')
  }

  function weekPlannedMin(week: any): number {
    return (week.sessions || []).reduce((sum: number, s: any) => sum + (s.durationMin || 0), 0)
  }

  function weekActualMin(week: any): number {
    return (week.sessions || []).reduce((sum: number, s: any) => {
      if (s.status === 'completed' || s.status === 'modified') {
        return sum + (s.actualDurationMin != null ? s.actualDurationMin : s.durationMin || 0)
      }
      return sum
    }, 0)
  }

  function getActiveRace(): any {
    return races.find((r: any) => r.isActive) || null
  }

  function priorityColor(p: string): string {
    if (p === 'A') return '#ef4444'
    if (p === 'B') return '#fb923c'
    return '#94a3b8'
  }

  function sessionsForDay(week: any, day: number): any[] {
    return (week.sessions || []).filter((s: any) => s.dayOfWeek === day)
  }

  $: activeRace = races.find((r: any) => r.isActive) || null
</script>

<div class="training-plan">
  {#if error}
    <div class="error-banner" role="alert" on:click={() => error = ''} on:keydown={(e) => e.key === 'Enter' && (error = '')}>
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="state-msg">
      <div class="spinner"></div>
      <p>Loading training plans...</p>
    </div>
  {:else if generating}
    <div class="state-msg">
      <div class="spinner"></div>
      <p>Generating your training plan...</p>
    </div>
  {:else if showPlanView && activePlan && weeks.length > 0}
    <!-- Plan Calendar View -->
    <div class="plan-view">
      <div class="plan-header">
        <div>
          <h2 class="plan-title">{activeRace?.name || 'Training Plan'}</h2>
          <span class="plan-subtitle">
            {#if activeRace}
              {activeRace.distanceKm} km &middot; {activeRace.raceDate?.substring(0, 10)} &middot; {weeksUntilRace(activeRace.raceDate)} weeks away
            {/if}
          </span>
        </div>
        <div class="plan-actions">
          <button class="btn btn-outline" on:click={() => { showPlanView = false }}>Back to Races</button>
          <button class="btn btn-primary" on:click={generatePlan}>Regenerate Plan</button>
        </div>
      </div>

      <div class="calendar">
        <div class="calendar-header">
          <div class="week-label-col">Week</div>
          {#each [1,2,3,4,5,6,7] as day}
            <div class="day-col">{DAY_NAMES[day]}</div>
          {/each}
          <div class="summary-col">Total</div>
        </div>

        {#each weeks as week}
          <div class="calendar-row" class:past-week={isPastWeek(week.weekStart)}>
            <div class="week-label-col">
              <span class="week-num">W{week.weekNumber}</span>
              <span class="week-date">{week.weekStart?.substring(5, 10)}</span>
            </div>
            {#each [1,2,3,4,5,6,7] as day}
              <div class="day-col">
                {#each sessionsForDay(week, day) as session}
                  <button
                    class="session-chip"
                    class:rest-chip={isRestType(session.type)}
                    style="background: {getSessionColor(session.type)}20; border-color: {getSessionColor(session.type)}; color: {isRestType(session.type) ? '#1e293b' : getSessionColor(session.type)}"
                    on:click={() => openSession(session)}
                    title="{formatType(session.type)} - {session.durationMin}min"
                  >
                    <span class="chip-type">{formatType(session.type)}</span>
                    <span class="chip-dur">{session.durationMin}m</span>
                  </button>
                {/each}
              </div>
            {/each}
            <div class="summary-col">
              <div class="week-summary">
                <span class="planned-min">{weekPlannedMin(week)}m</span>
                {#if weekActualMin(week) > 0}
                  <span class="actual-min">/ {weekActualMin(week)}m</span>
                {/if}
              </div>
              {#if weekPlannedMin(week) > 0}
                <div class="summary-bar">
                  <div class="summary-fill" style="width: {Math.min(100, Math.round((weekActualMin(week) / weekPlannedMin(week)) * 100))}%"></div>
                </div>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    </div>

    <!-- Session Detail Panel -->
    {#if selectedSession}
      <div class="session-overlay" on:click={closeSession} on:keydown={(e) => e.key === 'Escape' && closeSession()}>
        <div class="session-panel" on:click|stopPropagation on:keydown|stopPropagation>
          <div class="session-header">
            <div class="session-title-row">
              <span class="session-badge" style="background: {getSessionColor(selectedSession.type)}; color: {isRestType(selectedSession.type) ? '#1e293b' : 'white'}">
                {formatType(selectedSession.type)}
              </span>
              <span class="session-day">{DAY_NAMES[selectedSession.dayOfWeek] || ''}</span>
            </div>
            <button class="close-btn" on:click={closeSession} aria-label="Close">&times;</button>
          </div>

          <div class="session-stats">
            <div class="stat">
              <span class="stat-label">Duration</span>
              <span class="stat-value">{selectedSession.durationMin} min</span>
            </div>
            {#if selectedSession.distanceKm}
              <div class="stat">
                <span class="stat-label">Distance</span>
                <span class="stat-value">{selectedSession.distanceKm} km</span>
              </div>
            {/if}
            {#if selectedSession.hrZone}
              <div class="stat">
                <span class="stat-label">HR Zone</span>
                <span class="stat-value">Z{selectedSession.hrZone}</span>
              </div>
            {/if}
            {#if selectedSession.paceMinLow || selectedSession.paceMinHigh}
              <div class="stat">
                <span class="stat-label">Pace</span>
                <span class="stat-value">{selectedSession.paceMinLow || '?'} - {selectedSession.paceMinHigh || '?'} min/km</span>
              </div>
            {/if}
          </div>

          {#if selectedSession.notes}
            <div class="session-notes">
              <span class="notes-label">Notes</span>
              <p>{selectedSession.notes}</p>
            </div>
          {/if}

          <div class="session-status-section">
            <span class="status-label">Status:</span>
            <span class="status-value status-{selectedSession.status}">{selectedSession.status}</span>
          </div>

          {#if selectedSession.status === 'planned'}
            <div class="completion-form">
              <label class="completion-field">
                <span>Actual duration (min)</span>
                <input type="number" bind:value={completionDuration} placeholder="Optional" />
              </label>
              <label class="completion-field">
                <span>Actual distance (km)</span>
                <input type="number" step="0.1" bind:value={completionDistance} placeholder="Optional" />
              </label>
              <div class="completion-actions">
                <button class="btn btn-primary" on:click={() => updateSession(selectedSession.id, 'completed')}>
                  Mark Completed
                </button>
                <button class="btn btn-outline" on:click={() => updateSession(selectedSession.id, 'skipped')}>
                  Mark Skipped
                </button>
              </div>
            </div>
          {/if}

          {#if selectedSession.status === 'completed' && (selectedSession.actualDurationMin || selectedSession.actualDistanceKm)}
            <div class="actual-data">
              <span class="notes-label">Actual</span>
              <div class="session-stats">
                {#if selectedSession.actualDurationMin}
                  <div class="stat">
                    <span class="stat-label">Duration</span>
                    <span class="stat-value">{selectedSession.actualDurationMin} min</span>
                  </div>
                {/if}
                {#if selectedSession.actualDistanceKm}
                  <div class="stat">
                    <span class="stat-label">Distance</span>
                    <span class="stat-value">{selectedSession.actualDistanceKm} km</span>
                  </div>
                {/if}
              </div>
            </div>
          {/if}

          <button class="btn btn-chat" on:click={() => openAdjustChat(selectedSession)}>
            Adjust via Chat
          </button>
        </div>
      </div>
    {/if}

  {:else}
    <!-- Race List View -->
    <div class="race-list-header">
      <h2>Training Plans</h2>
      <button class="btn btn-primary" on:click={openNewRace}>New Race</button>
    </div>

    {#if races.length === 0}
      <div class="state-msg">
        <div class="empty-icon">&#128197;</div>
        <h2>No races yet</h2>
        <p>Create your first goal race to generate a personalised training plan.</p>
        <button class="btn btn-primary" on:click={openNewRace} style="margin-top: 16px">Create Race</button>
      </div>
    {:else}
      <div class="race-cards">
        {#each races as race}
          <div class="race-card" class:race-active={race.isActive}>
            <div class="race-card-header">
              <div class="race-title-row">
                <h3 class="race-name">{race.name}</h3>
                <span class="priority-badge" style="background: {priorityColor(race.priority)}">
                  {race.priority}
                </span>
                {#if race.isActive}
                  <span class="active-badge">Active</span>
                {/if}
              </div>
              <div class="race-meta">
                <span>{race.distanceKm} km</span>
                <span class="meta-sep">&middot;</span>
                <span>{race.terrain}</span>
                <span class="meta-sep">&middot;</span>
                <span>{race.raceDate?.substring(0, 10)}</span>
                <span class="meta-sep">&middot;</span>
                <span>{weeksUntilRace(race.raceDate)} weeks away</span>
                {#if race.goalTimeSec}
                  <span class="meta-sep">&middot;</span>
                  <span>Goal: {formatGoalTime(race.goalTimeSec)}</span>
                {/if}
              </div>
            </div>
            <div class="race-card-actions">
              {#if !race.isActive}
                <button class="btn btn-outline btn-sm" on:click={() => setActive(race.id)}>Set Active</button>
              {/if}
              {#if race.isActive && activePlan}
                <button class="btn btn-primary btn-sm" on:click={() => { showPlanView = true }}>View Plan</button>
              {/if}
              {#if race.isActive && !activePlan}
                <button class="btn btn-primary btn-sm" on:click={generatePlan}>Generate Plan</button>
              {/if}
              <button class="btn btn-outline btn-sm" on:click={() => openEditRace(race)}>Edit</button>
              <button class="btn btn-danger btn-sm" on:click={() => deleteRace(race.id)}>Delete</button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}

  <!-- Race Form Modal -->
  {#if showRaceForm}
    <div class="modal-overlay" on:click={() => showRaceForm = false} on:keydown={(e) => e.key === 'Escape' && (showRaceForm = false)}>
      <div class="modal-content" on:click|stopPropagation on:keydown|stopPropagation>
        <div class="modal-header">
          <h3>{editingRace ? 'Edit Race' : 'New Race'}</h3>
          <button class="close-btn" on:click={() => showRaceForm = false} aria-label="Close">&times;</button>
        </div>
        <div class="modal-body">
          <label class="form-field">
            <span>Race Name</span>
            <input type="text" bind:value={raceName} placeholder="e.g. Berlin Marathon" />
          </label>
          <div class="form-row">
            <label class="form-field">
              <span>Distance (km)</span>
              <input type="number" step="0.1" bind:value={raceDistance} placeholder="42.195" />
            </label>
            <label class="form-field">
              <span>Date</span>
              <input type="date" bind:value={raceDate} />
            </label>
          </div>
          <div class="form-row">
            <label class="form-field">
              <span>Terrain</span>
              <select bind:value={raceTerrain}>
                <option value="road">Road</option>
                <option value="trail">Trail</option>
                <option value="track">Track</option>
              </select>
            </label>
            <label class="form-field">
              <span>Priority</span>
              <select bind:value={racePriority}>
                <option value="A">A (Primary)</option>
                <option value="B">B (Secondary)</option>
                <option value="C">C (Low)</option>
              </select>
            </label>
          </div>
          <div class="form-row">
            <label class="form-field">
              <span>Elevation (m, optional)</span>
              <input type="number" bind:value={raceElevation} placeholder="e.g. 500" />
            </label>
            <label class="form-field">
              <span>Goal Time (optional)</span>
              <input type="text" bind:value={raceGoalTime} placeholder="e.g. 3:30:00" />
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-outline" on:click={() => showRaceForm = false}>Cancel</button>
          <button class="btn btn-primary" on:click={saveRace}>
            {editingRace ? 'Update' : 'Create'}
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .training-plan {
    flex: 1;
    overflow-y: auto;
    padding: 24px 16px;
    position: relative;
  }

  .error-banner {
    position: sticky;
    top: 0;
    background: #dc3545;
    color: white;
    padding: 10px 16px;
    border-radius: 8px;
    font-size: 0.9rem;
    z-index: 10;
    cursor: pointer;
    text-align: center;
    margin-bottom: 12px;
    animation: slideDown 0.3s ease;
  }

  @keyframes slideDown {
    from { opacity: 0; transform: translateY(-10px); }
    to { opacity: 1; transform: translateY(0); }
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

  .state-msg h2 { margin: 0; font-size: 1.3rem; }
  .state-msg p { margin: 0; font-size: 0.95rem; max-width: 360px; line-height: 1.5; }

  .empty-icon { font-size: 2.5rem; margin-bottom: 4px; }

  .spinner {
    width: 28px;
    height: 28px;
    border: 3px solid rgba(255, 255, 255, 0.15);
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* Buttons */
  .btn {
    padding: 8px 18px;
    border: none;
    border-radius: 10px;
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
    font-family: inherit;
    white-space: nowrap;
  }
  .btn:disabled { opacity: 0.6; cursor: not-allowed; }
  .btn-primary { background: #3b82f6; color: white; }
  .btn-primary:hover:not(:disabled) { background: #2563eb; }
  .btn-outline {
    background: transparent;
    color: #94a3b8;
    border: 1px solid rgba(255, 255, 255, 0.15);
  }
  .btn-outline:hover:not(:disabled) { background: rgba(255, 255, 255, 0.05); color: #e2e8f0; }
  .btn-danger { background: transparent; color: #f87171; border: 1px solid rgba(248, 113, 113, 0.3); }
  .btn-danger:hover:not(:disabled) { background: rgba(248, 113, 113, 0.1); }
  .btn-sm { padding: 5px 12px; font-size: 0.8rem; }
  .btn-chat {
    width: 100%;
    background: rgba(255, 255, 255, 0.08);
    color: #94a3b8;
    margin-top: 12px;
  }
  .btn-chat:hover { background: rgba(255, 255, 255, 0.12); color: #e2e8f0; }

  /* Race List */
  .race-list-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }
  .race-list-header h2 { margin: 0; font-size: 1.3rem; color: #e2e8f0; }

  .race-cards { display: flex; flex-direction: column; gap: 12px; }

  .race-card {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    padding: 16px;
    transition: background 0.15s;
  }
  .race-card:hover { background: rgba(255, 255, 255, 0.06); }
  .race-card.race-active { border-color: rgba(59, 130, 246, 0.4); }

  .race-card-header { margin-bottom: 12px; }

  .race-title-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
  }
  .race-name { margin: 0; font-size: 1.05rem; color: #e2e8f0; font-weight: 600; }

  .priority-badge {
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 0.7rem;
    font-weight: 700;
    color: white;
  }
  .active-badge {
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 0.7rem;
    font-weight: 600;
    background: rgba(59, 130, 246, 0.2);
    color: #3b82f6;
  }

  .race-meta {
    font-size: 0.8rem;
    color: #94a3b8;
    display: flex;
    align-items: center;
    gap: 4px;
    flex-wrap: wrap;
  }
  .meta-sep { opacity: 0.5; }

  .race-card-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  /* Plan View */
  .plan-view { display: flex; flex-direction: column; gap: 16px; }

  .plan-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 16px;
    flex-wrap: wrap;
  }
  .plan-title { margin: 0; font-size: 1.3rem; color: #e2e8f0; }
  .plan-subtitle { font-size: 0.85rem; color: #94a3b8; }
  .plan-actions { display: flex; gap: 8px; }

  /* Calendar */
  .calendar {
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    overflow: hidden;
  }

  .calendar-header {
    display: grid;
    grid-template-columns: 70px repeat(7, 1fr) 90px;
    background: rgba(255, 255, 255, 0.06);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .calendar-header .day-col,
  .calendar-header .week-label-col,
  .calendar-header .summary-col {
    padding: 8px 6px;
    font-size: 0.75rem;
    font-weight: 600;
    color: #94a3b8;
    text-transform: uppercase;
    text-align: center;
  }

  .calendar-row {
    display: grid;
    grid-template-columns: 70px repeat(7, 1fr) 90px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.04);
    transition: background 0.15s;
  }
  .calendar-row:last-child { border-bottom: none; }
  .calendar-row:hover { background: rgba(255, 255, 255, 0.02); }
  .calendar-row.past-week { opacity: 0.6; }

  .week-label-col {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 8px 4px;
    gap: 2px;
  }
  .week-num { font-size: 0.8rem; font-weight: 600; color: #e2e8f0; }
  .week-date { font-size: 0.7rem; color: #64748b; }

  .day-col {
    padding: 6px 4px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    min-height: 50px;
  }

  .session-chip {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1px;
    padding: 4px 6px;
    border-radius: 6px;
    border: 1px solid;
    cursor: pointer;
    font-family: inherit;
    width: 100%;
    transition: opacity 0.15s;
  }
  .session-chip:hover { opacity: 0.8; }
  .chip-type { font-size: 0.65rem; font-weight: 600; text-transform: capitalize; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 100%; }
  .chip-dur { font-size: 0.6rem; opacity: 0.8; }

  .summary-col {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 6px 8px;
    gap: 4px;
  }
  .week-summary { font-size: 0.75rem; color: #94a3b8; display: flex; gap: 4px; }
  .planned-min { color: #e2e8f0; }
  .actual-min { color: #4ade80; }

  .summary-bar {
    width: 100%;
    height: 4px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 2px;
    overflow: hidden;
  }
  .summary-fill {
    height: 100%;
    background: #4ade80;
    border-radius: 2px;
    transition: width 0.3s ease;
  }

  /* Session Detail Panel */
  .session-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .session-panel {
    background: rgba(15, 23, 36, 0.98);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 16px;
    padding: 24px;
    width: 400px;
    max-width: 90vw;
    max-height: 80vh;
    overflow-y: auto;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  }

  .session-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }
  .session-title-row { display: flex; align-items: center; gap: 8px; }

  .session-badge {
    padding: 4px 10px;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: capitalize;
  }
  .session-day { font-size: 0.9rem; color: #94a3b8; }

  .close-btn {
    background: none;
    border: none;
    color: #94a3b8;
    font-size: 1.4rem;
    cursor: pointer;
    padding: 0 4px;
    line-height: 1;
    transition: color 0.15s;
  }
  .close-btn:hover { color: #e2e8f0; }

  .session-stats {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    margin-bottom: 14px;
  }
  .stat {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .stat .stat-label { font-size: 0.7rem; color: #64748b; text-transform: uppercase; font-weight: 600; }
  .stat .stat-value { font-size: 0.9rem; color: #e2e8f0; }

  .session-notes {
    margin-bottom: 14px;
    padding: 10px;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 8px;
  }
  .notes-label { font-size: 0.7rem; color: #64748b; text-transform: uppercase; font-weight: 600; }
  .session-notes p { margin: 4px 0 0; font-size: 0.85rem; color: #e2e8f0; line-height: 1.5; }

  .session-status-section {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 14px;
  }
  .status-label { font-size: 0.8rem; color: #94a3b8; }
  .status-value {
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: capitalize;
    padding: 2px 8px;
    border-radius: 4px;
  }
  .status-planned { background: rgba(148, 163, 184, 0.15); color: #94a3b8; }
  .status-completed { background: rgba(74, 222, 128, 0.15); color: #4ade80; }
  .status-skipped { background: rgba(248, 113, 113, 0.15); color: #f87171; }
  .status-modified { background: rgba(251, 146, 60, 0.15); color: #fb923c; }

  .completion-form {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 14px;
  }
  .completion-field {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 0.8rem;
    color: #94a3b8;
  }
  .completion-field input {
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.08);
    color: white;
    padding: 8px 12px;
    font-family: inherit;
    font-size: 0.85rem;
    outline: none;
  }
  .completion-field input:focus { border-color: #3b82f6; }
  .completion-actions { display: flex; gap: 8px; }

  .actual-data { margin-bottom: 14px; }

  /* Modal */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    z-index: 200;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .modal-content {
    background: rgba(15, 23, 36, 0.98);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 16px;
    width: 480px;
    max-width: 90vw;
    max-height: 80vh;
    overflow-y: auto;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  }
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px 12px;
  }
  .modal-header h3 { margin: 0; font-size: 1.1rem; color: #e2e8f0; }
  .modal-body { padding: 0 24px; display: flex; flex-direction: column; gap: 12px; }
  .modal-footer { padding: 16px 24px; display: flex; justify-content: flex-end; gap: 8px; }

  .form-field {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 0.8rem;
    color: #94a3b8;
  }
  .form-field input,
  .form-field select {
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.08);
    color: white;
    padding: 8px 12px;
    font-family: inherit;
    font-size: 0.85rem;
    outline: none;
  }
  .form-field input:focus,
  .form-field select:focus { border-color: #3b82f6; }
  .form-field select { cursor: pointer; }
  .form-field select option { background: #1b2636; color: #e2e8f0; }

  .form-row { display: flex; gap: 12px; }
  .form-row .form-field { flex: 1; }
</style>
