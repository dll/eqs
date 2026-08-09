import { test, expect, type Page } from '@playwright/test'

/**
 * H5 端关键用户路径 E2E（PRD V7 12.3）
 * 登录 → 主题切换 → 语言切换 → 版本检查提示
 * 依赖：后端 8090 已启动且已 seed 演示数据
 */

// uni-app H5 将 <button>/<input> 渲染为 uni-button/uni-input 自定义元素
async function doLogin(page: Page, phone = '13900001111') {
  await page.goto('/#/pages/login/index')
  await page.waitForSelector('uni-input', { timeout: 15000 })
  await page.locator('uni-input input').first().fill(phone)
  await page.locator('uni-input input').nth(1).fill('123456')
  await page.locator('.submit-btn').click()
  // 登录成功后回到首页（uni-app 默认路由渲染为 #/），等待首页内容渲染
  await page.waitForSelector('text=推荐项目', { timeout: 15000 })
}

test.describe('H5 关键路径', () => {
  test('登录页可访问并完成登录', async ({ page }) => {
    await page.goto('/#/pages/login/index')
    await expect(page.locator('uni-input').first()).toBeVisible()
    await doLogin(page)
  })

  test('登录页可切换中英文', async ({ page }) => {
    await page.goto('/#/pages/login/index')
    await page.waitForSelector('uni-input', { timeout: 15000 })
    await page.click('.lang-btn:has-text("EN")')
    await expect(page.locator('text=Phone number')).toBeVisible()
    await page.click('.lang-btn:has-text("中文")')
    await expect(page.locator('text=请输入手机号')).toBeVisible()
  })

  test('用户中心可切换主题', async ({ page }) => {
    await doLogin(page)
    // 进入"我的" tab
    await page.click('.uni-tabbar >> text=我的')
    await page.waitForTimeout(1000)
    await expect(page.locator('text=信用分')).toBeVisible()
    // 打开主题选择
    await page.click('text=主题')
    await page.waitForTimeout(500)
    // 选择深色主题
    await page.click('text=深色主题')
    await page.waitForTimeout(500)
    // 校验 CSS 变量已应用
    const bg = await page.evaluate(() => document.documentElement.style.getPropertyValue('--bg-color'))
    expect(bg).toBe('#1e1e1e')
  })
})
