<template>
  <view class="container">
    <view class="filter-bar">
      <view :class="['filter-item', activeStatus === -1 ? 'active' : '']" @tap="filterByStatus(-1)">全部</view>
      <view :class="['filter-item', activeStatus === 1 ? 'active' : '']" @tap="filterByStatus(1)">进行中</view>
      <view :class="['filter-item', activeStatus === 2 ? 'active' : '']" @tap="filterByStatus(2)">待验收</view>
      <view :class="['filter-item', activeStatus === 3 ? 'active' : '']" @tap="filterByStatus(3)">已完成</view>
    </view>

    <view class="order-list">
      <view class="order-card" v-for="order in orders" :key="order.id" @tap="goToDetail(order.id)">
        <view class="order-info">
          <text class="order-title">订单 #{{ order.id }}</text>
          <text class="order-amount">金额：¥{{ order.amount }}</text>
          <text class="order-status">状态：{{ statusText(order.status) }}</text>
        </view>
      </view>
    </view>

    <view class="empty" v-if="orders.length === 0">
      <text>暂无订单</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const activeStatus = ref(-1)
const orders = ref([])

const filterByStatus = (status: number) => {
  activeStatus.value = status
}

const goToDetail = (id: number) => {
  uni.navigateTo({ url: `/pages/order/detail?id=${id}` })
}

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: '待签约', 1: '进行中', 2: '待验收', 3: '已完成', 4: '纠纷中'
  }
  return map[status] || '未知'
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.filter-bar {
  display: flex;
  background: #fff;
  padding: 20rpx;
  border-radius: 10rpx;
  margin-bottom: 20rpx;
}

.filter-item {
  flex: 1;
  text-align: center;
  font-size: 28rpx;
  color: #666;
}

.filter-item.active {
  color: #1890ff;
  font-weight: bold;
}

.order-card {
  background: #fff;
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.order-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 10rpx;
}

.order-amount, .order-status {
  font-size: 26rpx;
  color: #666;
  display: block;
  margin-bottom: 5rpx;
}

.empty {
  text-align: center;
  padding: 100rpx 0;
  color: #999;
}
</style>
