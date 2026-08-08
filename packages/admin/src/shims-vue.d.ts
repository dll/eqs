// 修复 element-plus 中文 locale 缺少类型声明
declare module 'element-plus/dist/locale/zh-cn.mjs' {
  import type { Language } from 'element-plus/es/locale'
  const zhCn: Language
  export default zhCn
}

// vue-tsc 下对 *.vue 的默认类型（补充开关）
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}