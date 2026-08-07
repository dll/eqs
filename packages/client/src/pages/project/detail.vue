<template>
  <view class="container">
    <view class="project-info" v-if="project">
      <view class="info-card">
        <text class="project-title">{{ project.title }}</text>
        <text class="project-type">{{ project.project_type }}</text>
        <text class="project-budget">预算：{{ project.budget_min }}-{{ project.budget_max }}万</text>
        <text class="project-status">状态：{{ statusText(project.status) }}</text>
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
        <button class="action-btn primary" v-if="project.status === 1" @tap="applyProject">报名</button>
        <button class="action-btn" @tap="contact">联系</button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const project = ref<any>(null)

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
  // TODO: Call API
  project.value = {
    id,
    title: '示例项目',
    project_type: '造价咨询',
    budget_min: 10,
    budget_max: 50,
    status: 1,
  }
}

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: '草稿', 1: '已发布', 2: '已接单', 3: '进行中', 4: '已完成'
  }
  return map[status] || '未知'
}

const applyProject = () => {
  uni.showToast({ title: '报名成功', icon: 'success' })
}

const contact = () => {
  // TODO: Open chat
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

.project-type, .project-budget, .project-status {
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
</style>
