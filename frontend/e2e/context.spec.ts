/**
 * E2E tests: context tab (S29, S33)
 * Covers: athlete profile form, pinned insights list, training summary table.
 */
import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Context"]')
})

test('renders the context page', async ({ page }) => {
  await expect(page.locator('.context')).toBeVisible()
})

test('shows Athlete Profile section', async ({ page }) => {
  await expect(page.locator('section h2').first()).toContainText('Athlete Profile')
})

test('profile form loads mock data', async ({ page }) => {
  await expect(page.locator('#age')).toHaveValue('32')
  await expect(page.locator('#max-hr')).toHaveValue('185')
})

test('can update age field', async ({ page }) => {
  const ageInput = page.locator('#age')
  await ageInput.fill('35')
  await expect(ageInput).toHaveValue('35')
})

test('save profile button submits successfully', async ({ page }) => {
  const ageInput = page.locator('#age')
  await ageInput.fill('35')
  await page.click('button.btn-primary')

  // Feedback should appear
  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })
  await expect(page.locator('.feedback.success')).toContainText('Profile saved')
})

test('shows Pinned Insights section', async ({ page }) => {
  const headings = page.locator('section h2')
  await expect(headings.nth(1)).toContainText('Pinned Insights')
})

test('displays mock pinned insight', async ({ page }) => {
  await expect(page.locator('.insight-item')).toHaveCount(1)
  await expect(page.locator('.insight-content')).toContainText('Focus on easy aerobic base')
})

test('can delete a pinned insight', async ({ page }) => {
  await expect(page.locator('.insight-item')).toHaveCount(1)
  await page.locator('.delete-btn').click()
  await expect(page.locator('.insight-item')).toHaveCount(0)
})

test('shows empty pinned insights message when no insights', async ({ page }) => {
  await page.addInitScript(() => {
    window.__WAILS_MOCK_STATE__.insights = []
  })
  await page.reload()
  await page.click('button[title="Context"]')
  await expect(page.locator('.empty-text').first()).toContainText('No pinned insights yet')
})

test('shows Training Summary section with activity table', async ({ page }) => {
  const headings = page.locator('section h2')
  await expect(headings.nth(2)).toContainText('Training Summary')
  await expect(page.locator('section').nth(2).locator('table')).toBeVisible()
})

test('shows empty training message when no activities', async ({ page }) => {
  await page.addInitScript(() => {
    window.__WAILS_MOCK_STATE__.activities = []
  })
  await page.reload()
  await page.click('button[title="Context"]')
  await expect(page.locator('.empty-text').last()).toContainText('No activities yet')
})
