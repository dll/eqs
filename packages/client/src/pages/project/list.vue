<template>
  <view class="container">
    <view class="filter-bar">
      <view :class="['filter-item', activeType === '' ? 'active' : '']" @tap="filterByType('')">全部</view>
      <view :class="['filter-item', activeType === '造价' ? 'active' : '']" @tap="filterByType('造价')">造价</view>
      <view :class="['filter-item', activeType === '监理' ? 'active' : '']" @tap="filterByType('监理')">监理</view>
      <view :class="['filter-item', activeType === '地勘' ? 'active' : '']" @tap="filterByType('地勘')">地勘</view>
      <view :class="['filter-item', activeType === '设计' ? 'active' : '']" @tap="filterByType('设计')">设计</view>
    </view>

    <view class="project-list">
      <view class="project-card" v-for="project in projects" :key="project.id" @tap="goToDetail(project.id)">
        <view class="project-header">
          <text class="project-title">{{ project.title }}</text>
          <text class="project-status">{{ statusText(project.status) }}</text>
        </view>
        <view class="project-info">
          <text class="info-item">类型：{{ project.project_type }}</text>
          <text class="info-item">预算：{{ project.budget_min }}-{{ project.budget_max }}万</text>
        </view>
        <view class="project-footer">
          <text class="publish-time">{{ project.publish_time }}</text>
        </view>
      </view>
    </view>

    <view class="empty" v-if="projects.length === 0">
      <text>暂无项目</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const activeType = ref('')
const projects = ref([])

onLoad((options) => {
  if (options?.type) {
    activeType.value = options.type
  }
  loadProjects()
})

const loadProjects = async () => {
  // TODO: Call API to load projects
}

const filterByType = (type: string) => {
  activeType.value = type
  loadProjects()
}

const goToDetail = (id: number) => {
  uni.navigateTo({ url: `/pages/project/detail?id=${id}` })
}

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: '草稿', 1: '已发布', 2: '已接单', 3: '进行中', 4: '已完成'
  }
  return map[status] || '未知'
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.filter-bar {
  display: flex;
  background: #fff;
  padding: 20rpx;
  border-radius: 10rpx;
  margin-bottom: 20rpx;
}

.filter-item {
  flex: 1;
  text-align: center;
  font-size: 28rpx;
  color: #666;
  padding: 10rpx 0;
}

.filter-item.active {
  color: #1890ff;
  font-weight: bold;
}

.project-card {
  background: #fff;
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.project-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 15rpx;
}

.project-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
}

.project-status {
  font-size: 24rpx;
  color: #1890ff;
}

.project-info {
  margin-bottom: 15rpx;
}

.info-item {
  font-size: 26rpx;
  color: #666;
  display: block;
  margin-bottom: 5rpx;
}

.project-footer {
  border-top: 1rpx solid #eee;
  padding-top: 15rpx;
}

.publish-time {
  font-size: 24rpx;
  color: #999;
}

.empty {
  text-align: center;
  padding: 100rpx 0;
  color: #999;
}
</style>
