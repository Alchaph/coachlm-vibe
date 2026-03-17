/**
 * E2E tests: settings tab (S16, S23, S26, S27)
 * Covers: backend selector, model fields, strava credentials, save button.
 */
import { test, expect } from '@playwright/test'
import path from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: path.join(__dirname, 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Settings"]')
  // Wait for settings to finish loading
  await expect(page.locator('.settings select#active-backend')).toBeVisible()
})

test('renders the settings page', async ({ page }) => {
  await expect(page.locator('.settings')).toBeVisible()
})

test('LLM Backend section is visible', async ({ page }) => {
  await expect(page.locator('section h2').first()).toContainText('LLM Backend')
})

test('backend selector defaults to free', async ({ page }) => {
  await expect(page.locator('#active-backend')).toHaveValue('free')
})

test('selecting claude shows API key and model fields', async ({ page }) => {
  await page.selectOption('#active-backend', 'claude')
  await expect(page.locator('#claude-api-key')).toBeVisible()
  await expect(page.locator('#claude-model')).toBeVisible()
})

test('selecting openai shows API key and model fields', async ({ page }) => {
  await page.selectOption('#active-backend', 'openai')
  await expect(page.locator('#openai-api-key')).toBeVisible()
  await expect(page.locator('#openai-model')).toBeVisible()
})

test('selecting local shows ollama endpoint and model fields', async ({ page }) => {
  await page.selectOption('#active-backend', 'local')
  await expect(page.locator('#ollama-endpoint')).toBeVisible()
  await expect(page.locator('#ollama-model')).toBeVisible()
})

test('selecting free shows no-setup note', async ({ page }) => {
  await page.selectOption('#active-backend', 'free')
  await expect(page.locator('.field-note').first()).toContainText('No setup required')
})

test('Strava section is visible', async ({ page }) => {
  await expect(page.locator('section h2').nth(2)).toContainText('Strava Connection')
})

test('can enter Strava Client ID', async ({ page }) => {
  await page.fill('#strava-client-id', '12345')
  await expect(page.locator('#strava-client-id')).toHaveValue('12345')
})

test('save button exists and is clickable', async ({ page }) => {
  await expect(page.locator('button.save-btn')).toBeVisible()
  await expect(page.locator('button.save-btn')).toBeEnabled()
})

test('save settings shows success feedback', async ({ page }) => {
  await page.click('button.save-btn')
  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })
  await expect(page.locator('.feedback.success')).toContainText('Settings saved')
})

test('Context Data section is visible', async ({ page }) => {
  await expect(page.locator('section h2').nth(3)).toContainText('Context Data')
})

test('Export Context button is visible', async ({ page }) => {
  await expect(page.locator('button', { hasText: 'Export Context' })).toBeVisible()
})

test('Import Context button is visible', async ({ page }) => {
  await expect(page.locator('button', { hasText: 'Import Context' })).toBeVisible()
})
