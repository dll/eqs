import { defineStore } from 'pinia'
import { ref } from 'vue'
import { request } from '@/utils/request'
import { setI18nLang, applyTabBarI18n } from '@/utils/i18n'

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
  const updateAvailable = ref(false)
  const latestVersion = ref('')
  const latestVersionNotes = ref('')
  const latestVersionMandatory = ref(false)
  const latestUpdateUrl = ref('')

  const loadSettings = async () => {
    try {
      const [cfgRes, prefsRes] = await Promise.all([
        request.get('/api/v1/config/public', { silent401: true }).catch(() => ({ configs: {} })),
        request.get('/api/v1/config/user/prefs', { silent401: true }).catch(() => ({ theme: 'print', lang: 'zh-CN' })),
      ])
      publicConfigs.value = cfgRes.configs || {}
      theme.value = prefsRes.theme || publicConfigs.value['theme.default'] || 'print'
      lang.value = prefsRes.lang || 'zh-CN'
      setI18nLang(lang.value)
      applyTheme(theme.value)
    } catch {
      setI18nLang('zh-CN')
      applyTheme('print')
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
    setI18nLang(l)
    applyTabBarI18n()
    try {
      await request.put('/api/v1/config/user/prefs', { lang: l })
    } catch {
      // 请求失败仅本地生效
    }
  }

  const applyTheme = (t: string) => {
    // 统一品牌色：工程蓝 #2563EB / 科技青 #06B6D4 / AI 紫 #8B5CF6（与 Admin 设计系统一致）
    const vars: Record<string, Record<string, string>> = {
      print: {
        '--bg-color': '#f4f6fb',
        '--text-color': '#1e293b',
        '--card-bg': '#ffffff',
        '--border-color': '#e6eaf3',
        '--muted-color': '#64748b',
        '--primary-color': '#2563eb',
        '--input-bg': '#f1f5f9',
      },
      dark: {
        '--bg-color': '#0f172a',
        '--text-color': '#e2e8f0',
        '--card-bg': '#1e293b',
        '--border-color': '#334155',
        '--muted-color': '#94a3b8',
        '--primary-color': '#60a5fa',
        '--input-bg': '#1e293b',
      },
      light: {
        '--bg-color': '#f4f6fb',
        '--text-color': '#1e293b',
        '--card-bg': '#ffffff',
        '--border-color': '#e6eaf3',
        '--muted-color': '#64748b',
        '--primary-color': '#2563eb',
        '--input-bg': '#f1f5f9',
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
      // 检查间隔（小时），默认 6 小时，来自 version.checkInterval 配置
      const intervalHours = Number(publicConfigs.value['version.checkInterval']) || 6
      const lastCheck = uni.getStorageSync('lastVersionCheck')
      if (lastCheck && Date.now() - Number(lastCheck) < intervalHours * 3600 * 1000) {
        return null
      }
      const res = await request.get('/api/v1/version/check?current=1.0.0&platform=h5', { silent401: true })
      uni.setStorageSync('lastVersionCheck', String(Date.now()))
      updateAvailable.value = !!res.update_available
      latestVersion.value = res.version || ''
      latestVersionNotes.value = res.release_notes || ''
      latestVersionMandatory.value = !!res.mandatory
      latestUpdateUrl.value = res.update_url || ''
      return res
    } catch {
      return null
    }
  }

  // applyProjectTheme 应用项目维度主题（覆盖用户主题）
  const applyProjectTheme = (projectTheme: string) => {
    if (projectTheme && (projectTheme === 'print' || projectTheme === 'dark' || projectTheme === 'light')) {
      applyTheme(projectTheme)
    } else {
      applyTheme(theme.value)
    }
  }

  return {
    theme, lang, publicConfigs, updateAvailable, latestVersion,
    latestVersionNotes, latestVersionMandatory, latestUpdateUrl,
    loadSettings, setTheme, setLang, checkVersion, applyTheme, applyProjectTheme,
  }
})
