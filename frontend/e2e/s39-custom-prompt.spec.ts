/**
 * E2E tests: S39 — Custom system prompt
 * Covers: textarea presence, entering text, saving, persistence on reload.
 */
import { test, expect } from '@playwright/test'
import path from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: path.join(__dirname, 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Settings"]')
  await expect(page.locator('.settings select#active-backend')).toBeVisible()
})

test('Custom Instructions section is visible', async ({ page }) => {
  await expect(page.locator('section h2', { hasText: 'Custom Instructions' })).toBeVisible()
})

test('custom system prompt textarea is present', async ({ page }) => {
  await expect(page.locator('#custom-system-prompt')).toBeVisible()
})

test('textarea has the correct placeholder text', async ({ page }) => {
  await expect(page.locator('#custom-system-prompt')).toHaveAttribute(
    'placeholder',
    /Add your own instructions/
  )
})

test('can type a custom prompt', async ({ page }) => {
  const textarea = page.locator('#custom-system-prompt')
  await textarea.fill('Always respond in German. Never suggest supplements.')
  await expect(textarea).toHaveValue('Always respond in German. Never suggest supplements.')
})

test('custom prompt is saved and persists on reload', async ({ page }) => {
  const textarea = page.locator('#custom-system-prompt')
  await textarea.fill('Always give pace in min/km.')
  await page.click('button.save-btn')

  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })

  // Reload settings page — mock state persists in the same page context
  await page.click('button[title="Chat"]')
  await page.click('button[title="Settings"]')
  await expect(page.locator('.settings select#active-backend')).toBeVisible()

  // Value should be what was saved to __WAILS_MOCK_STATE__
  await expect(page.locator('#custom-system-prompt')).toHaveValue('Always give pace in min/km.')
})

test('empty custom prompt does not break settings save', async ({ page }) => {
  // Ensure textarea is empty
  await page.locator('#custom-system-prompt').fill('')
  await page.click('button.save-btn')
  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })
})

test('custom prompt field appears between LLM Backend and Strava sections', async ({ page }) => {
  const sections = page.locator('section h2')
  const texts = await sections.allTextContents()
  const llmIdx = texts.findIndex(t => t.includes('LLM Backend'))
  const customIdx = texts.findIndex(t => t.includes('Custom Instructions'))
  const stravaIdx = texts.findIndex(t => t.includes('Strava Connection'))

  expect(llmIdx).toBeGreaterThanOrEqual(0)
  expect(customIdx).toBeGreaterThan(llmIdx)
  expect(stravaIdx).toBeGreaterThan(customIdx)
})
