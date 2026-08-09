<template>
  <view class="container">
    <view class="search-bar">
      <input class="search-input" :placeholder="$t('home.searchPlaceholder')" />
    </view>

    <view class="categories">
      <view class="category-item" v-for="cat in categories" :key="cat.id" @tap="goToProjectList(cat.type)">
        <image class="category-icon" :src="cat.icon" mode="aspectFit" />
        <text class="category-name">{{ $t(cat.nameKey) }}</text>
      </view>
    </view>

    <view class="section">
      <view class="section-header">
        <text class="section-title">{{ $t('home.recommended') }}</text>
        <text class="section-more" @tap="goToProjectList()">{{ $t('home.viewMore') }}</text>
      </view>
      <view class="project-list">
        <view class="project-card" v-for="project in projects" :key="project.id" @tap="goToProjectDetail(project.id)">
          <view class="project-info">
            <text class="project-title">{{ project.title }}</text>
            <text class="project-type">{{ toTypeKey(project.project_type) }}</text>
            <text class="project-budget">{{ $t('home.budget', { min: toWan(project.budget_min), max: toWan(project.budget_max) }) }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle, applyTabBarI18n } from '@/utils/i18n'
import { toWan, toTypeKey } from '@/lib/service'

const { $t } = useI18n()
usePageTitle('page.home', { onShow })

const categories = ref([
  { id: 1, nameKey: 'category.cost', type: 'cost', icon: '/static/category/price.png' },
  { id: 2, nameKey: 'category.supervision', type: 'supervision', icon: '/static/category/supervise.png' },
  { id: 3, nameKey: 'category.geotech', type: 'geotech', icon: '/static/category/survey.png' },
  { id: 4, nameKey: 'category.design', type: 'design', icon: '/static/category/design.png' },
])

const projects = ref<any[]>([])

const goToProjectList = (type?: string) => {
  uni.navigateTo({ url: `/pages/project/list?type=${type || ''}` })
}

const goToProjectDetail = (id: number) => {
  uni.navigateTo({ url: `/pages/project/detail?id=${id}` })
}

onShow(() => {
  applyTabBarI18n()
  loadProjects()
})

const loadProjects = async () => {
  try {
    const res = await request.get('/api/v1/project/list')
    projects.value = (res.projects || []).filter((p: any) => p.status === 1).slice(0, 5)
  } catch {
    // request 已提示
  }
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
  background: var(--card-bg);
  padding: 20rpx 30rpx;
  border-radius: 10rpx;
  font-size: 28rpx;
}

.categories {
  display: flex;
  justify-content: space-around;
  background: var(--card-bg);
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
  color: var(--text-color);
}

.section {
  background: var(--card-bg);
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
  color: var(--text-color);
}

.section-more {
  font-size: 26rpx;
  color: var(--primary-color);
}

.project-card {
  padding: 20rpx;
  border-bottom: 1rpx solid var(--border-color);
}

.project-title {
  font-size: 30rpx;
  color: var(--text-color);
  display: block;
  margin-bottom: 10rpx;
}

.project-type {
  font-size: 24rpx;
  color: var(--primary-color);
  display: block;
  margin-bottom: 10rpx;
}

.project-budget {
  font-size: 26rpx;
  color: var(--muted-color);
}
</style>
