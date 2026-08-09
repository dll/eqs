<template>
  <view class="container">
    <view class="user-card" v-if="userStore.user">
      <text class="user-name">{{ userStore.user.company_name || $t('mine.company') }}</text>
      <text class="user-phone">{{ userStore.user.phone }}</text>
      <text class="user-score">{{ $t('mine.creditScore') }}：{{ userStore.user.credit_score }}</text>
    </view>

    <view class="menu-list">
      <view class="menu-item" @tap="goTo('/pages/order/list')">
        <text>{{ $t('mine.myOrders') }}</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/project/list')">
        <text>{{ $t('mine.myProjects') }}</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="showThemePicker">
        <text>{{ $t('mine.theme') }}（{{ themeName }}）</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="showLangPicker">
        <text>{{ $t('mine.lang') }}（{{ langName }}）</text>
        <text class="arrow">></text>
      </view>
      <view class="menu-item" @tap="logout">
        <text>{{ $t('mine.logout') }}</text>
        <text class="arrow">></text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useUserStore } from '@/store/user'
import { useSettingsStore, THEMES, LANGS } from '@/store/settings'
import { useI18n } from '@/utils/i18n'

const userStore = useUserStore()
const settingsStore = useSettingsStore()
const { $t } = useI18n()

const themeName = computed(() => THEMES.find(t => t.id === settingsStore.theme)?.name || settingsStore.theme)
const langName = computed(() => LANGS.find(l => l.id === settingsStore.lang)?.name || settingsStore.lang)

const goTo = (url: string) => {
  uni.navigateTo({ url })
}

const showThemePicker = () => {
  uni.showActionSheet({
    itemList: THEMES.map(t => t.name),
    success: (res: any) => {
      const theme = THEMES[res.tapIndex]
      if (theme) settingsStore.setTheme(theme.id)
    },
  })
}

const showLangPicker = () => {
  uni.showActionSheet({
    itemList: LANGS.map(l => l.name),
    success: (res: any) => {
      const lang = LANGS[res.tapIndex]
      if (lang) settingsStore.setLang(lang.id)
    },
  })
}

const logout = () => {
  userStore.logout()
  uni.reLaunch({ url: '/pages/login/index' })
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.user-card {
  background: #1890ff;
  border-radius: 10rpx;
  padding: 40rpx 30rpx;
  margin-bottom: 30rpx;
  color: #fff;
}

.user-name {
  font-size: 36rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 10rpx;
}

.user-phone {
  font-size: 28rpx;
  display: block;
  margin-bottom: 10rpx;
  opacity: 0.9;
}

.user-score {
  font-size: 26rpx;
  opacity: 0.8;
}

.menu-list {
  background: #fff;
  border-radius: 10rpx;
}

.menu-item {
  display: flex;
  justify-content: space-between;
  padding: 30rpx;
  border-bottom: 1rpx solid #eee;
  font-size: 30rpx;
  color: #333;
}

.arrow {
  color: #999;
}
</style>
