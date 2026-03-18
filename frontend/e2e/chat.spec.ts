/**
 * E2E tests: chat tab (S12)
 * Covers: empty state, typing, sending a message, receiving a response, pin insight.
 */
import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

test.beforeEach(async ({ page }) => {
  await page.addInitScript({ path: join(dirname(fileURLToPath(import.meta.url)), 'mocks/wails.ts') })
  await page.goto('/')
})

test('shows empty state when no messages', async ({ page }) => {
  await expect(page.locator('.empty-state')).toBeVisible()
  await expect(page.locator('.empty-state h2')).toHaveText('CoachLM')
})

test('send button is disabled when input is empty', async ({ page }) => {
  await expect(page.locator('button.send-btn')).toBeDisabled()
})

test('send button is enabled after typing', async ({ page }) => {
  await page.locator('.input-area textarea').fill('How fast should I run tomorrow?')
  await expect(page.locator('button.send-btn')).toBeEnabled()
})

test('sends a message and receives a response', async ({ page }) => {
  const textarea = page.locator('.input-area textarea')
  await textarea.fill('How fast should I run tomorrow?')
  await page.click('button.send-btn')

  // User message should appear
  await expect(page.locator('.message.user .text')).toHaveText('How fast should I run tomorrow?')

  // Assistant response should appear (from mock)
  await expect(page.locator('.message.assistant .markdown')).toBeVisible({ timeout: 5000 })
})

test('input clears after sending', async ({ page }) => {
  const textarea = page.locator('.input-area textarea')
  await textarea.fill('What pace for a tempo run?')
  await page.click('button.send-btn')
  await expect(textarea).toHaveValue('')
})

test('Enter key submits message', async ({ page }) => {
  const textarea = page.locator('.input-area textarea')
  await textarea.fill('What is my weekly mileage target?')
  await textarea.press('Enter')

  await expect(page.locator('.message.user .text')).toHaveText('What is my weekly mileage target?')
})

test('Shift+Enter does not submit (adds newline)', async ({ page }) => {
  const textarea = page.locator('.input-area textarea')
  await textarea.fill('line one')
  await textarea.press('Shift+Enter')
  // No messages should have been sent
  await expect(page.locator('.message.user')).toHaveCount(0)
})

test('assistant response has pin button', async ({ page }) => {
  const textarea = page.locator('.input-area textarea')
  await textarea.fill('Suggest a workout')
  await page.click('button.send-btn')

  // Wait for assistant message
  const assistantMsg = page.locator('.message.assistant').first()
  await expect(assistantMsg).toBeVisible({ timeout: 5000 })

  // Hover over bubble to reveal pin button
  await assistantMsg.locator('.message-bubble').hover()
  await expect(assistantMsg.locator('.pin-btn')).toBeVisible()
})

test('pin button saves insight and shows feedback', async ({ page }) => {
  const textarea = page.locator('.input-area textarea')
  await textarea.fill('Suggest a workout')
  await page.click('button.send-btn')

  const assistantBubble = page.locator('.message.assistant .message-bubble').first()
  await expect(assistantBubble).toBeVisible({ timeout: 5000 })

  await assistantBubble.hover()
  await assistantBubble.locator('.pin-btn').click()

  // Pin feedback should appear briefly
  await expect(assistantBubble.locator('.pin-feedback')).toBeVisible({ timeout: 2000 })
})
