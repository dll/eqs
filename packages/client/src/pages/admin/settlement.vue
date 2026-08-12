<template>
  <view class="container">
    <text class="section-title">{{ $t('admin.commission') }}</text>
    <view
      v-for="c in commissions"
      :key="c.id"
      class="settle-card"
    >
      <view class="s-head">
        <text class="s-title">
          #{{ c.id }} {{ $t('admin.commission') }}
        </text>
        <text
          class="s-status"
          :style="{ color: c.status === 'collected' ? '#67C23A' : '#E6A23C' }"
        >
          {{ commissionStatusText(c.status) }}
        </text>
      </view>
      <view class="s-info">
        <text>{{ $t('admin.amount') }}：¥{{ c.commission ?? c.amount }}</text>
        <text>{{ $t('admin.orderId') }}：#{{ c.order_id }}</text>
      </view>
      <button
        v-if="c.status !== 'collected'"
        class="btn btn-collect"
        @tap="collect(c)"
      >
        {{ $t('admin.collect') }}
      </button>
    </view>

    <text class="section-title">{{ $t('admin.transactions') }}</text>
    <view
      v-for="t in transactions"
      :key="t.id"
      class="settle-card"
    >
      <view class="s-head">
        <text class="s-title">
          #{{ t.id }} ¥{{ t.amount }}
        </text>
        <text class="s-status">
          {{ typeText(t.type) }}
        </text>
      </view>
      <view class="s-info">
        <text>{{ $t('admin.orderId') }}：#{{ t.order_id }}｜{{ $t('admin.channel') }}：{{ t.channel }}</text>
        <text>{{ $t('admin.createdAt') }}：{{ fmtTime(t.created_at) }}</text>
      </view>
    </view>

    <view
      v-if="!commissions.length && !transactions.length"
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
usePageTitle('page.adminSettlement', { onShow })

const commissions = ref<any[]>([])
const transactions = ref<any[]>([])

const load = async () => {
  try {
    const r1 = await request.get('/api/v1/admin/commission/list', { silent401: true })
    commissions.value = r1.commissions || []
  } catch { commissions.value = [] }
  try {
    const r2 = await request.get('/api/v1/admin/transactions', { silent401: true })
    transactions.value = r2.transactions || []
  } catch { transactions.value = [] }
}

onShow(() => {
  if (userStore.user?.user_type !== 3) {
    uni.reLaunch({ url: '/pages/index/index' })
    return
  }
  load()
})

const collect = async (c: any) => {
  try {
    await request.post(`/api/v1/admin/commission/${c.id}/collect`)
    uni.showToast({ title: $t('common.success'), icon: 'success' })
    load()
  } catch { /* toast 已提示 */ }
}

const commissionStatusText = (s: string) => {
  const map: Record<string, string> = { pending: $t('admin.commissionPending'), collected: $t('admin.commissionCollected') }
  return map[s] || s
}

const typeText = (t: string) => {
  const map: Record<string, string> = {
    payment: $t('admin.typePayment'), settlement: $t('admin.typeSettlement'),
    refund: $t('admin.typeRefund'), freeze: $t('admin.typeFreeze'), unfreeze: $t('admin.typeUnfreeze'),
  }
  return map[t] || t
}

const fmtTime = (s: string) => (s ? s.replace('T', ' ').slice(0, 19) : '')
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.section-title {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: var(--text-color, #333);
  margin: 10rpx 0 16rpx;
}
.settle-card {
  background: var(--card-bg, #fff);
  border-radius: 14rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.s-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}
.s-title {
  font-size: 28rpx;
  font-weight: 600;
  color: var(--text-color, #333);
}
.s-status {
  font-size: 24rpx;
}
.s-info {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  font-size: 24rpx;
  color: var(--muted-color, #666);
}
.btn-collect {
  margin-top: 14rpx;
  background: #fdf6ec;
  color: #e6a23c;
  font-size: 26rpx;
  border-radius: 10rpx;
  line-height: 60rpx;
  height: 60rpx;
  padding: 0;
}
.empty {
  text-align: center;
  color: #999;
  padding: 60rpx 0;
  font-size: 26rpx;
}
</style>
