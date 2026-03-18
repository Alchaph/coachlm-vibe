import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Training Plan"]')
})

test('renders the training plan tab with race list', async ({ page }) => {
  await expect(page.locator('.training-plan')).toBeVisible()
  await expect(page.locator('.race-list-header h2')).toContainText('Training Plans')
})

test('shows race card with mock data', async ({ page }) => {
  const card = page.locator('.race-card').first()
  await expect(card).toBeVisible()
  await expect(card).toContainText('Berlin Marathon')
  await expect(card).toContainText('42.195 km')
  await expect(card).toContainText('road')
  await expect(card).toContainText('2026-10-15')
})

test('active race shows active badge', async ({ page }) => {
  await expect(page.locator('.active-badge')).toContainText('Active')
})

test('active race shows View Plan button', async ({ page }) => {
  await expect(page.locator('.race-card-actions button', { hasText: 'View Plan' })).toBeVisible()
})

test('clicking View Plan shows calendar view', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await expect(page.locator('.plan-view')).toBeVisible()
  await expect(page.locator('.calendar')).toBeVisible()
  await expect(page.locator('.plan-title')).toContainText('Berlin Marathon')
})

test('calendar shows week rows with session chips', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await expect(page.locator('.calendar-row')).toHaveCount(2)
  const chips = page.locator('.session-chip')
  await expect(chips).toHaveCount(7)
})

test('session chip shows type and duration', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  const firstChip = page.locator('.session-chip').first()
  await expect(firstChip).toContainText('easy')
  await expect(firstChip).toContainText('45m')
})

test('clicking session chip opens detail panel', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await page.locator('.session-chip').first().click()
  await expect(page.locator('.session-panel')).toBeVisible()
  await expect(page.locator('.session-badge')).toContainText('easy')
  await expect(page.locator('.stat-value', { hasText: '45 min' })).toBeVisible()
})

test('session detail shows notes', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await page.locator('.session-chip').first().click()
  await expect(page.locator('.session-notes p')).toContainText('Easy aerobic run')
})

test('session detail shows status as planned', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await page.locator('.session-chip').first().click()
  await expect(page.locator('.status-value')).toContainText('planned')
})

test('planned session has completion actions', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await page.locator('.session-chip').first().click()
  await expect(page.locator('.completion-actions button', { hasText: 'Mark Completed' })).toBeVisible()
  await expect(page.locator('.completion-actions button', { hasText: 'Mark Skipped' })).toBeVisible()
})

test('session detail has Adjust via Chat button', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await page.locator('.session-chip').first().click()
  await expect(page.locator('.btn-chat')).toContainText('Adjust via Chat')
})

test('Adjust via Chat switches to chat tab with pre-seeded message', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await page.locator('.session-chip').first().click()
  await page.click('.btn-chat')
  await expect(page.locator('.input-area textarea')).toBeVisible()
  const val = await page.locator('.input-area textarea').inputValue()
  expect(val).toContain('adjust this session')
})

test('closing session detail overlay', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await page.locator('.session-chip').first().click()
  await expect(page.locator('.session-panel')).toBeVisible()
  await page.locator('.session-panel .close-btn').click()
  await expect(page.locator('.session-panel')).not.toBeVisible()
})

test('Back to Races button returns to race list', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await expect(page.locator('.plan-view')).toBeVisible()
  await page.click('.btn-outline:has-text("Back to Races")')
  await expect(page.locator('.race-list-header')).toBeVisible()
})

test('New Race button opens race form modal', async ({ page }) => {
  await page.click('.race-list-header button:has-text("New Race")')
  await expect(page.locator('.modal-content')).toBeVisible()
  await expect(page.locator('.modal-header h3')).toContainText('New Race')
})

test('race form has all required fields', async ({ page }) => {
  await page.click('.race-list-header button:has-text("New Race")')
  await expect(page.locator('.form-field input[placeholder*="Berlin Marathon"]')).toBeVisible()
  await expect(page.locator('.form-field input[type="number"][placeholder*="42"]')).toBeVisible()
  await expect(page.locator('.form-field input[type="date"]')).toBeVisible()
  await expect(page.locator('.form-field select')).toHaveCount(2)
})

test('creating a new race via form', async ({ page }) => {
  await page.click('.race-list-header button:has-text("New Race")')
  await page.fill('input[placeholder*="Berlin Marathon"]', 'London Marathon')
  await page.fill('input[placeholder*="42"]', '42.195')
  await page.fill('input[type="date"]', '2026-11-01')
  await page.click('.modal-footer button:has-text("Create")')
  await expect(page.locator('.modal-content')).not.toBeVisible()
  await expect(page.locator('.race-card')).toHaveCount(2)
  await expect(page.locator('.race-card').last()).toContainText('London Marathon')
})

test('edit button opens race form with existing data', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("Edit")')
  await expect(page.locator('.modal-header h3')).toContainText('Edit Race')
  const nameInput = page.locator('input[placeholder*="Berlin Marathon"]')
  await expect(nameInput).toHaveValue('Berlin Marathon')
})

test('weekly summary shows planned minutes', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  const summaries = page.locator('.week-summary')
  await expect(summaries.first()).toBeVisible()
  await expect(summaries.first()).toContainText('185m')
})

test('empty state when no races', async ({ page }) => {
  await page.addInitScript(() => {
    window.__WAILS_MOCK_STATE__.races = []
    window.__WAILS_MOCK_STATE__.activePlan = null
    window.__WAILS_MOCK_STATE__.planWeeks = []
  })
  await page.reload()
  await page.click('button[title="Training Plan"]')
  await expect(page.locator('.state-msg')).toBeVisible()
  await expect(page.locator('.state-msg')).toContainText('No races yet')
})

test('mark session completed updates status', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await page.locator('.session-chip').first().click()
  await page.click('.completion-actions button:has-text("Mark Completed")')
  await expect(page.locator('.status-value')).toContainText('completed')
})

test('mark session skipped updates status', async ({ page }) => {
  await page.click('.race-card-actions button:has-text("View Plan")')
  await page.locator('.session-chip').first().click()
  await page.click('.completion-actions button:has-text("Mark Skipped")')
  await expect(page.locator('.status-value')).toContainText('skipped')
})
