<template>
  <view class="container">
    <view class="filter-bar">
      <view
        :class="['filter-item', activeStatus === -1 ? 'active' : '']"
        @tap="filterByStatus(-1)"
      >
        {{ $t('common.all') }}
      </view>
      <view
        :class="['filter-item', activeStatus === 0 ? 'active' : '']"
        @tap="filterByStatus(0)"
      >
        {{ $t('order.pending') }}
      </view>
      <view
        :class="['filter-item', activeStatus === 1 ? 'active' : '']"
        @tap="filterByStatus(1)"
      >
        {{ $t('order.inProgress') }}
      </view>
      <view
        :class="['filter-item', activeStatus === 2 ? 'active' : '']"
        @tap="filterByStatus(2)"
      >
        {{ $t('order.toAccept') }}
      </view>
      <view
        :class="['filter-item', activeStatus === 3 ? 'active' : '']"
        @tap="filterByStatus(3)"
      >
        {{ $t('order.completed') }}
      </view>
    </view>

    <view class="order-list">
      <view
        v-for="order in orders"
        :key="order.id"
        class="order-card"
        @tap="goToDetail(order.id)"
      >
        <view class="order-info">
          <text class="order-title">
            {{ $t('order.orderNo', { id: order.id }) }} · {{ order.project?.title || $t('order.project') }}
          </text>
          <text class="order-amount">
            {{ $t('order.amountPrefix', { amount: order.amount }) }}
          </text>
          <text class="order-status">
            {{ $t('order.statusPrefix') }}{{ statusText(order.status) }}
          </text>
        </view>
      </view>
    </view>

    <view
      v-if="orders.length === 0"
      class="empty"
    >
      <text>{{ $t('common.empty') }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle, applyTabBarI18n } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.orderList', { onShow })
const activeStatus = ref(-1)
const orders = ref<any[]>([])

const loadOrders = async () => {
  try {
    const res = await request.get('/api/v1/order/list')
    const all = res.orders || []
    orders.value = activeStatus.value === -1 ? all : all.filter((o: any) => o.status === activeStatus.value)
  } catch {
    // request 已提示
  }
}

onShow(() => {
  applyTabBarI18n()
  loadOrders()
})

const filterByStatus = (status: number) => {
  activeStatus.value = status
  loadOrders()
}

const goToDetail = (id: number) => {
  uni.navigateTo({ url: `/pages/order/detail?id=${id}` })
}

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: $t('order.pending'), 1: $t('order.inProgress'), 2: $t('order.toAccept'), 3: $t('order.completed'), 4: $t('order.dispute')
  }
  return map[status] || $t('order.unknown')
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.filter-bar {
  display: flex;
  background: var(--card-bg);
  padding: 20rpx;
  border-radius: 10rpx;
  margin-bottom: 20rpx;
}

.filter-item {
  flex: 1;
  text-align: center;
  font-size: 26rpx;
  color: var(--muted-color);
}

.filter-item.active {
  color: var(--primary-color);
  font-weight: bold;
}

.order-card {
  background: var(--card-bg);
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.order-title {
  font-size: 30rpx;
  font-weight: bold;
  color: var(--text-color);
  display: block;
  margin-bottom: 10rpx;
}

.order-amount, .order-status {
  font-size: 26rpx;
  color: var(--muted-color);
  display: block;
  margin-bottom: 5rpx;
}

.empty {
  text-align: center;
  padding: 100rpx 0;
  color: var(--muted-color);
}
</style>