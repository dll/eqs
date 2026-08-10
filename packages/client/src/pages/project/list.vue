<template>
  <view class="container">
    <view class="filter-bar">
      <view
        :class="['filter-item', activeType === '' ? 'active' : '']"
        @tap="filterByType('')"
      >
        {{ $t('common.all') }}
      </view>
      <view
        :class="['filter-item', activeType === 'cost' ? 'active' : '']"
        @tap="filterByType('cost')"
      >
        {{ $t('project.cost') }}
      </view>
      <view
        :class="['filter-item', activeType === 'supervision' ? 'active' : '']"
        @tap="filterByType('supervision')"
      >
        {{ $t('project.supervision') }}
      </view>
      <view
        :class="['filter-item', activeType === 'geotech' ? 'active' : '']"
        @tap="filterByType('geotech')"
      >
        {{ $t('project.geotech') }}
      </view>
      <view
        :class="['filter-item', activeType === 'design' ? 'active' : '']"
        @tap="filterByType('design')"
      >
        {{ $t('project.design') }}
      </view>
    </view>

    <view class="project-list">
      <view
        v-for="project in projects"
        :key="project.id"
        class="project-card"
        @tap="goToDetail(project.id)"
      >
        <view class="project-header">
          <text class="project-title">
            {{ project.title }}
          </text>
          <text class="project-status">
            {{ statusText(project.status) }}
          </text>
        </view>
        <view class="project-info">
          <text class="info-item">
            {{ $t('project.typePrefix') }}{{ toTypeLabel(project.service_type) }}
          </text>
          <text class="info-item">
            {{ $t('project.publisher') }}:{{ project.user?.company_name || $t('project.owner') }}
          </text>
          <text class="info-item">
            {{ $t('project.budget') }}:{{ formatPriceWan(project.budget_min) }} - {{ formatPriceWan(project.budget_max) }}
          </text>
        </view>
        <view class="project-footer">
          <text class="publish-time">
            {{ project.publish_time }}
          </text>
        </view>
      </view>
    </view>

    <view
      v-if="projects.length === 0"
      class="empty"
    >
      <text>{{ $t('common.empty') }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle, applyTabBarI18n } from '@/utils/i18n'
import { toTypeName, formatPriceWan } from '@/lib/service'

const { $t } = useI18n()
usePageTitle('page.projectList', { onShow })

const activeType = ref('')
const projects = ref<any[]>([])
const loading = ref(false)

onLoad((options) => {
  if (options?.type) {
    activeType.value = options.type
  }
})

onShow(() => {
  applyTabBarI18n()
  loadProjects()
})

const loadProjects = async () => {
  loading.value = true
  try {
    const params = activeType.value ? `?service_type=${activeType.value}` : ''
    const res = await request.get(`/api/v1/project/list${params}`)
    projects.value = res.projects || []
  } finally {
    loading.value = false
  }
}

const filterByType = (type: string) => {
  activeType.value = type
  loadProjects()
}

const goToDetail = (id: number) => {
  uni.navigateTo({ url: `/pages/project/detail?id=${id}` })
}

const toTypeLabel = (code: string) => $t('project.' + code) || toTypeName(code)

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: $t('project.draft'), 1: $t('project.published'), 2: $t('project.assigned'), 3: $t('project.inProgress'), 4: $t('project.completed')
  }
  return map[status] || $t('order.unknown')
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.filter-bar {
  display: flex;
  background: var(--card-bg);
  padding: 20rpx;
  border-radius: 10rpx;
  margin-bottom: 20rpx;
}

.filter-item {
  flex: 1;
  text-align: center;
  font-size: 28rpx;
  color: var(--muted-color);
  padding: 10rpx 0;
}

.filter-item.active {
  color: var(--primary-color);
  font-weight: bold;
}

.project-card {
  background: var(--card-bg);
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
  color: var(--text-color);
}

.project-status {
  font-size: 24rpx;
  color: var(--primary-color);
}

.project-info {
  margin-bottom: 15rpx;
}

.info-item {
  font-size: 26rpx;
  color: var(--muted-color);
  display: block;
  margin-bottom: 5rpx;
}

.project-footer {
  border-top: 1rpx solid var(--border-color);
  padding-top: 15rpx;
}

.publish-time {
  font-size: 24rpx;
  color: var(--muted-color);
}

.empty {
  text-align: center;
  padding: 100rpx 0;
  color: var(--muted-color);
}
</style>