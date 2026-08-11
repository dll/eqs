<template>
  <view class="container">
    <view
      v-if="projects.length"
      class="project-list"
    >
      <view
        v-for="p in projects"
        :key="p.id"
        class="project-card"
        @tap="goDetail(p.id)"
      >
        <view class="p-head">
          <text class="p-title">
            {{ p.title }}
          </text>
          <text class="p-status">
            {{ statusText(p.status) }}
          </text>
        </view>
        <text class="p-type">
          {{ p.service_type || p.project_type }}
        </text>
        <text class="p-budget">
          ¥{{ p.budget_min }} - ¥{{ p.budget_max }}
        </text>
      </view>
      <view
        v-if="hasMore"
        class="load-more"
        @tap="loadMore"
      >
        加载更多
      </view>
    </view>
    <view
      v-else
      class="empty"
    >
      暂无发单记录
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.projectMine', { onShow })

const projects = ref<any[]>([])
const page = ref(1)
const hasMore = ref(true)

const load = async (append = false) => {
  try {
    const res = await request.get(`/api/v1/project/mine?page=${page.value}&size=20`)
    const list = res.projects || []
    projects.value = append ? [...projects.value, ...list] : list
    hasMore.value = list.length >= 20
  } catch {
    // request 已提示
  }
}

onShow(() => { page.value = 1; load() })
onReachBottom(() => {
  if (!hasMore.value) return
  page.value++
  load(true)
})

const loadMore = () => {
  if (!hasMore.value) return
  page.value++
  load(true)
}

const goDetail = (id: number) => uni.navigateTo({ url: `/pages/project/detail?id=${id}` })

const statusText = (s: number) => {
  const map: Record<number, string> = {
    0: $t('project.draft'), 1: $t('project.published'), 2: $t('project.assigned'),
    3: $t('project.inProgress'), 4: $t('project.completed'), 5: $t('project.offline'),
  }
  return map[s] || s
}
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.project-card {
  background: var(--card-bg, #fff);
  border-radius: 14rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.p-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10rpx;
}
.p-title {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--text-color, #333);
  flex: 1;
}
.p-status {
  font-size: 22rpx;
  color: var(--primary-color);
  background: var(--input-bg, #f5f5f5);
  padding: 4rpx 14rpx;
  border-radius: 20rpx;
}
.p-type {
  font-size: 24rpx;
  color: var(--muted-color, #666);
  display: block;
  margin-bottom: 8rpx;
}
.p-budget {
  font-size: 26rpx;
  color: var(--text-color, #333);
}
.empty {
  text-align: center;
  color: var(--muted-color, #999);
  padding: 120rpx 0;
  font-size: 26rpx;
}
.load-more {
  text-align: center;
  color: var(--primary-color);
  padding: 20rpx;
  font-size: 26rpx;
}
</style>
