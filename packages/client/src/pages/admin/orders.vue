<template>
  <view class="container">
    <view
      v-for="o in orders"
      :key="o.id"
      class="order-card"
    >
      <view class="o-head">
        <text class="o-title">
          #{{ o.id }} {{ o.project?.title || '-' }}
        </text>
        <text class="o-status">
          {{ statusText(o.status) }}
        </text>
      </view>
      <view class="o-info">
        <text>{{ $t('admin.amount') }}：¥{{ o.amount }}</text>
        <text>{{ $t('admin.supplierId') }}：{{ o.supplier_id }}</text>
        <text>{{ $t('admin.createdAt') }}：{{ fmtTime(o.created_at) }}</text>
      </view>
    </view>

    <view
      v-if="!orders.length"
      class="empty"
    >
      {{ $t('admin.noData') }}
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
usePageTitle('page.adminOrders', { onShow })

const orders = ref<any[]>([])

const load = async () => {
  try {
    const res = await request.get('/api/v1/admin/orders', { silent401: true })
    orders.value = res.orders || []
  } catch {
    orders.value = []
  }
}

onShow(() => {
  if (userStore.user?.user_type !== 3) {
    uni.reLaunch({ url: '/pages/index/index' })
    return
  }
  load()
})

const fmtTime = (s: string) => (s ? s.replace('T', ' ').slice(0, 19) : '')

const statusText = (s: number) => {
  const map: Record<number, string> = { 0: $t('order.draft'), 1: $t('order.signed'), 2: $t('order.inProgress'), 3: $t('order.completed'), 4: $t('order.disputed'), 5: $t('order.cancelled') }
  return map[s] || s
}
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.order-card {
  background: var(--card-bg, #fff);
  border-radius: 14rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.o-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}
.o-title {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--text-color, #333);
  flex: 1;
}
.o-status {
  font-size: 24rpx;
  color: var(--muted-color, #666);
}
.o-info {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  font-size: 24rpx;
  color: var(--muted-color, #666);
}
.empty {
  text-align: center;
  color: #999;
  padding: 60rpx 0;
  font-size: 26rpx;
}
</style>
