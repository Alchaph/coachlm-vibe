/**
 * E2E tests: dashboard tab (S15, S30, S35)
 * Covers: activity table, stats bar, sync button behavior.
 */
import { test, expect } from '@playwright/test'
import path from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: path.join(__dirname, 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Dashboard"]')
})

test('renders the dashboard', async ({ page }) => {
  await expect(page.locator('.dashboard')).toBeVisible()
})

test('shows stats bar with activity count and total distance', async ({ page }) => {
  const statsBar = page.locator('.stats-bar')
  await expect(statsBar).toBeVisible()
  // Default mock: 42 activities, 380.5 km
  await expect(statsBar).toContainText('42')
  await expect(statsBar).toContainText('380.5 km')
})

test('shows activity table with rows', async ({ page }) => {
  await expect(page.locator('table')).toBeVisible()
  // Default mock returns 2 activities
  await expect(page.locator('tbody tr')).toHaveCount(2)
})

test('activity table has expected columns', async ({ page }) => {
  const headers = page.locator('thead th')
  await expect(headers).toHaveCount(7)
  await expect(headers.nth(0)).toContainText('Date')
  await expect(headers.nth(1)).toContainText('Name')
  await expect(headers.nth(2)).toContainText('Type')
  await expect(headers.nth(3)).toContainText('Distance')
  await expect(headers.nth(4)).toContainText('Duration')
  await expect(headers.nth(5)).toContainText('Pace')
  await expect(headers.nth(6)).toContainText('HR')
})

test('first activity row contains mock data', async ({ page }) => {
  const firstRow = page.locator('tbody tr').first()
  await expect(firstRow).toContainText('Morning Run')
  await expect(firstRow).toContainText('Run')
  await expect(firstRow).toContainText('10.5 km')
})

test('sync button is not shown when strava is not connected', async ({ page }) => {
  // Default mock: stravaConnected = false
  await expect(page.locator('button.btn-sync')).not.toBeVisible()
})

test('sync button is shown when strava is connected', async ({ page }) => {
  await page.addInitScript(() => {
    window.__WAILS_MOCK_STATE__.stravaConnected = true
  })
  await page.reload()
  await page.click('button[title="Dashboard"]')
  await expect(page.locator('button.btn-sync')).toBeVisible()
})

test('empty state when no activities', async ({ page }) => {
  await page.addInitScript(() => {
    window.__WAILS_MOCK_STATE__.activities = []
    window.__WAILS_MOCK_STATE__.stats = { totalCount: 0, totalDistanceKm: 0, earliestDate: '', latestDate: '' }
  })
  await page.reload()
  await page.click('button[title="Dashboard"]')
  await expect(page.locator('.state-msg')).toBeVisible()
  await expect(page.locator('.state-msg')).toContainText('No activities yet')
})
