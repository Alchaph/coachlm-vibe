/**
 * E2E tests: S46 — Free AI (Gemini) backend saving
 * Covers: selecting free backend, saving without API key, success feedback.
 */
import { test, expect } from '@playwright/test'
import path from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: path.join(__dirname, 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Settings"]')
  await expect(page.locator('.settings select#active-backend')).toBeVisible()
})

test('"Free (Gemini Flash)" option exists in backend selector', async ({ page }) => {
  const option = page.locator('#active-backend option[value="free"]')
  await expect(option).toHaveText('Free (Gemini Flash)')
})

test('selecting free shows no-setup-required note', async ({ page }) => {
  await page.selectOption('#active-backend', 'free')
  await expect(page.locator('.field-note').first()).toContainText('No setup required')
})

test('saving with free backend succeeds (no error shown)', async ({ page }) => {
  await page.selectOption('#active-backend', 'free')
  await page.click('button.save-btn')
  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })
  await expect(page.locator('.feedback.success')).toContainText('Settings saved')
})

test('free backend is saved to mock state', async ({ page }) => {
  await page.selectOption('#active-backend', 'free')
  await page.click('button.save-btn')
  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })

  // Verify in-page mock state
  const activeLlm = await page.evaluate(() => window.__WAILS_MOCK_STATE__.settings.activeLlm)
  expect(activeLlm).toBe('free')
})

test('free backend persists after tab switch and return', async ({ page }) => {
  await page.selectOption('#active-backend', 'free')
  await page.click('button.save-btn')
  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })

  await page.click('button[title="Chat"]')
  await page.click('button[title="Settings"]')
  await expect(page.locator('.settings select#active-backend')).toBeVisible()

  await expect(page.locator('#active-backend')).toHaveValue('free')
})

test('onboarding step 2 can select and continue with free backend', async ({ page }) => {
  // Restart with first-run flag
  await page.addInitScript(() => { window.__WAILS_MOCK_STATE__.isFirstRun = true })
  await page.reload()
  await expect(page.locator('.overlay')).toBeVisible()

  await page.click('button', { hasText: 'Get Started' })
  await page.selectOption('#onboarding-backend', 'free')
  await expect(page.locator('.field-note')).toContainText('No setup required')
  await page.click('button', { hasText: 'Next' })
  await expect(page.locator('.step h1')).toContainText('Connect Strava')
})
