import { defineStore } from 'pinia'
import { ref } from 'vue'
import { request } from '@/utils/request'

export const THEMES = [
  { id: 'print', name: '打印主题', description: '白底黑字，适合截图打印' },
  { id: 'dark', name: '深色主题', description: '深色背景，夜间友好' },
  { id: 'light', name: '浅色主题', description: '标准浅色界面' },
] as const

export const LANGS = [
  { id: 'zh-CN', name: '中文' },
  { id: 'en-US', name: 'English' },
] as const

export const useSettingsStore = defineStore('settings', () => {
  const theme = ref<string>('print')
  const lang = ref<string>('zh-CN')
  const publicConfigs = ref<Record<string, any>>({})
  const messages = ref<Record<string, string>>({})
  const updateAvailable = ref(false)
  const latestVersion = ref('')

  const loadSettings = async () => {
    try {
      const [cfgRes, prefsRes] = await Promise.all([
        request.get('/api/v1/config/public'),
        request.get('/api/v1/config/user/prefs'),
      ])
      publicConfigs.value = cfgRes.configs || {}
      theme.value = prefsRes.theme || publicConfigs.value['theme.default'] || 'print'
      lang.value = prefsRes.lang || 'zh-CN'
      applyTheme(theme.value)
      await loadMessages()
    } catch {
      applyTheme('print')
    }
  }

  const loadMessages = async () => {
    try {
      const res = await request.get(`/api/v1/i18n/${lang.value}`)
      messages.value = res.messages || {}
    } catch {
      messages.value = {}
    }
  }

  const setTheme = async (t: string) => {
    theme.value = t
    applyTheme(t)
    try {
      await request.put('/api/v1/config/user/prefs', { theme: t })
    } catch {
      // 请求失败仅本地生效
    }
  }

  const setLang = async (l: string) => {
    lang.value = l
    try {
      await request.put('/api/v1/config/user/prefs', { lang: l })
    } catch {
      // 请求失败仅本地生效
    }
    await loadMessages()
  }

  const applyTheme = (t: string) => {
    const vars: Record<string, Record<string, string>> = {
      print: {
        '--bg-color': '#ffffff',
        '--text-color': '#000000',
        '--card-bg': '#ffffff',
        '--border-color': '#e5e5e5',
        '--muted-color': '#666666',
        '--primary-color': '#1890ff',
      },
      dark: {
        '--bg-color': '#1e1e1e',
        '--text-color': '#eeeeee',
        '--card-bg': '#2a2a2a',
        '--border-color': '#444444',
        '--muted-color': '#999999',
        '--primary-color': '#4d9fff',
      },
      light: {
        '--bg-color': '#f5f5f5',
        '--text-color': '#333333',
        '--card-bg': '#ffffff',
        '--border-color': '#e5e5e5',
        '--muted-color': '#666666',
        '--primary-color': '#1890ff',
      },
    }
    const themeVars = vars[t] || vars.print
    const root = (document.documentElement || document.body) as HTMLElement
    Object.entries(themeVars).forEach(([key, value]) => {
      root.style.setProperty(key, value)
    })
    const page = document.querySelector('page') as HTMLElement | null
    if (page) {
      Object.entries(themeVars).forEach(([key, value]) => {
        page.style.setProperty(key, value)
      })
    }
  }

  const checkVersion = async () => {
    try {
      const res = await request.get('/api/v1/version/check?current=1.0.0&platform=h5')
      updateAvailable.value = !!res.update_available
      latestVersion.value = res.version || ''
      return res
    } catch {
      return null
    }
  }

  return {
    theme, lang, publicConfigs, messages, updateAvailable, latestVersion,
    loadSettings, setTheme, setLang, checkVersion, applyTheme,
  }
})
