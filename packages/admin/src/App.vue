<template>
  <el-config-provider :locale="elementLocale">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
// @ts-expect-error element-plus 未提供 en.mjs 的类型声明
import en from 'element-plus/dist/locale/en.mjs'
import { useI18n } from '@/utils/i18n'

const { lang, t } = useI18n()

const elementLocale = computed(() => (lang.value === 'en-US' ? en : zhCn))

watch(
  () => lang.value,
  () => {
    document.title = `${t('app.title')} - ${t('login.title')}`
  }
)
</script>

<style>
body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}
</style>