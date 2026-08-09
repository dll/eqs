<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { useUserStore } from '@/store/user'
import { useSettingsStore } from '@/store/settings'
import { useI18n, applyTabBarI18n } from '@/utils/i18n'

const { $t } = useI18n()

const showUpdateDialog = ref(false)
const updateInfo = ref({ version: '', notes: '', mandatory: false, url: '' })

onLaunch(async () => {
  const userStore = useUserStore()
  const settingsStore = useSettingsStore()
  await userStore.loadUser()
  await settingsStore.loadSettings()
  applyTabBarI18n()
  const res = await settingsStore.checkVersion()
  if (res?.update_available) {
    updateInfo.value = {
      version: res.version || '',
      notes: res.release_notes || '',
      mandatory: !!res.mandatory,
      url: res.update_url || '',
    }
    showUpdateDialog.value = true
  }
})

const doUpdate = () => {
  showUpdateDialog.value = false
  if (updateInfo.value.url) {
    window.open(updateInfo.value.url, '_blank')
  } else {
    // H5: 刷新并清缓存
    if ('caches' in window) {
      caches.keys().then(names => {
        names.forEach(name => caches.delete(name))
      })
    }
    window.location.reload()
  }
}

const dismissUpdate = () => {
  showUpdateDialog.value = false
}
</script>

<style>
page {
  background-color: var(--bg-color, #f5f5f5);
  color: var(--text-color, #333333);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* 打印媒体查询 — 打印主题白底黑字 */
@media print {
  page, body, html {
    background: #ffffff !important;
    color: #000000 !important;
  }
  .tabbar,
  .action-bar,
  .bid-modal,
  .lang-switch {
    display: none !important;
  }
  * {
    background: transparent !important;
    color: #000000 !important;
    box-shadow: none !important;
  }
}

/* 版本更新弹窗 */
.update-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}
.update-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 40rpx;
  width: 80%;
  max-width: 600rpx;
}
.update-title {
  font-size: 34rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 20rpx;
  text-align: center;
}
.update-version {
  font-size: 28rpx;
  color: #1890ff;
  display: block;
  text-align: center;
  margin-bottom: 16rpx;
}
.update-notes {
  font-size: 26rpx;
  color: #666;
  display: block;
  margin-bottom: 30rpx;
  line-height: 1.6;
}
.update-actions {
  display: flex;
  gap: 20rpx;
}
.update-btn {
  flex: 1;
  padding: 20rpx;
  border-radius: 10rpx;
  text-align: center;
  font-size: 28rpx;
}
.update-btn.primary {
  background: #1890ff;
  color: #fff;
}
.update-btn.secondary {
  background: #f5f5f5;
  color: #333;
}
</style>

<template>
  <view>
    <slot />

    <!-- 版本更新弹窗 -->
    <view class="update-mask" v-if="showUpdateDialog" @tap="updateInfo.mandatory ? null : dismissUpdate()">
      <view class="update-card" @tap.stop>
        <text class="update-title">{{ updateInfo.mandatory ? $t('version.mandatoryTitle') : $t('version.update') }}</text>
        <text class="update-version">v{{ updateInfo.version }}</text>
        <text class="update-notes" v-if="updateInfo.notes">{{ updateInfo.notes }}</text>
        <view class="update-actions">
          <view class="update-btn primary" @tap="doUpdate">{{ $t('version.updateNow') }}</view>
          <view class="update-btn secondary" v-if="!updateInfo.mandatory" @tap="dismissUpdate">{{ $t('version.updateLater') }}</view>
        </view>
      </view>
    </view>
  </view>
</template>
