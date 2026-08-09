import { defineConfig, devices } from '@playwright/test'

/**
 * E2E 测试（PRD V6 21.5 / V7 12.3，SAR IMP-01）
 * 关键路径：登录 → 主题切换 → 语言切换 → 版本检查提示
 * 运行前需确保后端 API (localhost:8090) 已启动并生成演示数据
 */
export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  fullyParallel: false,
  retries: 0,
  reporter: [['list']],
  use: {
    baseURL: 'http://127.0.0.1:3005',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://127.0.0.1:3005',
    reuseExistingServer: true,
    timeout: 120_000,
  },
})
