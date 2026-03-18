/**
 * E2E tests: S45 — Export context functionality
 * Covers: export triggers save dialog, returns success on confirm, silent cancel on dialog dismiss.
 */
import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
  await page.goto('/')
  await page.click('button[title="Settings"]')
  await expect(page.locator('.settings')).toBeVisible()
})

test('Export Context button is visible', async ({ page }) => {
  await expect(page.locator('button', { hasText: 'Export Context' })).toBeVisible()
})

test('clicking Export Context shows success feedback when dialog returns path', async ({ page }) => {
  // Default mock already returns '/tmp/mock-export.coachctx' from DialogSaveFile
  await page.locator('button', { hasText: 'Export Context' }).click()
  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })
  await expect(page.locator('.feedback.success')).toContainText('Context exported successfully')
})

test('cancelling export dialog shows no error', async ({ page }) => {
  // Override DialogSaveFile to return null (simulating cancel)
  await page.addInitScript(() => {
    window.runtime.DialogSaveFile = () => Promise.resolve(null)
  })
  await page.reload()
  await page.click('button[title="Settings"]')
  await expect(page.locator('.settings')).toBeVisible()

  await page.locator('button', { hasText: 'Export Context' }).click()

  // No feedback should appear (neither success nor error) for a cancel
  await page.waitForTimeout(500)
  await expect(page.locator('.feedback')).not.toBeVisible()
})

test('export error is shown if ExportContext fails', async ({ page }) => {
  await page.addInitScript(() => {
    window.go.main.App.ExportContext = () => Promise.reject(new Error('disk full'))
  })
  await page.reload()
  await page.click('button[title="Settings"]')
  await expect(page.locator('.settings')).toBeVisible()

  await page.locator('button', { hasText: 'Export Context' }).click()
  await expect(page.locator('.feedback.error')).toBeVisible({ timeout: 3000 })
  await expect(page.locator('.feedback.error')).toContainText('disk full')
})
