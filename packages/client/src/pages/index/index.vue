<template>
  <view class="container">
    <!-- 品牌头图：工程蓝→科技青渐变 + AI 光晕 -->
    <view class="hero">
      <view class="hero-glow hero-glow-blue" />
      <view class="hero-glow hero-glow-ai" />
      <text class="hero-title">
        {{ $t('app.title') }}
      </text>
      <text class="hero-sub">
        Agile · AI · Engineering Service
      </text>
      <view class="search-bar">
        <text class="search-icon">🔍</text>
        <input
          class="search-input"
          :placeholder="$t('home.searchPlaceholder')"
        >
      </view>
    </view>

    <view class="categories">
      <view
        v-for="cat in categories"
        :key="cat.id"
        class="category-item"
        @tap="goToProjectList(cat.type)"
      >
        <view class="category-icon-wrap">
          <image
            class="category-icon"
            :src="cat.icon"
            mode="aspectFit"
          />
        </view>
        <text class="category-name">
          {{ $t(cat.nameKey) }}
        </text>
      </view>
    </view>

    <view class="section">
      <view class="section-header">
        <text class="section-title">
          {{ $t('home.recommended') }}
        </text>
        <text
          class="section-more"
          @tap="goToProjectList()"
        >
          {{ $t('home.viewMore') }} →
        </text>
      </view>
      <view class="project-list">
        <view
          v-for="project in projects"
          :key="project.id"
          class="project-card"
          @tap="goToProjectDetail(project.id)"
        >
          <view class="project-info">
            <view class="project-top">
              <text class="project-title">
                {{ project.title }}
              </text>
              <text class="project-type">
                {{ toTypeKey(project.project_type) }}
              </text>
            </view>
            <text class="project-budget">
              💰 {{ $t('home.budget', { min: toWan(project.budget_min), max: toWan(project.budget_max) }) }}
            </text>
          </view>
          <view class="project-arrow">
            →
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

/* 品牌头图 */
.hero {
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #2563eb 0%, #06b6d4 100%);
  border-radius: 20rpx;
  padding: 44rpx 32rpx 36rpx;
  margin-bottom: 24rpx;
}
.hero-glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(30rpx);
  opacity: .4;
}
.hero-glow-blue {
  width: 200rpx;
  height: 200rpx;
  background: #8b5cf6;
  top: -60rpx;
  right: -40rpx;
}
.hero-glow-ai {
  width: 160rpx;
  height: 160rpx;
  background: #06b6d4;
  bottom: -50rpx;
  left: -30rpx;
}
.hero-title {
  display: block;
  color: #fff;
  font-size: 40rpx;
  font-weight: 800;
  letter-spacing: 1rpx;
  position: relative;
  z-index: 1;
}
.hero-sub {
  display: block;
  color: rgba(255, 255, 255, .8);
  font-size: 22rpx;
  letter-spacing: 1rpx;
  margin-top: 8rpx;
  position: relative;
  z-index: 1;
}
.search-bar {
  display: flex;
  align-items: center;
  gap: 12rpx;
  background: rgba(255, 255, 255, .92);
  border-radius: 14rpx;
  padding: 14rpx 22rpx;
  margin-top: 24rpx;
  position: relative;
  z-index: 1;
}
.search-icon {
  font-size: 26rpx;
}
.search-input {
  flex: 1;
  font-size: 28rpx;
  color: var(--text-color);
}

/* 分类 */
.categories {
  display: flex;
  justify-content: space-around;
  background: var(--card-bg);
  padding: 28rpx 0;
  border-radius: 16rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(30, 41, 59, .05);
}
.category-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10rpx;
}
.category-icon-wrap {
  width: 88rpx;
  height: 88rpx;
  border-radius: 22rpx;
  background: var(--bg-color);
  display: flex;
  align-items: center;
  justify-content: center;
}
.category-icon {
  width: 52rpx;
  height: 52rpx;
}
.category-name {
  font-size: 24rpx;
  color: var(--text-color);
  font-weight: 500;
}

/* 推荐项目 */
.section {
  background: var(--card-bg);
  border-radius: 16rpx;
  padding: 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(30, 41, 59, .05);
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}
.section-title {
  font-size: 32rpx;
  font-weight: 700;
  color: var(--text-color);
}
.section-more {
  font-size: 24rpx;
  color: var(--primary-color);
}
.project-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 22rpx 4rpx;
  border-bottom: 1rpx solid var(--border-color);
}
.project-card:last-child {
  border-bottom: none;
}
.project-info {
  flex: 1;
}
.project-top {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 10rpx;
}
.project-title {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--text-color);
  flex: 1;
}
.project-type {
  font-size: 22rpx;
  color: var(--primary-color);
  background: color-mix(in srgb, var(--primary-color) 10%, transparent);
  padding: 4rpx 14rpx;
  border-radius: 20rpx;
}
.project-budget {
  font-size: 24rpx;
  color: var(--muted-color);
}
.project-arrow {
  color: var(--primary-color);
  font-size: 30rpx;
}
</style>
