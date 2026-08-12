<template>
  <view class="container">
    <view
      v-for="u in users"
      :key="u.id"
      class="user-card"
    >
      <view class="u-head">
        <text class="u-title">
          #{{ u.id }} {{ u.phone }}
        </text>
        <text
          class="u-status"
          :style="{ color: u.status === 1 ? '#67C23A' : '#F56C6C' }"
        >
          {{ u.status === 1 ? $t('admin.statusActive') : $t('admin.statusDisabled') }}
        </text>
      </view>
      <view class="u-info">
        <text>{{ $t('admin.userType') }}：{{ userTypeText(u.user_type) }}</text>
        <text>{{ $t('admin.company') }}：{{ u.company_name || '-' }}</text>
        <text>{{ $t('admin.creditScore') }}：{{ u.credit_score }}</text>
      </view>
      <view class="u-actions">
        <button
          class="btn btn-detail"
          @tap="viewDetail(u)"
        >
          {{ $t('admin.detail') }}
        </button>
        <button
          class="btn"
          :class="u.status === 1 ? 'btn-danger' : 'btn-success'"
          @tap="toggleStatus(u)"
        >
          {{ u.status === 1 ? $t('admin.disable') : $t('admin.enable') }}
        </button>
      </view>
    </view>

    <view
      v-if="!users.length"
      class="empty"
    >
      {{ $t('admin.noData') }}
    </view>

    <!-- 用户详情 -->
    <view
      v-if="detail"
      class="detail-mask"
      @tap="detail = null"
    >
      <view
        class="detail-box"
        @tap.stop
      >
        <text class="detail-title">
          {{ $t('admin.detail') }} #{{ detail.user.id }}
        </text>
        <view class="detail-row">
          <text class="d-label">{{ $t('admin.phone') }}</text>
          <text>{{ detail.user.phone }}</text>
        </view>
        <view class="detail-row">
          <text class="d-label">{{ $t('admin.userType') }}</text>
          <text>{{ userTypeText(detail.user.user_type) }}</text>
        </view>
        <view class="detail-row">
          <text class="d-label">{{ $t('admin.company') }}</text>
          <text>{{ detail.user.company_name || '-' }}</text>
        </view>
        <view class="detail-row">
          <text class="d-label">{{ $t('admin.creditScore') }}</text>
          <text>{{ detail.user.credit_score }}</text>
        </view>
        <view class="detail-stats">
          <text class="ds-item">{{ $t('admin.statProjects') }} {{ detail.stats.projects }}</text>
          <text class="ds-item">{{ $t('admin.statOrdersOwner') }} {{ detail.stats.orders_as_owner }}</text>
          <text class="ds-item">{{ $t('admin.statOrdersSupplier') }} {{ detail.stats.orders_as_supplier }}</text>
          <text class="ds-item">{{ $t('admin.statQualifications') }} {{ detail.stats.qualifications }}</text>
        </view>
        <button
          class="btn btn-close"
          @tap="detail = null"
        >
          {{ $t('admin.close') }}
        </button>
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
usePageTitle('page.adminUsers', { onShow })

const users = ref<any[]>([])
const detail = ref<any>(null)

const load = async () => {
  try {
    const res = await request.get('/api/v1/admin/users', { silent401: true })
    users.value = res.users || []
  } catch {
    users.value = []
  }
}

onShow(() => {
  if (userStore.user?.user_type !== 3) {
    uni.reLaunch({ url: '/pages/index/index' })
    return
  }
  load()
})

const viewDetail = async (u: any) => {
  try {
    detail.value = await request.get(`/api/v1/admin/users/${u.id}`, { silent401: true })
  } catch { /* toast 已提示 */ }
}

const toggleStatus = async (u: any) => {
  const next = u.status === 1 ? 0 : 1
  try {
    await request.put(`/api/v1/admin/users/${u.id}/status`, { status: next })
    uni.showToast({ title: $t('common.success'), icon: 'success' })
    load()
  } catch { /* toast 已提示 */ }
}

const userTypeText = (t: number) => {
  const map: Record<number, string> = { 1: $t('role.client'), 2: $t('role.supplier'), 3: $t('role.admin'), 4: $t('role.expert') }
  return map[t] || $t('role.unknown')
}
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.user-card {
  background: var(--card-bg, #fff);
  border-radius: 14rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.u-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}
.u-title {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--text-color, #333);
}
.u-status {
  font-size: 24rpx;
}
.u-info {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  font-size: 24rpx;
  color: var(--muted-color, #666);
}
.u-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 16rpx;
}
.btn {
  flex: 1;
  font-size: 26rpx;
  border-radius: 10rpx;
  line-height: 60rpx;
  height: 60rpx;
  padding: 0;
  margin: 0;
}
.btn-detail {
  background: #eef4ff;
  color: #2563eb;
}
.btn-danger {
  background: #fef0f0;
  color: #f56c6c;
}
.btn-success {
  background: #f0f9eb;
  color: #67c23a;
}
.empty {
  text-align: center;
  color: #999;
  padding: 60rpx 0;
  font-size: 26rpx;
}
.detail-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, .5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}
.detail-box {
  width: 82%;
  background: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
}
.detail-title {
  font-size: 32rpx;
  font-weight: 700;
  display: block;
  margin-bottom: 20rpx;
  color: #333;
}
.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 12rpx 0;
  font-size: 26rpx;
  border-bottom: 1rpx solid #f0f0f0;
}
.d-label {
  color: #999;
}
.detail-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin: 20rpx 0;
}
.ds-item {
  background: #f5f7fb;
  border-radius: 8rpx;
  padding: 8rpx 14rpx;
  font-size: 22rpx;
  color: #2563eb;
}
.btn-close {
  background: #f0f2f5;
  color: #666;
  margin-top: 10rpx;
}
</style>
