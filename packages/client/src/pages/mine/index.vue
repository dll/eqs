<template>
  <view class="container">
    <view
      v-if="userStore.user"
      class="user-card"
    >
      <text class="user-name">
        {{ userStore.user.company_name || $t('mine.company') }}
      </text>
      <text class="user-phone">
        {{ userStore.user.phone }}
      </text>
      <text class="user-score">
        {{ $t('mine.creditScore') }}:{{ userStore.user.credit_score }}
      </text>
    </view>

    <view class="menu-list">
      <view
        v-if="userStore.user?.user_type === 3"
        class="menu-item"
        @tap="goTo('/pages/admin/index')"
      >
        <text>{{ $t('mine.adminCenter') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        class="menu-item"
        @tap="goTo('/pages/message/index')"
      >
        <view class="menu-left">
          <text>{{ $t('mine.notice') }}</text>
          <text
            v-if="unread > 0"
            class="badge"
          >
            {{ unread }}
          </text>
        </view>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        class="menu-item"
        @tap="goTo('/pages/order/list')"
      >
        <text>{{ $t('mine.myOrders') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        v-if="!isSupplier"
        class="menu-item"
        @tap="goTo('/pages/project/mine')"
      >
        <text>{{ $t('mine.myProjects') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        v-if="isSupplier"
        class="menu-item"
        @tap="goTo('/pages/bid/mine')"
      >
        <text>{{ $t('mine.myBids') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        class="menu-item"
        @tap="goTo('/pages/dispute/list')"
      >
        <text>{{ $t('mine.disputes') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        v-if="isSupplier"
        class="menu-item"
        @tap="goTo('/pages/qualification/index')"
      >
        <text>{{ $t('mine.qual') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        v-if="isSupplier"
        class="menu-item"
        @tap="goTo('/pages/template/list')"
      >
        <text>{{ $t('mine.templates') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        v-if="isSupplier"
        class="menu-item"
        @tap="goTo('/pages/case/mine')"
      >
        <text>{{ $t('mine.cases') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        class="menu-item"
        @tap="showThemePicker"
      >
        <text>{{ $t('mine.themeCurrent', { name: themeName }) }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        class="menu-item"
        @tap="showLangPicker"
      >
        <text>{{ $t('mine.langCurrent', { name: langName }) }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        class="menu-item"
        @tap="goTo('/pages/tools/estimate')"
      >
        <text>{{ $t('mine.tools') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        class="menu-item"
        @tap="goTo('/pages/member/index')"
      >
        <text>{{ $t('mine.member') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
      <view
        class="menu-item"
        @tap="logout"
      >
        <text>{{ $t('mine.logout') }}</text>
        <text class="arrow">
          >
        </text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onHide, onShow } from '@dcloudio/uni-app'
import { useUserStore } from '@/store/user'
import { useSettingsStore, THEMES, LANGS } from '@/store/settings'
import { useI18n, usePageTitle, applyTabBarI18n } from '@/utils/i18n'
import { request } from '@/utils/request'
import { connectNotify } from '@/utils/realtime'

const userStore = useUserStore()
const settingsStore = useSettingsStore()
const { $t } = useI18n()
usePageTitle('page.mine', { onShow })

// V10：消息未读角标
const unread = ref(0)
const loadUnread = async () => {
  try {
    const res = await request.get('/api/v1/notification/unread-count', { silent401: true }).catch(() => ({ unread: 0 }))
    unread.value = res.unread || 0
  } catch {
    unread.value = 0
  }
}

// V9：站内实时推送（H5 EventSource / 其他端轮询兜底），未读角标实时刷新
let disconnectNotify: (() => void) | null = null

onShow(() => {
  applyTabBarI18n()
  loadUnread()
  disconnectNotify?.()
  disconnectNotify = connectNotify((data) => {
    if (data && typeof data.unread === 'number') {
      unread.value = data.unread
    }
  })
})

onHide(() => {
  disconnectNotify?.()
  disconnectNotify = null
})

const themeName = computed(() => $t('theme.' + settingsStore.theme))
const langName = computed(() => $t('lang.' + settingsStore.lang))
const isSupplier = computed(() => userStore.user?.user_type === 2)

const goTo = (url: string) => {
  uni.navigateTo({ url })
}

const showThemePicker = () => {
  uni.showActionSheet({
    itemList: THEMES.map(x => $t('theme.' + x.id)),
    success: (res: any) => {
      const theme = THEMES[res.tapIndex]
      if (theme) settingsStore.setTheme(theme.id)
    },
  })
}

const showLangPicker = () => {
  uni.showActionSheet({
    itemList: LANGS.map(l => $t('lang.' + l.id)),
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
  background: var(--primary-color);
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
  background: var(--card-bg);
  border-radius: 10rpx;
}

.menu-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 30rpx;
  border-bottom: 1rpx solid var(--border-color);
  font-size: 30rpx;
  color: var(--text-color);
}
.menu-left {
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.badge {
  min-width: 32rpx;
  height: 32rpx;
  padding: 0 8rpx;
  border-radius: 16rpx;
  background: #ef4444;
  color: #fff;
  font-size: 20rpx;
  line-height: 32rpx;
  text-align: center;
}

.arrow {
  color: var(--muted-color);
}
</style>