<template>
  <view class="container">
    <view
      v-for="d in disputes"
      :key="d.id"
      class="dispute-card"
      @tap="viewDetail(d)"
    >
      <view class="d-head">
        <text class="d-title">
          #{{ d.id }} {{ $t('admin.dispute') }}
        </text>
        <text
          class="d-status"
          :class="'st-' + d.status"
        >
          {{ statusText(d.status) }}
        </text>
      </view>
      <view class="d-info">
        <text>{{ $t('admin.orderId') }}：#{{ d.order_id }}｜{{ $t('dispute.initiator') }}：#{{ d.initiator_id }}</text>
        <text>{{ $t('admin.reason') }}：{{ d.reason }}</text>
      </view>
      <view
        v-if="d.status !== 'closed'"
        class="d-actions"
      >
        <button
          class="btn btn-close"
          @tap.stop="closeDispute(d)"
        >
          {{ $t('admin.closeDispute') }}
        </button>
      </view>
    </view>

    <view
      v-if="!disputes.length"
      class="empty"
    >
      {{ $t('admin.noData') }}
    </view>

    <!-- 详情 -->
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
          {{ $t('admin.disputeDetail') }} #{{ detail.id }}
        </text>
        <view class="detail-row">
          <text class="d-label">
            {{ $t('admin.status') }}
          </text>
          <text>{{ statusText(detail.status) }}</text>
        </view>
        <view class="detail-row">
          <text class="d-label">
            {{ $t('admin.orderId') }}
          </text>
          <text>#{{ detail.order_id }}</text>
        </view>
        <view class="detail-row">
          <text class="d-label">
            {{ $t('admin.reason') }}
          </text>
          <text>{{ detail.reason }}</text>
        </view>
        <view class="detail-row">
          <text class="d-label">
            {{ $t('admin.claim') }}
          </text>
          <text>{{ detail.claim }}</text>
        </view>
        <view class="detail-row">
          <text class="d-label">
            {{ $t('admin.resolution') }}
          </text>
          <text>{{ detail.resolution_type || '-' }}</text>
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
usePageTitle('page.adminDisputes', { onShow })

const disputes = ref<any[]>([])
const detail = ref<any>(null)

const load = async () => {
  try {
    const res = await request.get('/api/v1/admin/disputes?size=50', { silent401: true })
    disputes.value = res.disputes || []
  } catch {
    disputes.value = []
  }
}

onShow(() => {
  if (userStore.user?.user_type !== 3) {
    uni.reLaunch({ url: '/pages/index/index' })
    return
  }
  load()
})

const viewDetail = (d: any) => {
  detail.value = d
}

const closeDispute = async (d: any) => {
  uni.showModal({
    title: $t('admin.closeDispute'),
    content: $t('admin.closeDisputeConfirm'),
    success: async (res: any) => {
      if (!res.confirm) return
      try {
        await request.post(`/api/v1/dispute/${d.id}/close`, { resolution_type: 'agreement', settle_amount: 0 })
        uni.showToast({ title: $t('common.success'), icon: 'success' })
        load()
      } catch { /* toast 已提示 */ }
    },
  })
}

const statusText = (s: string) => {
  const map: Record<string, string> = {
    evidence: $t('dispute.status.evidence'),
    review: $t('dispute.status.review'),
    mediation: $t('dispute.status.mediation'),
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
.dispute-card {
  background: var(--card-bg, #fff);
  border-radius: 14rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.d-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}
.d-title {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--text-color, #333);
}
.d-status {
  font-size: 24rpx;
}
.st-evidence {
  color: #e6a23c;
}
.st-review {
  color: #2563eb;
}
.st-mediation {
  color: #8b5cf6;
}
.st-closed {
  color: #909399;
}
.d-info {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  font-size: 24rpx;
  color: var(--muted-color, #666);
}
.d-actions {
  margin-top: 14rpx;
}
.btn {
  font-size: 26rpx;
  border-radius: 10rpx;
  line-height: 60rpx;
  height: 60rpx;
  padding: 0 30rpx;
}
.btn-close {
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
  gap: 20rpx;
}
.d-label {
  color: #999;
  flex-shrink: 0;
}
</style>
