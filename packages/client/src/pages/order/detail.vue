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

    <view class="actions-card">
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
  } catch {
    // request 已提示
  }
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

const deliver = async (_ms: any) => {
  uni.showToast({ title: $t('order.deliverToast'), icon: 'none' })
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

const openDispute = () => {
  uni.showToast({ title: $t('order.disputeToast'), icon: 'none' })
}

const goPay = () => {
  uni.showToast({ title: $t('order.payToast'), icon: 'none' })
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
</style>