<template>
  <view class="container">
    <view
      v-if="bids.length"
      class="bid-list"
    >
      <view
        v-for="b in bids"
        :key="b.id"
        class="bid-card"
      >
        <view class="b-head">
          <text class="b-amount">
            ¥{{ b.amount }}
          </text>
          <text class="b-status">
            {{ statusText(b.status) }}
          </text>
        </view>
        <text class="b-meta">
          项目 #{{ b.project_id }} · 工期 {{ b.service_days }} 天
        </text>
        <text class="b-time">
          报价时间：{{ fmtTime(b.created_at) }}
        </text>
        <button
          v-if="b.status === 'submitted'"
          class="withdraw-btn"
          @tap="withdraw(b)"
        >
          撤回报价
        </button>
      </view>
      <view
        v-if="hasMore"
        class="load-more"
        @tap="loadMore"
      >
        加载更多
      </view>
    </view>
    <view
      v-else
      class="empty"
    >
      暂无报价记录
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.bidMine', { onShow })

const bids = ref<any[]>([])
const page = ref(1)
const hasMore = ref(true)

const load = async (append = false) => {
  try {
    const res = await request.get(`/api/v1/bid/mine?page=${page.value}&size=20`)
    const list = res.bids || []
    bids.value = append ? [...bids.value, ...list] : list
    hasMore.value = list.length >= 20
  } catch {
    // request 已提示
  }
}

onShow(() => { page.value = 1; load() })
onReachBottom(() => {
  if (!hasMore.value) return
  page.value++
  load(true)
})

const loadMore = () => {
  if (!hasMore.value) return
  page.value++
  load(true)
}

// 撤回报价（截止前未中选可撤回）
const withdraw = (b: any) => {
  uni.showModal({
    title: '撤回报价',
    content: `确认撤回 #${b.id} 的报价？`,
    success: async (r: any) => {
      if (!r.confirm) return
      try {
        await request.put(`/api/v1/bid/${b.id}/withdraw`)
        uni.showToast({ title: '已撤回', icon: 'none' })
        page.value = 1
        load()
      } catch {
        // request 已提示
      }
    },
  })
}

const statusText = (s: string) => {
  const map: Record<string, string> = {
    submitted: $t('bid.submitted'),
    selected: $t('bid.selected'),
    rejected: $t('bid.rejected'),
    withdrawn: $t('bid.withdrawn'),
  }
  return map[s] || s
}

const fmtTime = (s: string) => (s ? s.replace('T', ' ').slice(0, 16) : '')
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.bid-card {
  background: var(--card-bg, #fff);
  border-radius: 14rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.b-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10rpx;
}
.b-amount {
  font-size: 34rpx;
  font-weight: 700;
  color: var(--primary-color, #2563eb);
}
.b-status {
  font-size: 22rpx;
  color: var(--muted-color, #666);
  background: var(--input-bg, #f5f5f5);
  padding: 4rpx 14rpx;
  border-radius: 20rpx;
}
.b-meta {
  font-size: 26rpx;
  color: var(--text-color, #333);
  display: block;
  margin-bottom: 8rpx;
}
.b-time {
  font-size: 22rpx;
  color: var(--muted-color, #999);
}
.withdraw-btn {
  margin-top: 16rpx;
  background: var(--input-bg, #f5f5f5);
  color: var(--danger-color, #ef4444);
  border: 1rpx solid var(--danger-color, #ef4444);
  border-radius: 10rpx;
  font-size: 26rpx;
  padding: 10rpx 0;
}
.empty {
  text-align: center;
  color: var(--muted-color, #999);
  padding: 120rpx 0;
  font-size: 26rpx;
}
.load-more {
  text-align: center;
  color: var(--primary-color);
  padding: 20rpx;
  font-size: 26rpx;
}
</style>
