import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Settings"]')
  await expect(page.locator('.settings')).toBeVisible()
})

test('Gemini 2.0 Flash label is visible in settings', async ({ page }) => {
  await expect(page.locator('.gemini-label')).toContainText('Gemini 2.0 Flash')
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

test('Advanced Ollama section is collapsible', async ({ page }) => {
  const summary = page.locator('details.advanced-section summary')
  await expect(summary).toBeVisible()
  
  await expect(page.locator('details.advanced-section')).not.toHaveAttribute('open', '')
  
  await summary.click()
  await expect(page.locator('details.advanced-section')).toHaveAttribute('open', '')
})
