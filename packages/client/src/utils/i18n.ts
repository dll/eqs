import zhCN from '../locales/zh-CN.json'
import enUS from '../locales/en-US.json'
import { ref, watch } from 'vue'

const messages: Record<string, Record<string, string>> = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

const currentLang = ref<string>('zh-CN')

export function setI18nLang(lang: string) {
  currentLang.value = lang in messages ? lang : 'zh-CN'
}

export function getI18nLang(): string {
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

// 同步 tabBar 文案（uni-app H5/App 静态 tabBar 需 API 更新）
const TAB_ITEMS = [
  { index: 0, text: 'nav.home' },
  { index: 1, text: 'nav.project' },
  { index: 2, text: 'nav.order' },
  { index: 3, text: 'nav.mine' },
]

export function applyTabBarI18n() {
  try {
    TAB_ITEMS.forEach((item) => uni.setTabBarItem({ index: item.index, text: t(item.text) }))
  } catch {
    // 运行环境不支持时忽略
  }
}

export function applyNavTitle(titleKey: string) {
  try {
    uni.setNavigationBarTitle({ title: t(titleKey) })
  } catch {
    // 忽略
  }
}

interface PageNavigationHooks {
  onLoad?: (callback: () => void) => void
  onShow?: (callback: () => void) => void
}

// 统一处理页面导航栏标题：进入页面时设置，语言切换时自动刷新
export function usePageTitle(titleKey: string, hooks?: PageNavigationHooks) {
  const apply = () => applyNavTitle(titleKey)
  if (hooks?.onShow) {
    hooks.onShow(apply)
  } else if (hooks?.onLoad) {
    hooks.onLoad(apply)
  }
  watch(currentLang, apply)
  return { apply }
}

// Vue composable — 在 <script setup> 中使用
export function useI18n() {
  return { t, $t: t, lang: currentLang }
}