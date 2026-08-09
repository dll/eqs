import zhCN from '../locales/zh-CN.json'
import enUS from '../locales/en-US.json'
import { ref } from 'vue'

const messages: Record<string, Record<string, string>> = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

const currentLang = ref<string>(localStorage.getItem('admin-lang') || 'zh-CN')

export function setAdminLang(lang: string) {
  currentLang.value = lang in messages ? lang : 'zh-CN'
  localStorage.setItem('admin-lang', currentLang.value)
}

export function getAdminLang() {
  return currentLang.value
}

export function t(key: string, params?: Record<string, string | number>): string {
  const dict = messages[currentLang.value] || messages['zh-CN']
  let text = dict[key] || messages['zh-CN'][key] || key
  if (params) {
    Object.entries(params).forEach(([k, v]) => {
      text = text.replace(new RegExp(`\\{${k}\\}`, 'g'), String(v))
    })
  }
  return text
}

// Vue composable — 在 <script setup> 中使用
export function useI18n() {
  return { t, $t: t, lang: currentLang, setAdminLang }
}
