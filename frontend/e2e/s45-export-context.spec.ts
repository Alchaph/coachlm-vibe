/**
 * E2E tests: S45 — Export context functionality (now in Context tab per S50)
 * Covers: export triggers save dialog, returns success on confirm, silent cancel on dialog dismiss.
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

test('Export Context button is visible', async ({ page }) => {
  await expect(page.locator('button', { hasText: 'Export Context' })).toBeVisible()
})

test('clicking Export Context shows success feedback when dialog returns path', async ({ page }) => {
  await page.locator('button', { hasText: 'Export Context' }).click()
  await expect(page.locator('.feedback.success')).toBeVisible({ timeout: 3000 })
  await expect(page.locator('.feedback.success')).toContainText('Context exported successfully')
})

test('cancelling export dialog shows no error', async ({ page }) => {
  await page.addInitScript(() => {
    window.go.main.App.ExportContextWithDialog = () => Promise.resolve(null)
  })
  await page.reload()
  await page.click('button[title="Context"]')
  await expect(page.locator('.context')).toBeVisible()

  await page.locator('button', { hasText: 'Export Context' }).click()

  await page.waitForTimeout(500)
  await expect(page.locator('.feedback.error')).not.toBeVisible()
})

test('export error is shown if ExportContextWithDialog fails', async ({ page }) => {
  await page.addInitScript(() => {
    window.go.main.App.ExportContextWithDialog = () => Promise.reject(new Error('disk full'))
  })
  await page.reload()
  await page.click('button[title="Context"]')
  await expect(page.locator('.context')).toBeVisible()

  await page.locator('button', { hasText: 'Export Context' }).click()
  await expect(page.locator('.feedback.error')).toBeVisible({ timeout: 3000 })
  await expect(page.locator('.feedback.error')).toContainText('disk full')
})
