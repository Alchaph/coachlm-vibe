/**
 * E2E tests: S43 — Advice-first coaching with on-demand plans
 * Covers: Generate Training Plan button, inline goal input panel, validation,
 * plan request sends structured prompt via SendMessage.
 */
import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

const mockPath = join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts')

test.describe('Chat — Generate Training Plan button', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript({ path: mockPath })
    await page.goto('/')
    // Chat tab is the default
    await expect(page.locator('.chat-app')).toBeVisible()
  })

  test('plan button is visible in chat input area', async ({ page }) => {
    await expect(page.locator('button[aria-label="Generate Training Plan"]')).toBeVisible()
  })

  test('plan button has clipboard icon', async ({ page }) => {
    const btn = page.locator('button[aria-label="Generate Training Plan"]')
    await expect(btn.locator('svg')).toBeVisible()
  })

  test('clicking plan button opens goal input panel', async ({ page }) => {
    await expect(page.locator('.plan-input-panel')).not.toBeVisible()
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await expect(page.locator('.plan-input-panel')).toBeVisible()
  })

  test('goal input panel has race type, target date, and target time fields', async ({ page }) => {
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    const panel = page.locator('.plan-input-panel')
    await expect(panel).toBeVisible()

    await expect(panel.locator('input[type="text"][placeholder*="5K"]')).toBeVisible()
    await expect(panel.locator('input[type="date"]')).toBeVisible()
    await expect(panel.locator('input[type="text"][placeholder*="3:30"]')).toBeVisible()
  })

  test('goal input panel has "Training Plan Goal" header', async ({ page }) => {
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await expect(page.locator('.plan-header')).toContainText('Training Plan Goal')
  })

  test('clicking close button hides goal input panel', async ({ page }) => {
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await expect(page.locator('.plan-input-panel')).toBeVisible()
    await page.locator('.plan-close').click()
    await expect(page.locator('.plan-input-panel')).not.toBeVisible()
  })

  test('clicking plan button again toggles panel closed', async ({ page }) => {
    const btn = page.locator('button[aria-label="Generate Training Plan"]')
    await btn.click()
    await expect(page.locator('.plan-input-panel')).toBeVisible()
    await btn.click()
    await expect(page.locator('.plan-input-panel')).not.toBeVisible()
  })

  test('submitting with all empty fields shows validation error', async ({ page }) => {
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await page.locator('.plan-submit').click()
    await expect(page.locator('.plan-error')).toBeVisible()
    await expect(page.locator('.plan-error')).toContainText('Enter at least a race type, date, or target time')
    // Panel stays open
    await expect(page.locator('.plan-input-panel')).toBeVisible()
  })

  test('validation error clears when panel is reopened', async ({ page }) => {
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await page.locator('.plan-submit').click()
    await expect(page.locator('.plan-error')).toBeVisible()
    // Close and reopen
    const btn = page.locator('button[aria-label="Generate Training Plan"]')
    await btn.click()
    await btn.click()
    await expect(page.locator('.plan-error')).not.toBeVisible()
  })

  test('submitting with race type sends plan request and shows messages', async ({ page }) => {
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await page.locator('input[placeholder*="5K"]').fill('Half Marathon')
    await page.locator('.plan-submit').click()

    // Panel should close
    await expect(page.locator('.plan-input-panel')).not.toBeVisible()

    // User message should appear with structured prompt
    const userMsg = page.locator('.message.user .text')
    await expect(userMsg).toBeVisible()
    await expect(userMsg).toContainText('Generate a structured training plan')
    await expect(userMsg).toContainText('Half Marathon')

    // Assistant response should appear (from mock)
    const assistantMsg = page.locator('.message.assistant .markdown')
    await expect(assistantMsg).toBeVisible({ timeout: 5000 })
    await expect(assistantMsg).toContainText('Based on your recent training data')
  })

  test('submitting with only target time sends plan request', async ({ page }) => {
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await page.locator('input[placeholder*="3:30"]').fill('sub-20')
    await page.locator('.plan-submit').click()

    await expect(page.locator('.plan-input-panel')).not.toBeVisible()
    const userMsg = page.locator('.message.user .text')
    await expect(userMsg).toBeVisible()
    await expect(userMsg).toContainText('sub-20')
  })

  test('plan button is disabled while loading', async ({ page }) => {
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await page.locator('input[placeholder*="5K"]').fill('5K')
    await page.locator('.plan-submit').click()

    // During loading, the plan button should be disabled
    await expect(page.locator('button[aria-label="Generate Training Plan"]')).toBeDisabled()

    // Wait for response to complete
    await expect(page.locator('.message.assistant')).toBeVisible({ timeout: 5000 })
    // After loading, plan button re-enables
    await expect(page.locator('button[aria-label="Generate Training Plan"]')).toBeEnabled()
  })

  test('Generate Plan submit button has correct text', async ({ page }) => {
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await expect(page.locator('.plan-submit')).toContainText('Generate Plan')
  })

  test('fields reset after submitting a plan request', async ({ page }) => {
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await page.locator('input[placeholder*="5K"]').fill('Marathon')
    await page.locator('input[placeholder*="3:30"]').fill('3:30:00')
    await page.locator('.plan-submit').click()

    // Wait for response
    await expect(page.locator('.message.assistant')).toBeVisible({ timeout: 5000 })

    // Reopen panel — fields should be empty
    await page.locator('button[aria-label="Generate Training Plan"]').click()
    await expect(page.locator('input[placeholder*="5K"]')).toHaveValue('')
    await expect(page.locator('input[placeholder*="3:30"]')).toHaveValue('')
  })
})
