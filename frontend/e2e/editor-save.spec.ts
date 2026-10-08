import { test, expect } from '@playwright/test'

const API = 'http://localhost:8080'
const FRONTEND = 'http://localhost:5173'

const EMAIL = `test-${Date.now()}@example.com`
const PASSWORD = 'password123'

let accessToken: string
let refreshToken: string
let projectId: string
let fileId: string

async function registerAndLogin(request: any) {
  await request.post(`${API}/auth/register`, {
    data: { email: EMAIL, password: PASSWORD, display_name: 'tester' },
  })

  const loginRes = await request.post(`${API}/auth/login`, {
    data: { email: EMAIL, password: PASSWORD },
  })
  const loginData = await loginRes.json()
  accessToken = loginData.access_token
  refreshToken = loginData.refresh_token
}

test.beforeAll(async ({ request }) => {
  await registerAndLogin(request)

  const projectRes = await request.post(`${API}/projects`, {
    data: { name: 'Test Project' },
    headers: { Authorization: `Bearer ${accessToken}` },
  })
  const projectData = await projectRes.json()
  projectId = projectData.id

  const fileRes = await request.post(`${API}/projects/${projectId}/files`, {
    data: { name: 'main.js', path: '/main.js', language: 'javascript' },
    headers: { Authorization: `Bearer ${accessToken}` },
  })
  const fileData = await fileRes.json()
  fileId = fileData.id
})

async function loginAndSetStorage(page: any) {
  await page.goto(`${FRONTEND}/login`)
  await page.fill('input[type="email"]', EMAIL)
  await page.fill('input[type="password"]', PASSWORD)
  await page.click('button:has-text("Login")')
  await page.waitForURL(`${FRONTEND}/projects`, { timeout: 10000 })
}

test.describe('Editor save and restore', () => {
  test('saves and restores content_text after reload', async ({ page }) => {
    await loginAndSetStorage(page)
    await page.goto(`${FRONTEND}/projects/${projectId}/files/${fileId}`)

    await page.waitForSelector('.monaco-editor', { timeout: 10000 })
    await page.waitForFunction(() => !document.body.innerText.includes('Loading...'), { timeout: 10000 })

    const text = 'function hello() {\n  return 42;\n}'
    await page.click('.monaco-editor')
    await page.keyboard.type(text)

    await page.keyboard.press('Control+s')

    await page.waitForFunction(() => {
      const status = document.body.innerText
      return status.includes('Saved')
    }, { timeout: 5000 })

    const contentRes = await page.request.get(`${API}/files/${fileId}/content`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    const contentData = await contentRes.json()
    expect(contentData.content).toContain('function hello()')

    await page.reload()

    await page.waitForSelector('.monaco-editor', { timeout: 10000 })
    await page.waitForFunction(() => !document.body.innerText.includes('Loading...'), { timeout: 10000 })

    const editorText = await page.locator('.monaco-editor .view-lines').innerText()
    expect(editorText.replace(/\s+/g, ' ')).toContain('function hello()')
    expect(editorText.replace(/\s+/g, ' ')).toContain('return 42')
  })

  test('collaborative editing works in two tabs', async ({ browser }) => {
    const context1 = await browser.newContext()
    const context2 = await browser.newContext()

    const page1 = await context1.newPage()
    const page2 = await context2.newPage()

    await loginAndSetStorage(page1)
    await loginAndSetStorage(page2)

    await page1.goto(`${FRONTEND}/projects/${projectId}/files/${fileId}`)
    await page2.goto(`${FRONTEND}/projects/${projectId}/files/${fileId}`)

    await page1.waitForSelector('.monaco-editor', { timeout: 10000 })
    await page2.waitForSelector('.monaco-editor', { timeout: 10000 })

    await page1.waitForFunction(() => !document.body.innerText.includes('Loading...'))
    await page2.waitForFunction(() => !document.body.innerText.includes('Loading...'))

    await page1.click('.monaco-editor')
    await page1.keyboard.type('// from tab 1')

    await page1.waitForTimeout(1000)

    const text2 = await page2.locator('.monaco-editor .view-lines').innerText()
    expect(text2.replace(/\s+/g, ' ')).toContain('// from tab 1')

    await page2.click('.monaco-editor')
    await page2.keyboard.type('// from tab 2')

    await page2.waitForTimeout(1000)

    const text1 = await page1.locator('.monaco-editor .view-lines').innerText()
    expect(text1.replace(/\s+/g, ' ')).toContain('// from tab 2')

    await context1.close()
    await context2.close()
  })
})
