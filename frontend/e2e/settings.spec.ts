/**
 * E2E tests: settings tab (S16, S23, S26, S27)
 * Covers: backend selector, model fields, strava credentials, save button.
 */
import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Settings"]')
  // Wait for settings to finish loading
  await expect(page.locator('.settings')).toBeVisible()
})

test('renders the settings page', async ({ page }) => {
  await expect(page.locator('.settings')).toBeVisible()
})

test('AI Model section is visible', async ({ page }) => {
  await expect(page.locator('section h2').first()).toContainText('AI Model')
})

test('Ollama label is visible', async ({ page }) => {
  await expect(page.locator('.ollama-label')).toContainText('Ollama')
})

test('Ollama endpoint and model fields are visible', async ({ page }) => {
  await expect(page.locator('#ollama-endpoint')).toBeVisible()
  await expect(page.locator('#ollama-model')).toBeVisible()
})

test('No Claude or OpenAI fields exist', async ({ page }) => {
  await expect(page.locator('#claude-api-key')).not.toBeVisible()
  await expect(page.locator('#openai-api-key')).not.toBeVisible()
  await expect(page.locator('#active-backend')).not.toBeVisible()
})

test('Strava section is visible', async ({ page }) => {
  await expect(page.locator('section h2').nth(2)).toContainText('Strava Connection')
})

test('no Strava credential fields exist', async ({ page }) => {
  await expect(page.locator('#strava-client-id')).not.toBeVisible()
  await expect(page.locator('#strava-client-secret')).not.toBeVisible()
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
