/**
 * E2E tests: sidebar navigation (tab switching)
 * Covers the core shell navigation present in App.svelte.
 */
import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
  await page.goto('/')
})

test('loads with chat tab active by default', async ({ page }) => {
  // Chat input area should be visible
  await expect(page.locator('.input-area textarea')).toBeVisible()
})

test('switches to Dashboard tab', async ({ page }) => {
  await page.click('button[title="Dashboard"]')
  await expect(page.locator('.dashboard')).toBeVisible()
})

test('switches to Context tab', async ({ page }) => {
  await page.click('button[title="Context"]')
  await expect(page.locator('.context')).toBeVisible()
})

test('switches to Settings tab', async ({ page }) => {
  await page.click('button[title="Settings"]')
  await expect(page.locator('.settings')).toBeVisible()
})

test('returns to Chat tab from another tab', async ({ page }) => {
  await page.click('button[title="Dashboard"]')
  await expect(page.locator('.dashboard')).toBeVisible()
  await page.click('button[title="Chat"]')
  await expect(page.locator('.input-area textarea')).toBeVisible()
})
