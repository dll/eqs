<template>
  <view class="container">
    <view
      v-if="project"
      class="project-info"
    >
      <view class="info-card">
        <text class="project-title">
          {{ project.title }}
        </text>
        <text class="project-type">
          {{ $t('project.' + project.project_type) || project.project_type }}
        </text>
        <text class="project-budget">
          {{ $t('project.budget') }}：¥{{ project.budget_min }} - ¥{{ project.budget_max }}
        </text>
        <text class="project-status">
          {{ $t('project.status') }}：{{ statusText(project.status) }}
        </text>
        <text
          v-if="project.description"
          class="project-desc"
        >
          {{ project.description }}
        </text>
      </view>

      <view
        v-if="isOwner()"
        class="theme-card"
      >
        <text class="card-title">
          {{ $t('mine.theme') }}
        </text>
        <view class="theme-options">
          <view
            v-for="item in THEMES"
            :key="item.id"
            :class="['theme-option', projectTheme === item.id ? 'active' : '']"
            @tap="setProjectTheme(item.id)"
          >
            <text class="theme-name">
              {{ item.name }}
            </text>
            <text class="theme-desc">
              {{ item.description }}
            </text>
          </view>
          <view
            :class="['theme-option', projectTheme === '' ? 'active' : '']"
            @tap="setProjectTheme('')"
          >
            <text class="theme-name">
              {{ $t('common.all') }}
            </text>
            <text class="theme-desc">
              {{ $t('project.themeFollow') }}
            </text>
          </view>
        </view>
      </view>

      <view class="progress-card">
        <text class="card-title">
          {{ $t('project.progress') }}
        </text>
        <view class="timeline">
          <view
            v-for="(step, i) in steps"
            :key="i"
            class="timeline-item"
            :class="step.done ? 'done' : ''"
          >
            <view class="dot" />
            <text class="step-text">
              {{ step.label }}
            </text>
          </view>
        </view>
      </view>

      <view class="action-bar">
        <button
          v-if="project.status === 1 && isSupplier()"
          class="action-btn primary"
          @tap="applyProject"
        >
          {{ $t('project.apply') }}
        </button>
        <button
          v-if="isOwner() && (project.status === 0 || project.status === 1)"
          class="action-btn"
          @tap="editProject"
        >
          {{ $t('project.edit') }}
        </button>
        <button
          v-if="isOwner() && (project.status === 0 || project.status === 1)"
          class="action-btn danger"
          @tap="deleteProject"
        >
          {{ $t('project.delete') }}
        </button>
        <button
          class="action-btn"
          @tap="contact"
        >
          {{ $t('project.contact') }}
        </button>
      </view>
    </view>

    <view
      v-if="showBidModal"
      class="bid-modal"
    >
      <view
        class="modal-mask"
        @tap="showBidModal = false"
      />
      <view class="modal-card">
        <text class="modal-title">
          {{ $t('project.bidTitle') }}
        </text>
        <input
          v-model.number="bidForm.amount"
          class="modal-input"
          type="digit"
          :placeholder="$t('project.bidAmount')"
        >
        <input
          v-model.number="bidForm.service_days"
          class="modal-input"
          type="number"
          :placeholder="$t('project.bidDays')"
        >
        <view class="modal-actions">
          <button
            class="modal-btn"
            @tap="showBidModal = false"
          >
            {{ $t('common.cancel') }}
          </button>
          <button
            class="modal-btn primary"
            @tap="submitBid"
          >
            {{ $t('common.submit') }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useUserStore } from '@/store/user'
import { useSettingsStore, THEMES } from '@/store/settings'
import { useI18n, usePageTitle } from '@/utils/i18n'

const project = ref<any>(null)
usePageTitle('page.projectDetail', { onLoad })

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

// 离开页面恢复用户级主题
onUnload(() => {
  settingsStore.applyTheme(settingsStore.theme)
})

const loadProject = async (id: string) => {
  try {
    const res = await request.get(`/api/v1/project/${id}`)
    project.value = res.project
    projectTheme.value = project.value.theme || ''
    // 应用项目维度主题（覆盖用户主题）
    settingsStore.applyProjectTheme(projectTheme.value)
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

// V10：编辑项目（跳转发布页编辑模式，回填表单后 PUT）
const editProject = () => {
  uni.navigateTo({ url: `/pages/project/create?id=${project.value.id}` })
}

// V9：删除/下架项目（无业务往来物理删除，否则下架）
const deleteProject = () => {
  uni.showModal({
    title: $t('project.delete'),
    content: $t('project.deleteConfirm'),
    success: async (r: any) => {
      if (!r.confirm) return
      try {
        const res = await request.delete(`/api/v1/project/${project.value.id}`)
        uni.showToast({ title: (res as any)?.offline ? $t('project.offlined') : $t('project.deleted'), icon: 'none' })
        setTimeout(() => uni.navigateBack(), 800)
      } catch {
        // request 已提示
      }
    },
  })
}

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
    settingsStore.applyProjectTheme(theme)
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
  uni.showToast({ title: $t('project.contactClose'), icon: 'none' })
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.info-card, .progress-card, .theme-card {
  background: var(--card-bg);
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.project-title {
  font-size: 36rpx;
  font-weight: bold;
  color: var(--text-color);
  display: block;
  margin-bottom: 15rpx;
}

.project-type, .project-budget, .project-status, .project-desc {
  font-size: 28rpx;
  color: var(--muted-color);
  display: block;
  margin-bottom: 10rpx;
}

.card-title {
  font-size: 30rpx;
  font-weight: bold;
  color: var(--text-color);
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
  border: 2rpx solid var(--border-color);
  border-radius: 10rpx;
  text-align: center;
}

.theme-option.active {
  border-color: var(--primary-color);
  background: #e6f7ff;
}

.theme-name {
  font-size: 26rpx;
  font-weight: bold;
  display: block;
}

.theme-desc {
  font-size: 22rpx;
  color: var(--muted-color);
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
  background: var(--primary-color);
}

.step-text {
  font-size: 26rpx;
  color: var(--muted-color);
}

.action-bar {
  display: flex;
  gap: 20rpx;
  margin-top: 30rpx;
}

.action-btn {
  flex: 1;
  background: var(--input-bg);
  border-radius: 10rpx;
  font-size: 28rpx;
}

.action-btn.primary {
  background: var(--primary-color);
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
  background: var(--card-bg);
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
  background: var(--input-bg);
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
  background: var(--input-bg);
  border-radius: 10rpx;
  font-size: 28rpx;
}

.modal-btn.primary {
  background: var(--primary-color);
  color: #fff;
}
</style>
