import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const mocks = vi.hoisted(() => ({
  request: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

vi.mock('@/utils/request', () => ({
  request: mocks.request,
}))

import { useSettingsStore, THEMES, LANGS } from './settings'


// node 环境下 mock document（settings 使用 CSS 变量）
;(globalThis as any).document = {
  documentElement: { style: { setProperty: () => {} } },
  querySelector: () => null,
}

describe('settings store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mocks.request.get.mockReset()
    mocks.request.put.mockReset()
    ;(document as any).documentElement.style.cssText = ''
  })

  it('主题定义包含打印/深色/浅色', () => {
    expect(THEMES.map(t => t.id)).toEqual(['print', 'dark', 'light'])
    expect(LANGS.map(l => l.id)).toEqual(['zh-CN', 'en-US'])
  })

  it('loadSettings 加载配置与偏好并应用主题', async () => {
    const store = useSettingsStore()
    mocks.request.get
      .mockResolvedValueOnce({ configs: { 'theme.default': 'light' } })
      .mockResolvedValueOnce({ theme: 'dark', lang: 'en-US' })
      .mockResolvedValueOnce({ messages: { 'nav.home': 'Home' } })
    await store.loadSettings()
    expect(store.theme).toBe('dark')
    expect(store.lang).toBe('en-US')
    expect(store.messages['nav.home']).toBe('Home')
  })

  it('setTheme 应用 CSS 变量并同步后端', async () => {
    const store = useSettingsStore()
    mocks.request.put.mockResolvedValue({})
    await store.setTheme('dark')
    expect(store.theme).toBe('dark')
    expect(mocks.request.put).toHaveBeenCalledWith('/api/v1/config/user/prefs', { theme: 'dark' })
  })

  it('setLang 更新语言并重新加载文案', async () => {
    const store = useSettingsStore()
    mocks.request.put.mockResolvedValue({})
    mocks.request.get.mockResolvedValue({ messages: { 'nav.home': '首页' } })
    await store.setLang('zh-CN')
    expect(store.lang).toBe('zh-CN')
    expect(store.messages['nav.home']).toBe('首页')
  })

  it('checkVersion 检测新版本', async () => {
    const store = useSettingsStore()
    mocks.request.get.mockResolvedValue({ update_available: true, version: '1.1.0' })
    const res = await store.checkVersion()
    expect(res.update_available).toBe(true)
    expect(store.updateAvailable).toBe(true)
    expect(store.latestVersion).toBe('1.1.0')
  })

  it('applyTheme 设置打印主题变量', () => {
    const calls: Record<string, string> = {}
    ;(globalThis as any).document.documentElement.style = {
      setProperty: (k: string, v: string) => { calls[k] = v },
    }
    const store = useSettingsStore()
    store.applyTheme('print')
    expect(calls['--bg-color']).toBe('#ffffff')
    expect(calls['--text-color']).toBe('#000000')
    expect(calls['--primary-color']).toBe('#1890ff')
  })
})
