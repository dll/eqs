<template>
  <view
    v-if="order"
    class="container"
  >
    <view class="info-card">
      <text class="order-title">
        {{ $t('order.orderNo', { id: order.id }) }}
      </text>
      <text class="order-item">
        {{ $t('order.amountPrefix', { amount: order.amount }) }}
      </text>
      <text class="order-item">
        {{ $t('order.statusPrefix') }}{{ statusText(order.status) }}
      </text>
      <text
        v-if="order.project"
        class="order-item"
      >
        {{ $t('order.projectPrefix', { title: order.project.title }) }}
      </text>
      <text
        v-if="order.signed_at"
        class="order-item"
      >
        {{ $t('order.signedAtPrefix', { time: order.signed_at }) }}
      </text>
    </view>

    <view
      v-if="contract"
      class="contract-card"
    >
      <text class="card-title">
        {{ $t('order.contract') }}
      </text>
      <text class="contract-no">
        {{ $t('order.contractNoPrefix', { no: contract.contract_no }) }}
      </text>
      <text class="contract-status">
        {{ $t('order.contract.status', { status: contractStatusText(contract.status) }) }}
      </text>
      <button
        v-if="contract.status === 'draft'"
        class="mini-btn"
        @tap="signContract"
      >
        {{ $t('order.contract.sign') }}
      </button>
    </view>

    <view class="milestone-card">
      <text class="card-title">
        {{ $t('order.milestones') }}
      </text>
      <view
        v-for="ms in milestones"
        :key="ms.id"
        class="milestone-item"
      >
        <view class="ms-head">
          <text class="ms-name">
            {{ ms.name }}
          </text>
          <text class="ms-ratio">
            {{ ms.ratio }}%
          </text>
        </view>
        <text class="ms-detail">
          {{ $t('order.milestone.amount', { amount: ms.amount }) }} · {{ milestoneStatusText(ms.status) }}
        </text>
        <view
          v-if="ms.status === 'pending'"
          class="ms-actions"
        >
          <button
            class="mini-btn"
            @tap="deliver(ms)"
          >
            {{ $t('order.milestone.deliver') }}
          </button>
          <button
            class="mini-btn primary"
            @tap="settle(ms)"
          >
            {{ $t('order.milestone.settle') }}
          </button>
        </view>
        <view
          v-if="ms.status === 'submitted'"
          class="ms-actions"
        >
          <button
            class="mini-btn primary"
            @tap="accept(ms, true)"
          >
            {{ $t('order.milestone.accept') }}
          </button>
          <button
            class="mini-btn"
            @tap="accept(ms, false)"
          >
            {{ $t('order.milestone.reject') }}
          </button>
        </view>
        <button
          v-if="ms.status === 'accepted'"
          class="mini-btn"
          @tap="settle(ms)"
        >
          {{ $t('order.milestone.submitSettle') }}
        </button>
      </view>
    </view>

    <view class="milestone-card">
      <text class="card-title">
        {{ $t('order.attendance') }}
      </text>
      <button
        class="mini-btn primary checkin-btn"
        @tap="checkin"
      >
        {{ $t('order.checkin') }}
      </button>
      <text
        v-if="checkinStatus"
        class="checkin-status"
      >
        {{ checkinStatus }}
      </text>
      <view
        v-if="attendanceList.length"
        class="att-list"
      >
        <view
          v-for="a in attendanceList"
          :key="a.id"
          class="att-item"
        >
          <text>{{ $t('order.checkinAt', { time: a.check_in_at }) }}</text>
          <text class="att-loc">
            {{ a.longitude }},{{ a.latitude }}
          </text>
        </view>
      </view>
      <text
        v-else
        class="att-empty"
      >
        {{ $t('order.noAttendance') }}
      </text>
    </view>

    <view class="actions-card">
      <button
        v-if="order.status === 0"
        class="action-btn danger"
        @tap="cancelOrder"
      >
        {{ $t('order.cancel') }}
      </button>
      <button
        class="action-btn"
        @tap="openDispute"
      >
        {{ $t('order.dispute') }}
      </button>
      <button
        class="action-btn"
        @tap="goPay"
      >
        {{ $t('order.pay') }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.orderDetail', { onLoad })
const order = ref<any>(null)
const contract = ref<any>(null)
const milestones = ref<any[]>([])
const attendanceList = ref<any[]>([])
const checkinStatus = ref('')

onLoad((options) => {
  if (options?.id) {
    loadOrder(options.id)
  }
})

const loadOrder = async (id: string) => {
  try {
    const res = await request.get(`/api/v1/order/${id}`)
    order.value = res.order
    milestones.value = res.milestones || []
    contract.value = res.contract || null
    loadAttendance()
  } catch {
    // request 已提示
  }
}

const loadAttendance = async () => {
  if (!order.value?.id) return
  try {
    const res = await request.get(`/api/v1/order/${order.value.id}/attendance`, { silent401: true }).catch(() => ({ records: [] }))
    attendanceList.value = res.records || []
  } catch {
    attendanceList.value = []
  }
}

const checkin = () => {
  uni.getLocation({
    type: 'gcj02',
    success: async (loc) => {
      try {
        await request.post('/api/v1/attendance/checkin', {
          order_id: order.value?.id,
          longitude: loc.longitude,
          latitude: loc.latitude,
          distance_meters: 0,
        })
        checkinStatus.value = $t('order.checkinOk')
        loadAttendance()
      } catch {
        checkinStatus.value = $t('order.checkinFail')
      }
    },
    fail: () => {
      checkinStatus.value = $t('order.checkinNoLocation')
    },
  })
}

// V9：取消未签约订单（二次确认后 POST /order/:id/cancel）
const cancelOrder = () => {
  uni.showModal({
    title: $t('order.cancel'),
    content: $t('order.cancelConfirm'),
    success: async (r: any) => {
      if (!r.confirm) return
      try {
        await request.post(`/api/v1/order/${order.value.id}/cancel`, { reason: $t('order.cancelReasonDefault') })
        uni.showToast({ title: $t('order.cancelled'), icon: 'none' })
        loadOrder(String(order.value.id))
      } catch {
        // request 已提示
      }
    },
  })
}

const statusText = (status: any) => {
  const m: Record<number, string> = {
    0: $t('order.pending'), 1: $t('order.inProgress'), 2: $t('order.toAccept'), 3: $t('order.completed'), 4: $t('order.dispute')
  }
  return m[status] || $t('order.unknown')
}

const contractStatusText = (status: string) => {
  const cm: Record<string, string> = {
    draft: $t('order.contract.draft'),
    signing: $t('order.contract.signing'),
    signed: $t('order.contract.signed'),
    voided: $t('order.contract.voided'),
  }
  return cm[status] || status
}

const milestoneStatusText = (status: string) => {
  const mm: Record<string, string> = {
    pending: $t('order.milestone.pending'),
    submitted: $t('order.milestone.submitted'),
    accepted: $t('order.milestone.accepted'),
    settled: $t('order.milestone.settled'),
  }
  return mm[status] || status
}

const signContract = async () => {
  try {
    await request.post(`/api/v1/contract/${contract.value.id}/sign`)
    uni.showToast({ title: $t('order.signSuccess'), icon: 'success' })
    loadOrder(order.value.id)
  } catch {
    // request 已提示
  }
}

const deliver = async (ms: any) => {
  // P1-05：服务方上传交付物（文件 URL 简化版，MVP 先登记）
  uni.chooseImage({
    count: 1,
    success: async (res) => {
      const path = res.tempFilePaths[0]
      const name = path.split('/').pop() || '交付物'
      try {
        await request.post(`/api/v1/milestone/${ms.id}/deliver`, {
          file_name: name,
          file_url: path,
        })
        uni.showToast({ title: $t('order.deliverOk'), icon: 'success' })
        loadOrder(order.value?.id)
      } catch {
        // request 已提示
      }
    },
    fail: () => {
      // 用户取消选择
    },
  })
}

const settle = async (ms: any) => {
  try {
    await request.post(`/api/v1/milestone/${ms.id}/settle`)
    uni.showToast({ title: $t('order.settleSubmitted'), icon: 'success' })
    loadOrder(order.value?.id)
  } catch {
    // request 已提示
  }
}

const accept = async (ms: any, ok: boolean) => {
  try {
    await request.post(`/api/v1/milestone/${ms.id}/accept`, { accept: ok, comment: ok ? $t('order.acceptPassed') : $t('order.acceptRejected') })
    uni.showToast({ title: ok ? $t('order.acceptPassed') : $t('order.acceptRejected'), icon: 'success' })
    loadOrder(order.value?.id)
  } catch {
    // request 已提示
  }
}

const openDispute = async () => {
  // P1-05：发起争议（默认理由，可后续完善）
  if (!order.value?.id) return
  try {
    await request.post('/api/v1/dispute/create', {
      order_id: order.value.id,
      reason: $t('order.disputeDefaultReason'),
      claim: $t('order.disputeDefaultClaim'),
    })
    uni.showToast({ title: $t('order.disputeCreated'), icon: 'success' })
    loadOrder(order.value?.id)
  } catch {
    // request 已提示
  }
}

const goPay = async () => {
  if (!order.value?.id) return
  try {
    await request.post('/api/v1/pay/create', {
      order_id: order.value.id,
      amount: order.value.amount,
      channel: 'mock',
    })
    uni.showToast({ title: $t('order.payOk'), icon: 'success' })
    loadOrder(order.value?.id)
  } catch {
    // request 已提示
  }
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.info-card, .contract-card, .milestones-card, .actions-card {
  background: var(--card-bg);
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.order-title {
  font-size: 36rpx;
  font-weight: bold;
  color: var(--text-color);
  display: block;
  margin-bottom: 15rpx;
}

.order-item, .contract-no, .contract-status {
  font-size: 28rpx;
  color: var(--muted-color);
  display: block;
  margin-bottom: 8rpx;
}

.card-title {
  font-size: 30rpx;
  font-weight: bold;
  color: var(--text-color);
  display: block;
  margin-bottom: 20rpx;
}

.contract-card {
  background: var(--card-bg);
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.milestones-card {
  background: var(--card-bg);
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.milestone-item {
  border-top: 1rpx solid var(--border-color);
  padding: 20rpx 0;
}

.ms-head {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8rpx;
}

.ms-name {
  font-size: 28rpx;
  color: var(--text-color);
}

.ms-ratio {
  font-size: 26rpx;
  color: var(--primary-color);
}

.ms-detail {
  font-size: 24rpx;
  color: var(--muted-color);
  display: block;
  margin-bottom: 12rpx;
}

.ms-actions {
  display: flex;
  gap: 15rpx;
}

.mini-btn {
  background: var(--input-bg);
  border-radius: 8rpx;
  font-size: 24rpx;
  padding: 10rpx 0;
  flex: 1;
}

.mini-btn.primary {
  background: var(--primary-color);
  color: #fff;
}

.actions-card {
  background: var(--card-bg);
  border-radius: 10rpx;
  padding: 30rpx;
  display: flex;
  gap: 20rpx;
}

.action-btn {
  flex: 1;
  background: var(--input-bg);
  border-radius: 10rpx;
  font-size: 28rpx;
}

.checkin-btn {
  margin-bottom: 16rpx;
}

.checkin-status {
  display: block;
  font-size: 26rpx;
  color: var(--primary-color);
  margin-bottom: 10rpx;
}

.att-list {
  border-top: 1rpx solid var(--border-color, #eee);
}

.att-item {
  padding: 14rpx 0;
  border-bottom: 1rpx solid var(--border-color, #eee);
  font-size: 24rpx;
  color: var(--text-color, #333);
}

.att-loc {
  color: var(--muted-color, #999);
  margin-left: 10rpx;
}

.att-empty {
  font-size: 24rpx;
  color: var(--muted-color, #999);
}
</style>