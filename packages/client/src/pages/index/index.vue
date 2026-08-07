<template>
  <view class="container">
    <view class="search-bar">
      <input class="search-input" placeholder="搜索服务、项目..." />
    </view>

    <view class="categories">
      <view class="category-item" v-for="cat in categories" :key="cat.id" @tap="goToProjectList(cat.type)">
        <image class="category-icon" :src="cat.icon" mode="aspectFit" />
        <text class="category-name">{{ cat.name }}</text>
      </view>
    </view>

    <view class="section">
      <view class="section-header">
        <text class="section-title">推荐项目</text>
        <text class="section-more" @tap="goToProjectList()">查看更多</text>
      </view>
      <view class="project-list">
        <view class="project-card" v-for="project in projects" :key="project.id" @tap="goToProjectDetail(project.id)">
          <view class="project-info">
            <text class="project-title">{{ project.title }}</text>
            <text class="project-type">{{ project.project_type }}</text>
            <text class="project-budget">预算：{{ project.budget_min }}-{{ project.budget_max }}万</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const categories = ref([
  { id: 1, name: '造价咨询', type: '造价', icon: '/static/category/price.png' },
  { id: 2, name: '工程监理', type: '监理', icon: '/static/category/supervise.png' },
  { id: 3, name: '地质勘察', type: '地勘', icon: '/static/category/survey.png' },
  { id: 4, name: '工程设计', type: '设计', icon: '/static/category/design.png' },
])

const projects = ref([])

const goToProjectList = (type?: string) => {
  uni.navigateTo({ url: `/pages/project/list?type=${type || ''}` })
}

const goToProjectDetail = (id: number) => {
  uni.navigateTo({ url: `/pages/project/detail?id=${id}` })
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.search-bar {
  margin-bottom: 30rpx;
}

.search-input {
  background: #fff;
  padding: 20rpx 30rpx;
  border-radius: 10rpx;
  font-size: 28rpx;
}

.categories {
  display: flex;
  justify-content: space-around;
  background: #fff;
  padding: 30rpx 0;
  border-radius: 10rpx;
  margin-bottom: 30rpx;
}

.category-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.category-icon {
  width: 80rpx;
  height: 80rpx;
  margin-bottom: 10rpx;
}

.category-name {
  font-size: 24rpx;
  color: #333;
}

.section {
  background: #fff;
  border-radius: 10rpx;
  padding: 20rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.section-more {
  font-size: 26rpx;
  color: #1890ff;
}

.project-card {
  padding: 20rpx;
  border-bottom: 1rpx solid #eee;
}

.project-title {
  font-size: 30rpx;
  color: #333;
  display: block;
  margin-bottom: 10rpx;
}

.project-type {
  font-size: 24rpx;
  color: #1890ff;
  display: block;
  margin-bottom: 10rpx;
}

.project-budget {
  font-size: 26rpx;
  color: #666;
}
</style>
