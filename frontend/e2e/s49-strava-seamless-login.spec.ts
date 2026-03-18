/**
 * E2E tests: S49 — Seamless Strava login (no credential fields)
 * Covers: no credential inputs in settings or onboarding, connect flow via mock.
 */
import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

test.describe('Settings — Strava seamless login', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
    await page.goto('/')
    await page.click('button[title="Settings"]')
    await expect(page.locator('.settings')).toBeVisible()
  })

  test('no Strava credential input fields in settings', async ({ page }) => {
    await expect(page.locator('#strava-client-id')).not.toBeVisible()
    await expect(page.locator('#strava-client-secret')).not.toBeVisible()
  })

  test('Connect Strava button is visible when credentials available', async ({ page }) => {
    await expect(page.locator('button', { hasText: 'Connect Strava' })).toBeVisible()
  })

  test('Connect Strava button is not disabled', async ({ page }) => {
    await expect(page.locator('button', { hasText: 'Connect Strava' })).toBeEnabled()
  })

  test('clicking Connect Strava triggers auth flow', async ({ page }) => {
    await page.locator('button', { hasText: 'Connect Strava' }).click()
    await expect(page.locator('.feedback')).toBeVisible({ timeout: 3000 })
  })

  test('shows unavailable note when credentials absent', async ({ page }) => {
    await page.addInitScript(() => {
      window.go.main.App.GetStravaCredentialsAvailable = () =>
        new Promise((r) => setTimeout(() => r(false), 50))
    })
    await page.goto('/')
    await page.click('button[title="Settings"]')
    await expect(page.locator('.settings')).toBeVisible()
    await expect(page.locator('.strava-unavailable')).toBeVisible()
    await expect(page.locator('.strava-unavailable')).toContainText('Not available in this build')
  })

  test('Disconnect button visible when connected', async ({ page }) => {
    await page.addInitScript(() => {
      window.__WAILS_MOCK_STATE__.stravaConnected = true
    })
    await page.goto('/')
    await page.click('button[title="Settings"]')
    await expect(page.locator('.settings')).toBeVisible()
    await expect(page.locator('button', { hasText: 'Disconnect' })).toBeVisible()
  })
})

test.describe('Onboarding — Strava seamless login', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
    await page.addInitScript(() => {
      window.__WAILS_MOCK_STATE__.isFirstRun = true
    })
    await page.goto('/')
    await expect(page.locator('.overlay')).toBeVisible()
    await page.locator('.wizard button', { hasText: 'Get Started' }).click()
    await expect(page.locator('.step h1')).toContainText('Connect Strava')
  })

  test('no credential input fields in onboarding step 2', async ({ page }) => {
    await expect(page.locator('#onboarding-strava-client-id')).not.toBeVisible()
    await expect(page.locator('#onboarding-strava-client-secret')).not.toBeVisible()
  })

  test('Connect Strava button visible in onboarding', async ({ page }) => {
    await expect(page.locator('.wizard button', { hasText: 'Connect Strava' })).toBeVisible()
  })

  test('Connect Strava button is enabled without credential inputs', async ({ page }) => {
    await expect(page.locator('.wizard button', { hasText: 'Connect Strava' })).toBeEnabled()
  })

  test('shows unavailable note when credentials absent in onboarding', async ({ page }) => {
    await page.addInitScript(() => {
      window.go.main.App.GetStravaCredentialsAvailable = () =>
        new Promise((r) => setTimeout(() => r(false), 50))
    })
    await page.goto('/')
    await expect(page.locator('.overlay')).toBeVisible()
    await page.locator('.wizard button', { hasText: 'Get Started' }).click()
    await expect(page.locator('.step h1')).toContainText('Connect Strava')
    await expect(page.locator('.strava-unavailable')).toBeVisible()
    await expect(page.locator('.strava-unavailable')).toContainText('Not available in this build')
  })
})
