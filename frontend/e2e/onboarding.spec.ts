/**
 * E2E tests: onboarding wizard (S25)
 * Covers: all 5 steps, skip flows, finish/complete flow.
 */
import { test, expect } from '@playwright/test'
import path from 'path'

const firstRunScript = () => {
  window.__WAILS_MOCK_STATE__.isFirstRun = true
}

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: path.join(__dirname, 'mocks/wails.ts') })
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
  await page.click('button', { hasText: 'Get Started' })
  await expect(page.locator('.step h1')).toContainText('Choose Your AI Backend')
})

test('step 2 backend selector defaults to free', async ({ page }) => {
  await page.click('button', { hasText: 'Get Started' })
  await expect(page.locator('#onboarding-backend')).toHaveValue('free')
})

test('step 2 shows no-setup note for free backend', async ({ page }) => {
  await page.click('button', { hasText: 'Get Started' })
  await expect(page.locator('.field-note')).toContainText('No setup required')
})

test('step 2 Next moves to step 3 (Connect Strava)', async ({ page }) => {
  await page.click('button', { hasText: 'Get Started' })
  await page.click('button', { hasText: 'Next' })
  await expect(page.locator('.step h1')).toContainText('Connect Strava')
})

test('step 3 Back returns to step 2', async ({ page }) => {
  await page.click('button', { hasText: 'Get Started' })
  await page.click('button', { hasText: 'Next' })
  await page.click('button', { hasText: 'Back' })
  await expect(page.locator('.step h1')).toContainText('Choose Your AI Backend')
})

test('step 3 Skip moves to step 4 (Athlete Profile)', async ({ page }) => {
  await page.click('button', { hasText: 'Get Started' })
  await page.click('button', { hasText: 'Next' })
  await page.click('button', { hasText: 'Skip' })
  await expect(page.locator('.step h1')).toContainText('Athlete Profile')
})

test('step 4 Skip moves to step 5 (You\'re All Set)', async ({ page }) => {
  await page.click('button', { hasText: 'Get Started' })
  await page.click('button', { hasText: 'Next' })
  await page.click('button', { hasText: 'Skip' })
  // On step 4 there are two Skip buttons; use the one in actions
  await page.locator('.actions button', { hasText: 'Skip' }).click()
  await expect(page.locator('.step h1')).toContainText("You're All Set")
})

test('step 5 Start Chatting button finishes onboarding', async ({ page }) => {
  // Skip through all steps quickly
  await page.click('button', { hasText: 'Get Started' })
  await page.click('button', { hasText: 'Next' })
  await page.click('button', { hasText: 'Skip' })
  await page.locator('.actions button', { hasText: 'Skip' }).click()

  await page.click('button', { hasText: 'Start Chatting' })

  // Onboarding overlay should disappear
  await expect(page.locator('.overlay')).not.toBeVisible({ timeout: 5000 })
  // Chat tab should be shown
  await expect(page.locator('.input-area textarea')).toBeVisible()
})

test('progress dots advance through steps', async ({ page }) => {
  // Step 1: first dot should be active
  await expect(page.locator('.progress .dot.active')).toHaveCount(1)
  // Move to step 2
  await page.click('button', { hasText: 'Get Started' })
  await expect(page.locator('.progress .dot.done')).toHaveCount(1)
  await expect(page.locator('.progress .dot.active')).toHaveCount(1)
})

test('step 2 selecting claude shows API key field', async ({ page }) => {
  await page.click('button', { hasText: 'Get Started' })
  await page.selectOption('#onboarding-backend', 'claude')
  await expect(page.locator('#onboarding-claude-api-key')).toBeVisible()
})
