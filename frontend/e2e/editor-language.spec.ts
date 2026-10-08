import { test, expect } from '@playwright/test'

const API = 'http://localhost:8080'
const FRONTEND = 'http://localhost:5173'

const EMAIL = `lang-test-${Date.now()}@example.com`
const PASSWORD = 'password123'

async function registerAndLogin(request: any) {
  await request.post(`${API}/auth/register`, {
    data: { email: EMAIL, password: PASSWORD, display_name: 'lang tester' },
  })

  const loginRes = await request.post(`${API}/auth/login`, {
    data: { email: EMAIL, password: PASSWORD },
  })
  const loginData = await loginRes.json()
  return loginData.access_token
}

async function loginAndSetStorage(page: any) {
  await page.goto(`${FRONTEND}/login`)
  await page.fill('input[type="email"]', EMAIL)
  await page.fill('input[type="password"]', PASSWORD)
  await page.click('button:has-text("Login")')
  await page.waitForURL(`${FRONTEND}/projects`, { timeout: 10000 })
}

test.describe('Editor language sync', () => {
  test('uses file language for syntax highlighting', async ({ request, page }) => {
    const accessToken = await registerAndLogin(request)

    const projectRes = await request.post(`${API}/projects`, {
      data: { name: 'Language Test Project' },
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    const projectData = await projectRes.json()
    const projectId = projectData.id

    const fileRes = await request.post(`${API}/projects/${projectId}/files`, {
      data: { name: 'main.go', path: '/main.go', language: 'go' },
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    const fileData = await fileRes.json()
    const fileId = fileData.id

    await loginAndSetStorage(page)
    await page.goto(`${FRONTEND}/projects/${projectId}/files/${fileId}`)

    await page.waitForSelector('.monaco-editor', { timeout: 10000 })

    const language = await page.evaluate(() => {
      const model = (window as any).monaco?.editor?.getModels()?.[0]
      return model?._languageId || model?.getLanguageId?.()
    })

    expect(language).toBe('go')
  })
})
