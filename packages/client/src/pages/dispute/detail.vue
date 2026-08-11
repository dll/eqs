<template>
  <view class="container">
    <view v-if="dispute" class="detail">
      <view class="card">
        <view class="row">
          <text class="label">{{ $t('dispute.order') }}</text>
          <text class="value">#{{ dispute.order_id }}</text>
        </view>
        <view class="row">
          <text class="label">{{ $t('dispute.status') }}</text>
          <text class="value">{{ statusText(dispute.status) }}</text>
        </view>
        <view class="row">
          <text class="label">{{ $t('dispute.reason') }}</text>
          <text class="value">{{ dispute.reason }}</text>
        </view>
        <view class="row">
          <text class="label">{{ $t('dispute.claim') }}</text>
          <text class="value">{{ dispute.claim }}</text>
        </view>
      </view>

      <view v-if="evidence.length" class="card">
        <view class="card-title">{{ $t('dispute.evidence') }}</view>
        <view v-for="e in evidence" :key="e.id" class="ev-item">
          <text>{{ e.content }}</text>
        </view>
      </view>

      <view v-if="assignments.length" class="card">
        <view class="card-title">{{ $t('dispute.experts') }}</view>
        <view v-for="a in assignments" :key="a.id" class="ev-item">
          <text>{{ $t('dispute.expertId') }}:{{ a.expert_user_id }}</text>
          <text v-if="a.vote" class="vote"> {{ a.vote }}</text>
        </view>
      </view>
    </view>
    <view v-else class="empty">
      <text>{{ $t('common.loading') }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.disputeDetail', { onLoad })

const dispute = ref<any>(null)
const evidence = ref<any[]>([])
const assignments = ref<any[]>([])

onLoad((options) => {
  if (options?.id) load(options.id)
})

const load = async (id: string) => {
  try {
    const res = await request.get(`/api/v1/dispute/${id}`, { silent401: true })
    dispute.value = res.dispute
    evidence.value = res.evidence || []
    assignments.value = res.assignments || []
  } catch {
    dispute.value = null
  }
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
.card {
  background: var(--card-bg, #fff);
  border-radius: 10rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.card-title {
  font-size: 30rpx;
  font-weight: bold;
  color: var(--text-color, #333);
  margin-bottom: 14rpx;
}
.row {
  display: flex;
  justify-content: space-between;
  padding: 10rpx 0;
  font-size: 28rpx;
}
.label {
  color: var(--muted-color, #666);
  flex-shrink: 0;
  margin-right: 20rpx;
}
.value {
  color: var(--text-color, #333);
  text-align: right;
}
.ev-item {
  padding: 10rpx 0;
  font-size: 26rpx;
  color: var(--text-color, #333);
  border-bottom: 1rpx solid var(--border-color, #eee);
}
.ev-item:last-child {
  border-bottom: none;
}
.vote {
  color: var(--primary-color);
}
.empty {
  text-align: center;
  padding: 100rpx 0;
  color: var(--muted-color, #999);
}
</style>
