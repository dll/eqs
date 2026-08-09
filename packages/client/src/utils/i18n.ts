import zhCN from '../locales/zh-CN.json'
import enUS from '../locales/en-US.json'

const messages: Record<string, Record<string, string>> = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

let currentLang = 'zh-CN'

export function setI18nLang(lang: string) {
  currentLang = lang in messages ? lang : 'zh-CN'
}

export function t(key: string, params?: Record<string, string | number>): string {
  const dict = messages[currentLang] || messages['zh-CN']
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
  return { t, $t: t }
}
