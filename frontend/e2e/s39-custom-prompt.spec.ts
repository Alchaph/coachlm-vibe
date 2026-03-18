/**
 * E2E tests: S39 — Custom system prompt (now in Context tab per S50)
 * Covers: textarea presence, entering text, saving, persistence on reload.
 */
import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Context"]')
  await expect(page.locator('.context')).toBeVisible()
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
  await page.click('button.btn-primary:has-text("Save Profile")')

  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })

  await page.click('button[title="Chat"]')
  await page.click('button[title="Context"]')
  await expect(page.locator('.context')).toBeVisible()

  await expect(page.locator('#custom-system-prompt')).toHaveValue('Always give pace in min/km.')
})

test('empty custom prompt does not break profile save', async ({ page }) => {
  await page.locator('#custom-system-prompt').fill('')
  await page.click('button.btn-primary:has-text("Save Profile")')
  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })
})

test('custom prompt field appears between Athlete Profile and Pinned Insights sections', async ({ page }) => {
  await expect(page.locator('section h2', { hasText: 'Athlete Profile' })).toBeVisible()
  await expect(page.locator('section h2', { hasText: 'Custom Instructions' })).toBeVisible()
  await expect(page.locator('section h2', { hasText: 'Pinned Insights' })).toBeVisible()

  const sections = page.locator('section h2')
  const texts = await sections.allTextContents()
  const profileIdx = texts.findIndex(t => t.includes('Athlete Profile'))
  const customIdx = texts.findIndex(t => t.includes('Custom Instructions'))
  const insightsIdx = texts.findIndex(t => t.includes('Pinned Insights'))

  expect(profileIdx).toBeGreaterThanOrEqual(0)
  expect(customIdx).toBeGreaterThan(profileIdx)
  expect(insightsIdx).toBeGreaterThan(customIdx)
})
