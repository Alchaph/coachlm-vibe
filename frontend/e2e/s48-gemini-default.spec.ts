import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Settings"]')
  await expect(page.locator('.settings')).toBeVisible()
})

test('Ollama label is visible in settings', async ({ page }) => {
  await expect(page.locator('.ollama-label')).toContainText('Ollama')
})

test('No Claude or OpenAI fields exist in settings', async ({ page }) => {
  await expect(page.locator('#claude-api-key')).not.toBeVisible()
  await expect(page.locator('#openai-api-key')).not.toBeVisible()
})

test('No backend selector dropdown exists', async ({ page }) => {
  await expect(page.locator('#active-backend')).not.toBeVisible()
})

test('Saving settings succeeds', async ({ page }) => {
  await page.click('button.save-btn')
  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })
  await expect(page.locator('.feedback.success')).toContainText('Settings saved')
})

test('Ollama endpoint field is visible', async ({ page }) => {
  await expect(page.locator('#ollama-endpoint')).toBeVisible()
})

test('Ollama model field is visible', async ({ page }) => {
  await expect(page.locator('#ollama-model')).toBeVisible()
})
