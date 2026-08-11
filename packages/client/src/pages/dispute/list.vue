<template>
  <view class="container">
    <view
      v-if="disputes.length"
      class="d-list"
    >
      <view
        v-for="d in disputes"
        :key="d.id"
        class="d-item"
        @tap="openDetail(d.id)"
      >
        <view class="d-row">
          <text class="d-order">
            订单#{{ d.order_id }}
          </text>
          <text
            class="d-status"
            :class="'s-' + d.status"
          >
            {{ statusText(d.status) }}
          </text>
        </view>
        <text class="d-reason">
          {{ d.reason }}
        </text>
        <text class="d-claim">
          {{ d.claim }}
        </text>
      </view>
    </view>
    <view
      v-else
      class="empty"
    >
      <text>{{ $t('dispute.empty') }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.disputeList', { onShow })

interface Dispute {
  id: number
  order_id: number
  initiator_id: number
  reason: string
  claim: string
  status: string
}
const disputes = ref<Dispute[]>([])

onShow(() => load())

const load = async () => {
  try {
    const res = await request.get('/api/v1/dispute/mine', { silent401: true }).catch(() => ({ disputes: [] }))
    disputes.value = res.disputes || []
  } catch {
    // request 已提示
  }
}

const openDetail = (id: number) => {
  uni.navigateTo({ url: `/pages/dispute/detail?id=${id}` })
}

const statusText = (s: string) => {
  const map: Record<string, string> = {
    evidence: $t('dispute.status.evidence'),
    review: $t('dispute.status.review'),
    mediation: $t('dispute.status.mediation'),
    reconsideration: $t('dispute.status.reconsideration'),
    closed: $t('dispute.status.closed'),
  }
  return map[s] || s
}
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.d-item {
  background: var(--card-bg, #fff);
  border-radius: 10rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.d-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10rpx;
}
.d-order {
  font-size: 28rpx;
  color: var(--muted-color, #666);
}
.d-status {
  font-size: 24rpx;
  padding: 4rpx 14rpx;
  border-radius: 20rpx;
  color: #fff;
}
.s-evidence { background: #e6a23c; }
.s-review { background: #409eff; }
.s-mediation { background: #f56c6c; }
.s-reconsideration { background: #909399; }
.s-closed { background: #67c23a; }
.d-reason {
  font-size: 30rpx;
  font-weight: bold;
  color: var(--text-color, #333);
  display: block;
  margin-bottom: 8rpx;
}
.d-claim {
  font-size: 26rpx;
  color: var(--muted-color, #666);
  display: block;
}
.empty {
  text-align: center;
  padding: 100rpx 0;
  color: var(--muted-color, #999);
}
</style>
