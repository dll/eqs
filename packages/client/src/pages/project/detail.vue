<template>
  <view class="container">
    <view class="project-info" v-if="project">
      <view class="info-card">
        <text class="project-title">{{ project.title }}</text>
        <text class="project-type">{{ $t('project.' + project.project_type) || project.project_type }}</text>
        <text class="project-budget">{{ $t('project.budget') }}：¥{{ project.budget_min }} - ¥{{ project.budget_max }}</text>
        <text class="project-status">{{ $t('project.status') }}：{{ statusText(project.status) }}</text>
        <text class="project-desc" v-if="project.description">{{ project.description }}</text>
      </view>

      <view class="theme-card" v-if="isOwner()">
        <text class="card-title">{{ $t('mine.theme') }}</text>
        <view class="theme-options">
          <view
            v-for="item in THEMES"
            :key="item.id"
            :class="['theme-option', projectTheme === item.id ? 'active' : '']"
            @tap="setProjectTheme(item.id)"
          >
            <text class="theme-name">{{ item.name }}</text>
            <text class="theme-desc">{{ item.description }}</text>
          </view>
          <view :class="['theme-option', projectTheme === '' ? 'active' : '']" @tap="setProjectTheme('')">
            <text class="theme-name">{{ $t('common.all') }}</text>
            <text class="theme-desc">跟随系统</text>
          </view>
        </view>
      </view>

      <view class="progress-card">
        <text class="card-title">{{ $t('project.progress') }}</text>
        <view class="timeline">
          <view class="timeline-item" v-for="(step, i) in steps" :key="i" :class="step.done ? 'done' : ''">
            <view class="dot" />
            <text class="step-text">{{ step.label }}</text>
          </view>
        </view>
      </view>

      <view class="action-bar">
        <button class="action-btn primary" v-if="project.status === 1 && isSupplier()" @tap="applyProject">{{ $t('project.apply') }}</button>
        <button class="action-btn" @tap="contact">{{ $t('project.contact') }}</button>
      </view>
    </view>

    <view class="bid-modal" v-if="showBidModal">
      <view class="modal-mask" @tap="showBidModal = false" />
      <view class="modal-card">
        <text class="modal-title">{{ $t('project.bidTitle') }}</text>
        <input class="modal-input" type="digit" v-model.number="bidForm.amount" :placeholder="$t('project.bidAmount')" />
        <input class="modal-input" type="number" v-model.number="bidForm.service_days" :placeholder="$t('project.bidDays')" />
        <view class="modal-actions">
          <button class="modal-btn" @tap="showBidModal = false">{{ $t('common.cancel') }}</button>
          <button class="modal-btn primary" @tap="submitBid">{{ $t('common.submit') }}</button>
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
import { useSettingsStore, THEMES } from '@/store/settings'
import { useI18n } from '@/utils/i18n'

const project = ref<any>(null)
const userStore = useUserStore()
const settingsStore = useSettingsStore()
const { $t } = useI18n()
const projectTheme = ref('')

const steps = ref([
  { label: $t('project.published'), done: true },
  { label: $t('project.assigned'), done: false },
  { label: $t('project.inProgress'), done: false },
  { label: $t('project.completed'), done: false },
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
    projectTheme.value = project.value.theme || ''
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
    0: $t('project.draft'),
    1: $t('project.published'),
    2: $t('project.assigned'),
    3: $t('project.inProgress'),
    4: $t('project.completed'),
  }
  return map[status] || ''
}

const isOwner = () => userStore.user?.id === project.value?.user_id
const isSupplier = () => userStore.user?.user_type === 2

const showBidModal = ref(false)
const bidForm = ref({ amount: 0, service_days: 7 })

const applyProject = () => {
  showBidModal.value = true
}

const setProjectTheme = async (theme: string) => {
  if (!project.value) return
  try {
    await request.put(`/api/v1/project/${project.value.id}/theme`, { theme })
    projectTheme.value = theme
    uni.showToast({ title: $t('common.success'), icon: 'success' })
  } catch {
    // request 已提示
  }
}

const submitBid = async () => {
  if (!bidForm.value.amount || bidForm.value.amount <= 0) {
    uni.showToast({ title: $t('project.bidAmountRequired'), icon: 'none' })
    return
  }
  try {
    await request.post('/api/v1/bid/submit', {
      project_id: project.value.id,
      amount: bidForm.value.amount,
      service_days: bidForm.value.service_days,
    })
    uni.showToast({ title: $t('project.bidSuccess'), icon: 'success' })
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

.info-card, .progress-card, .theme-card {
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

.theme-options {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.theme-option {
  flex: 1;
  min-width: 140rpx;
  padding: 16rpx;
  border: 2rpx solid #e5e5e5;
  border-radius: 10rpx;
  text-align: center;
}

.theme-option.active {
  border-color: #1890ff;
  background: #e6f7ff;
}

.theme-name {
  font-size: 26rpx;
  font-weight: bold;
  display: block;
}

.theme-desc {
  font-size: 22rpx;
  color: #999;
  display: block;
  margin-top: 4rpx;
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
