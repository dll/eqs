<template>
  <view class="container">
    <view class="order-info" v-if="order">
      <view class="info-card">
        <text class="order-title">订单 #{{ order.id }}</text>
        <text class="order-amount">金额：¥{{ order.amount }}</text>
        <text class="order-status">状态：{{ statusText(order.status) }}</text>
      </view>

      <view class="action-bar">
        <button class="action-btn primary" v-if="order.status === 2" @tap="confirmDelivery">确认验收</button>
        <button class="action-btn" @tap="uploadDeliverable">上传交付物</button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const order = ref<any>(null)

onLoad((options) => {
  if (options?.id) {
    order.value = { id: options.id, amount: 50000, status: 1 }
  }
})

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: '待签约', 1: '进行中', 2: '待验收', 3: '已完成', 4: '纠纷中'
  }
  return map[status] || '未知'
}

const confirmDelivery = () => {
  uni.showToast({ title: '验收成功', icon: 'success' })
}

const uploadDeliverable = () => {
  // TODO: Open file picker
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.info-card {
  background: #fff;
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.order-title {
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 15rpx;
}

.order-amount, .order-status {
  font-size: 28rpx;
  color: #666;
  display: block;
  margin-bottom: 10rpx;
}

.action-bar {
  display: flex;
  gap: 20rpx;
}

.action-btn {
  flex: 1;
  background: #f5f5f5;
  border-radius: 10rpx;
  font-size: 28rpx;
}

.action-btn.primary {
  background: #1890ff;
  color: #fff;
}
</style>
