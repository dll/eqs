<template>
  <view class="container">
    <view v-if="messages.length" class="msg-list">
      <view v-for="m in messages" :key="m.id" class="msg-item" @tap="openOrder(m)">
        <view class="msg-row">
          <text class="msg-title">{{ m.title }}</text>
          <text class="msg-time">{{ fmtTime(m.created_at) }}</text>
        </view>
        <text class="msg-content">{{ m.content }}</text>
      </view>
    </view>
    <view v-else class="empty">
      <text>{{ $t('message.empty') }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.notice', { onShow })

interface Notice {
  id: number
  title: string
  content: string
  type: string
  order_id?: number
  created_at: string
}
const messages = ref<Notice[]>([])

onShow(() => load())

const load = async () => {
  try {
    const res = await request.get('/api/v1/notification/list', { silent401: true }).catch(() => ({ notifications: [] }))
    messages.value = res.notifications || []
  } catch {
    // request 已提示
  }
}

const fmtTime = (s: string) => (s ? s.replace('T', ' ').slice(5, 16) : '')

const openOrder = (m: Notice) => {
  if (m.order_id) uni.navigateTo({ url: `/pages/order/detail?id=${m.order_id}` })
}
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.msg-list .msg-item {
  background: var(--card-bg, #fff);
  border-radius: 10rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.msg-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10rpx;
}
.msg-title {
  font-size: 30rpx;
  font-weight: bold;
  color: var(--text-color, #333);
}
.msg-time {
  font-size: 22rpx;
  color: var(--muted-color, #999);
}
.msg-content {
  font-size: 26rpx;
  color: var(--text-color, #333);
  display: block;
  line-height: 1.6;
}
.empty {
  text-align: center;
  padding: 100rpx 0;
  color: var(--muted-color, #999);
}
</style>
