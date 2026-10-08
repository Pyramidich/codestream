import { test, expect } from '@playwright/test'

const API = 'http://localhost:8080'
const FRONTEND = 'http://localhost:5173'

const ownerEmail = `owner-${Date.now()}@example.com`
const memberEmail = `member-${Date.now()}@example.com`
const password = 'password123'

async function registerAndLogin(request: any, email: string) {
  await request.post(`${API}/auth/register`, {
    data: { email, password, display_name: email.split('@')[0] },
  })

  const loginRes = await request.post(`${API}/auth/login`, {
    data: { email, password },
  })
  const loginData = await loginRes.json()
  return loginData.access_token
}

async function loginAndSetStorage(page: any, email: string) {
  await page.goto(`${FRONTEND}/login`)
  await page.fill('input[type="email"]', email)
  await page.fill('input[type="password"]', password)
  await page.click('button:has-text("Login")')
  await page.waitForURL(`${FRONTEND}/projects`, { timeout: 10000 })
}

test.describe('Project members management', () => {
  test('owner can invite a member by email and then remove them', async ({ request, browser }) => {
    const ownerToken = await registerAndLogin(request, ownerEmail)
    const memberToken = await registerAndLogin(request, memberEmail)

    const projectRes = await request.post(`${API}/projects`, {
      data: { name: 'Member Test Project' },
      headers: { Authorization: `Bearer ${ownerToken}` },
    })
    const projectData = await projectRes.json()
    const projectId = projectData.id

    const ownerContext = await browser.newContext()
    const ownerPage = await ownerContext.newPage()
    await loginAndSetStorage(ownerPage, ownerEmail)

    await ownerPage.goto(`${FRONTEND}/projects/${projectId}/settings`)
    await ownerPage.waitForSelector('text=Members', { timeout: 10000 })

    await ownerPage.fill('input[type="email"]', memberEmail)
    await ownerPage.selectOption('select', 'editor')
    await ownerPage.click('button:has-text("Add")')

    await ownerPage.waitForTimeout(1000)

    const pageText = await ownerPage.locator('body').innerText()
    console.log('Page body after add:', pageText)

    if (pageText.includes(memberEmail)) {
      console.log('Member email found in page body')
    } else if (pageText.includes('Failed to add member')) {
      throw new Error(`Failed to add member: ${pageText}`)
    }

    const memberContext = await browser.newContext()
    const memberPage = await memberContext.newPage()
    await loginAndSetStorage(memberPage, memberEmail)

    await memberPage.goto(`${FRONTEND}/projects`)
    await expect(memberPage.locator(`text=Member Test Project`)).toBeVisible()

    await ownerPage.click('button:has-text("Remove")')
    await ownerPage.waitForTimeout(500)

    await memberPage.goto(`${FRONTEND}/projects`)
    await expect(memberPage.locator(`text=Member Test Project`)).not.toBeVisible()

    await ownerContext.close()
    await memberContext.close()
  })
})
