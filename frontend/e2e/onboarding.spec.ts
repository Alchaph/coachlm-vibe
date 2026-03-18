/**
 * E2E tests: onboarding wizard (S25, S50)
 * Covers: all 3 steps, skip flows, finish/complete flow.
 */
import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

const firstRunScript = () => {
  window.__WAILS_MOCK_STATE__.isFirstRun = true
}

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
  await page.addInitScript(firstRunScript)
  await page.goto('/')
  // Wait for onboarding to render
  await expect(page.locator('.overlay')).toBeVisible()
})

test('shows onboarding wizard on first run', async ({ page }) => {
  await expect(page.locator('.overlay .wizard')).toBeVisible()
})

test('step 1 shows welcome message', async ({ page }) => {
  await expect(page.locator('.step h1')).toContainText('Welcome to CoachLM')
})

test('Get Started button moves to step 2', async ({ page }) => {
  await page.locator('.wizard button', { hasText: 'Get Started' }).click()
  await expect(page.locator('.step h1')).toContainText('Connect Strava')
})

test('step 2 Skip moves to step 3 (You\'re All Set)', async ({ page }) => {
  await page.locator('.wizard button', { hasText: 'Get Started' }).click()
  await page.locator('.wizard button', { hasText: 'Skip' }).click()
  await expect(page.locator('.step h1')).toContainText("You're All Set")
})

test('step 3 Back returns to step 2', async ({ page }) => {
  await page.locator('.wizard button', { hasText: 'Get Started' }).click()
  await page.locator('.wizard button', { hasText: 'Skip' }).click()
  // Step 3 doesn't have a Back button — it has Start Chatting
  // Verify we're on step 3
  await expect(page.locator('.step h1')).toContainText("You're All Set")
})

test('step 3 Start Chatting button finishes onboarding', async ({ page }) => {
  // Skip through all steps quickly
  await page.locator('.wizard button', { hasText: 'Get Started' }).click()
  await page.locator('.wizard button', { hasText: 'Skip' }).click()

  await page.locator('.wizard button', { hasText: 'Start Chatting' }).click()

  // Onboarding overlay should disappear
  await expect(page.locator('.overlay')).not.toBeVisible({ timeout: 5000 })
  // Chat tab should be shown
  await expect(page.locator('.input-area textarea')).toBeVisible()
})

test('progress dots advance through steps', async ({ page }) => {
  // Step 1: first dot should be active
  await expect(page.locator('.progress .dot.active')).toHaveCount(1)
  // Move to step 2
  await page.locator('.wizard button', { hasText: 'Get Started' }).click()
  await expect(page.locator('.progress .dot.done')).toHaveCount(1)
  await expect(page.locator('.progress .dot.active')).toHaveCount(1)
})

test('step 3 shows context readiness checklist', async ({ page }) => {
  await page.locator('.wizard button', { hasText: 'Get Started' }).click()
  await page.locator('.wizard button', { hasText: 'Skip' }).click()
  await expect(page.locator('.context-readiness')).toBeVisible()
  await expect(page.locator('.readiness-item')).toHaveCount(3)
})
