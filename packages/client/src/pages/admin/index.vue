<template>
  <view class="container">
    <!-- 数据概览 -->
    <view class="stats-grid">
      <view class="stat-card">
        <text class="stat-num">{{ stats.user_count }}</text>
        <text class="stat-label">{{ $t('admin.totalUsers') }}</text>
      </view>
      <view class="stat-card">
        <text class="stat-num">{{ stats.project_count }}</text>
        <text class="stat-label">{{ $t('admin.totalProjects') }}</text>
      </view>
      <view class="stat-card">
        <text class="stat-num">{{ stats.order_count }}</text>
        <text class="stat-label">{{ $t('admin.totalOrders') }}</text>
      </view>
      <view class="stat-card">
        <text class="stat-num">¥{{ stats.settled_amount }}</text>
        <text class="stat-label">{{ $t('admin.totalSettled') }}</text>
      </view>
    </view>

    <!-- 模块入口 -->
    <view class="menu-grid">
      <view
        v-for="m in modules"
        :key="m.url"
        class="menu-card"
        @tap="goTo(m.url)"
      >
        <text class="menu-icon">{{ m.icon }}</text>
        <text class="menu-title">{{ m.title }}</text>
        <text class="arrow">></text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '@/store/user'
import { useI18n, usePageTitle } from '@/utils/i18n'
import { request } from '@/utils/request'

const userStore = useUserStore()
const { $t } = useI18n()
usePageTitle('page.adminHome', { onShow })

const stats = ref<any>({})

const modules = [
  { icon: '👥', url: '/pages/admin/users', title: $t('admin.users') },
  { icon: '📋', url: '/pages/admin/audit', title: $t('admin.audit') },
  { icon: '🗂️', url: '/pages/admin/projects', title: $t('admin.projects') },
  { icon: '🧾', url: '/pages/admin/orders', title: $t('admin.orders') },
  { icon: '💰', url: '/pages/admin/settlement', title: $t('admin.settlement') },
  { icon: '⚖️', url: '/pages/admin/disputes', title: $t('admin.disputes') },
]

const loadStats = async () => {
  try {
    const res = await request.get('/api/v1/admin/stats', { silent401: true })
    stats.value = res || {}
  } catch {
    stats.value = {}
  }
}

onShow(() => {
  if (userStore.user?.user_type !== 3) {
    uni.reLaunch({ url: '/pages/index/index' })
    return
  }
  loadStats()
})

const goTo = (url: string) => {
  uni.navigateTo({ url })
}
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16rpx;
  margin-bottom: 30rpx;
}
.stat-card {
  background: linear-gradient(135deg, #2563eb, #06b6d4);
  border-radius: 14rpx;
  padding: 26rpx;
  color: #fff;
}
.stat-num {
  font-size: 34rpx;
  font-weight: 700;
  display: block;
}
.stat-label {
  font-size: 22rpx;
  opacity: .9;
  margin-top: 6rpx;
}
.menu-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16rpx;
}
.menu-card {
  background: var(--card-bg, #fff);
  border-radius: 14rpx;
  padding: 28rpx;
  display: flex;
  align-items: center;
  gap: 14rpx;
}
.menu-icon {
  font-size: 34rpx;
}
.menu-title {
  flex: 1;
  font-size: 28rpx;
  color: var(--text-color, #333);
}
.arrow {
  color: var(--muted-color, #999);
  font-size: 26rpx;
}
</style>
