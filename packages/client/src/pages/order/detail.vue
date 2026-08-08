<template>
  <view class="container" v-if="order">
    <view class="info-card">
      <text class="order-title">订单 #{{ order.id }}</text>
      <text class="order-item">金额：¥{{ order.amount }}</text>
      <text class="order-item">状态：{{ statusText(order.status) }}</text>
      <text class="order-item" v-if="order.project">项目：{{ order.project.title }}</text>
      <text class="order-item" v-if="order.signed_at">签约时间：{{ order.signed_at }}</text>
    </view>

    <view class="contract-card" v-if="contract">
      <text class="card-title">合同</text>
      <text class="contract-no">合同号：{{ contract.contract_no }}</text>
      <text class="contract-status">状态：{{ contract.status }}</text>
      <button v-if="contract.status === 'draft'" class="mini-btn" @tap="signContract">签署合同</button>
    </view>

    <view class="milestone-card">
      <text class="card-title">付款节点</text>
      <view class="milestone-item" v-for="ms in milestones" :key="ms.id">
        <view class="ms-head">
          <text class="ms-name">{{ ms.name }}</text>
          <text class="ms-ratio">{{ ms.ratio }}%</text>
        </view>
        <text class="ms-detail">金额 ¥{{ ms.amount }} · {{ ms.status }}</text>
        <view class="ms-actions" v-if="ms.status === 'pending'">
          <button class="mini-btn" @tap="deliver(ms)">上传交付</button>
          <button class="mini-btn primary" @tap="settle(ms)">结算</button>
        </view>
        <view class="ms-actions" v-if="ms.status === 'submitted'">
          <button class="mini-btn primary" @tap="accept(ms, true)">验收通过</button>
          <button class="mini-btn" @tap="accept(ms, false)">驳回</button>
        </view>
        <button v-if="ms.status === 'accepted'" class="mini-btn" @tap="settle(ms)">提交结算</button>
      </view>
    </view>

    <view class="actions-card">
      <button class="action-btn" @tap="openDispute">发起争议</button>
      <button class="action-btn" @tap="goPay">资金明细</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '@/utils/request'

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
    0: '待签约', 1: '进行中', 2: '待验收', 3: '已完成', 4: '纠纷中'
  }
  if (typeof status === 'string') {
    const cm: Record<string, string> = { draft: '草稿', signing: '签署中', signed: '已签署', voided: '已作废' }
    return cm[status] || status
  }
  return m[status] || '未知'
}

const signContract = async () => {
  try {
    await request.post(`/api/v1/contract/${contract.value.id}/sign`)
    uni.showToast({ title: '签署完成', icon: 'success' })
    loadOrder(order.value.id)
  } catch {
    // request 已提示
  }
}

const deliver = async (ms: any) => {
  uni.showToast({ title: '请通过文件页面上传', icon: 'none' })
}

const settle = async (ms: any) => {
  try {
    await request.post(`/api/v1/milestone/${ms.id}/settle`)
    uni.showToast({ title: '结算指令已提交', icon: 'success' })
    loadOrder(order.value?.id)
  } catch {
    // request 已提示
  }
}

const accept = async (ms: any, ok: boolean) => {
  try {
    await request.post(`/api/v1/milestone/${ms.id}/accept`, { accept: ok, comment: ok ? '验收通过' : '驳回' })
    uni.showToast({ title: ok ? '验收通过' : '已驳回', icon: 'success' })
    loadOrder(order.value?.id)
  } catch {
    // request 已提示
  }
}

const openDispute = () => {
  uni.showToast({ title: '争议处理详见详情', icon: 'none' })
}

const goPay = () => {
  uni.showToast({ title: '资金明细为记账流水', icon: 'none' })
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.info-card, .contract-card, .milestones-card, .actions-card {
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

.order-item, .contract-no, .contract-status {
  font-size: 28rpx;
  color: #666;
  display: block;
  margin-bottom: 8rpx;
}

.card-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 20rpx;
}

.contract-card {
  background: #fff;
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.milestones-card {
  background: #fff;
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.milestone-item {
  border-top: 1rpx solid #eee;
  padding: 20rpx 0;
}

.ms-head {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8rpx;
}

.ms-name {
  font-size: 28rpx;
  color: #333;
}

.ms-ratio {
  font-size: 26rpx;
  color: #1890ff;
}

.ms-detail {
  font-size: 24rpx;
  color: #999;
  display: block;
  margin-bottom: 12rpx;
}

.ms-actions {
  display: flex;
  gap: 15rpx;
}

.mini-btn {
  background: #f5f5f5;
  border-radius: 8rpx;
  font-size: 24rpx;
  padding: 10rpx 0;
  flex: 1;
}

.mini-btn.primary {
  background: #1890ff;
  color: #fff;
}

.actions-card {
  background: #fff;
  border-radius: 10rpx;
  padding: 30rpx;
  display: flex;
  gap: 20rpx;
}

.action-btn {
  flex: 1;
  background: #f5f5f5;
  border-radius: 10rpx;
  font-size: 28rpx;
}
</style>