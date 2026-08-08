<template>
  <view class="container">
    <view class="project-info" v-if="project">
      <view class="info-card">
        <text class="project-title">{{ project.title }}</text>
        <text class="project-type">{{ project.project_type }}</text>
        <text class="project-budget">预算：¥{{ project.budget_min }} - ¥{{ project.budget_max }}</text>
        <text class="project-status">状态：{{ statusText(project.status) }}</text>
        <text class="project-desc" v-if="project.description">{{ project.description }}</text>
      </view>

      <view class="progress-card">
        <text class="card-title">项目进度</text>
        <view class="timeline">
          <view class="timeline-item" v-for="(step, i) in steps" :key="i" :class="step.done ? 'done' : ''">
            <view class="dot" />
            <text class="step-text">{{ step.label }}</text>
          </view>
        </view>
      </view>

      <view class="action-bar">
        <button class="action-btn primary" v-if="project.status === 1 && isSupplier()" @tap="applyProject">报名</button>
        <button class="action-btn" @tap="contact">联系</button>
      </view>
    </view>

    <view class="bid-modal" v-if="showBidModal">
      <view class="modal-mask" @tap="showBidModal = false" />
      <view class="modal-card">
        <text class="modal-title">提交报价</text>
        <input class="modal-input" type="digit" v-model.number="bidForm.amount" placeholder="报价金额（元）" />
        <input class="modal-input" type="number" v-model.number="bidForm.service_days" placeholder="服务天数" />
        <view class="modal-actions">
          <button class="modal-btn" @tap="showBidModal = false">取消</button>
          <button class="modal-btn primary" @tap="submitBid">提交</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useUserStore } from '@/store/user'

const project = ref<any>(null)
const userStore = useUserStore()

const steps = ref([
  { label: '已发布', done: true },
  { label: '已接单', done: false },
  { label: '进行中', done: false },
  { label: '已完成', done: false },
])

onLoad((options) => {
  if (options?.id) {
    loadProject(options.id)
  }
})

const loadProject = async (id: string) => {
  try {
    const res = await request.get(`/api/v1/project/${id}`)
    project.value = res.project
    steps.value[0].done = project.value.status >= 1
    steps.value[1].done = project.value.status >= 2
    steps.value[2].done = project.value.status >= 3
    steps.value[3].done = project.value.status >= 4
  } catch {
    // request 已提示
  }
}

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: '草稿', 1: '已发布', 2: '已接单', 3: '进行中', 4: '已完成'
  }
  return map[status] || '未知'
}

const isSupplier = () => userStore.user?.user_type === 2

const showBidModal = ref(false)
const bidForm = ref({ amount: 0, service_days: 7 })

const applyProject = () => {
  showBidModal.value = true
}

const submitBid = async () => {
  if (!bidForm.value.amount || bidForm.value.amount <= 0) {
    uni.showToast({ title: '请填写报价金额', icon: 'none' })
    return
  }
  try {
    await request.post('/api/v1/bid/submit', {
      project_id: project.value.id,
      amount: bidForm.value.amount,
      service_days: bidForm.value.service_days,
    })
    uni.showToast({ title: '报价成功', icon: 'success' })
    showBidModal.value = false
  } catch {
    // request 已提示
  }
}

const contact = () => {
  uni.showToast({ title: '暂未开放站内信', icon: 'none' })
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.info-card, .progress-card {
  background: #fff;
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.project-title {
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 15rpx;
}

.project-type, .project-budget, .project-status, .project-desc {
  font-size: 28rpx;
  color: #666;
  display: block;
  margin-bottom: 10rpx;
}

.card-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 20rpx;
}

.timeline {
  padding-left: 20rpx;
}

.timeline-item {
  display: flex;
  align-items: center;
  margin-bottom: 20rpx;
}

.dot {
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  background: #ddd;
  margin-right: 15rpx;
}

.timeline-item.done .dot {
  background: #1890ff;
}

.step-text {
  font-size: 26rpx;
  color: #666;
}

.action-bar {
  display: flex;
  gap: 20rpx;
  margin-top: 30rpx;
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

.bid-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 100;
}

.bid-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
}

.modal-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
}

.modal-card {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: #fff;
  border-radius: 20rpx 20rpx 0 0;
  padding: 40rpx;
}

.bid-modal .modal-card {
  bottom: 20rpx;
  left: 20rpx;
  right: 20rpx;
  border-radius: 20rpx;
}

.modal-title {
  font-size: 32rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 30rpx;
}

.modal-input {
  background: #f5f5f5;
  padding: 24rpx;
  border-radius: 10rpx;
  margin-bottom: 20rpx;
  font-size: 30rpx;
}

.modal-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 20rpx;
}

.modal-btn {
  flex: 1;
  background: #f5f5f5;
  border-radius: 10rpx;
  font-size: 28rpx;
}

.modal-btn.primary {
  background: #1890ff;
  color: #fff;
}
</style>