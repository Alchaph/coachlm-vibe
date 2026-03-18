import { dirname, join } from 'path'
import { fileURLToPath } from 'url'
import { test, expect } from '@playwright/test'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

test.describe('Cloud Sync Settings', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript({ path: join(__dirname, 'mocks', 'wails.ts') })
    await page.goto('/')
    await page.click('button[title="Settings"]')
    await expect(page.locator('.settings')).toBeVisible()
  })

  test('Cloud sync section visible in settings', async ({ page }) => {
    const section = page.locator('section', { hasText: 'Cloud Sync' })
    await expect(section.locator('h2', { hasText: 'Cloud Sync' })).toBeVisible()
    await expect(section.locator('.status-badge', { hasText: 'Not Connected' })).toBeVisible()
  })

  test('Provider selection shows correct fields', async ({ page }) => {
    await expect(page.locator('label', { hasText: 'Endpoint URL' })).toBeVisible()
    await expect(page.locator('label', { hasText: 'Bucket Name' })).toBeVisible()
    await expect(page.locator('label', { hasText: 'Access Key' })).toBeVisible()
    await expect(page.locator('label', { hasText: 'Secret Key' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'Connect S3' })).toBeVisible()

    await page.selectOption('select#cloud-provider', 'Google Drive')
    await expect(page.locator('label', { hasText: 'Endpoint URL' })).not.toBeVisible()
    await expect(page.locator('button', { hasText: 'Connect Google Drive' })).toBeVisible()
  })

  test('S3 connect form validation', async ({ page }) => {
    await page.click('button:has-text("Connect S3")')
    await expect(page.locator('.feedback.error')).toHaveText('Please fill in all S3 fields')
  })

  test('Connected state shows correct UI', async ({ page }) => {
    await page.fill('#s3-endpoint', 'https://s3.example.com')
    await page.fill('#s3-bucket', 'my-bucket')
    await page.fill('#s3-access-key', 'AKIA123')
    await page.fill('#s3-secret-key', 'secret123')

    await page.click('button:has-text("Connect S3")')
    await expect(page.locator('.feedback.success')).toHaveText('Connected to S3 successfully')

    await expect(page.locator('.status-badge.connected', { hasText: 'Connected' })).toBeVisible()
    await expect(page.locator('text=Provider: S3')).toBeVisible()
    await expect(page.locator('button', { hasText: 'Sync Now' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'Disconnect' })).toBeVisible()
  })

  test('Disconnect clears state', async ({ page }) => {
    await page.selectOption('select#cloud-provider', 'Google Drive')
    await page.click('button:has-text("Connect Google Drive")')
    await expect(page.locator('.feedback.success')).toHaveText('Connected to Google Drive successfully')

    page.on('dialog', dialog => dialog.accept())
    await page.click('button:has-text("Disconnect")')
    await expect(page.locator('.feedback.success')).toHaveText('Cloud Sync disconnected')

    const section = page.locator('section', { hasText: 'Cloud Sync' })
    await expect(section.locator('.status-badge', { hasText: 'Not Connected' })).toBeVisible()
    await expect(page.locator('select#cloud-provider')).toBeVisible()
  })
})
