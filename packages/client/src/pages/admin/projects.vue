<template>
  <view class="container">
    <view
      v-for="p in projects"
      :key="p.id"
      class="project-card"
    >
      <view class="p-head">
        <text class="p-title">
          #{{ p.id }} {{ p.title }}
        </text>
        <text class="p-status">
          {{ statusText(p.status) }}
        </text>
      </view>
      <view class="p-info">
        <text>{{ $t('admin.serviceType') }}：{{ toTypeName(p.service_type) }}</text>
        <text>{{ $t('admin.owner') }}：{{ p.user_id }}｜{{ $t('admin.budget') }}：¥{{ p.budget_min }} - ¥{{ p.budget_max }}</text>
      </view>
      <view class="p-actions">
        <button
          class="btn btn-change"
          @tap="changeProject(p)"
        >
          {{ $t('admin.change') }}
        </button>
        <button
          class="btn btn-warn"
          @tap="withdrawProject(p)"
        >
          {{ $t('admin.withdraw') }}
        </button>
        <button
          class="btn btn-danger"
          @tap="abolishProject(p)"
        >
          {{ $t('admin.abolish') }}
        </button>
      </view>
    </view>

    <view
      v-if="!projects.length"
      class="empty"
    >
      {{ $t('admin.noData') }}
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '@/store/user'
import { useI18n, usePageTitle } from '@/utils/i18n'
import { request } from '@/utils/request'
import { toTypeName } from '@/lib/service'

const userStore = useUserStore()
const { $t } = useI18n()
usePageTitle('page.adminProjects', { onShow })

const projects = ref<any[]>([])

const load = async () => {
  try {
    const res = await request.get('/api/v1/project/list', { silent401: true })
    projects.value = res.projects || []
  } catch {
    projects.value = []
  }
}

onShow(() => {
  if (userStore.user?.user_type !== 3) {
    uni.reLaunch({ url: '/pages/index/index' })
    return
  }
  load()
})

const changeProject = (p: any) => {
  uni.navigateTo({ url: `/pages/project/detail?id=${p.id}` })
}

const withdrawProject = async (p: any) => {
  uni.showModal({
    title: $t('admin.withdraw'),
    content: $t('admin.withdrawConfirm'),
    success: async (res: any) => {
      if (!res.confirm) return
      try {
        await request.put(`/api/v1/project/${p.id}/withdraw`)
        uni.showToast({ title: $t('common.success'), icon: 'success' })
        load()
      } catch { /* toast 已提示 */ }
    },
  })
}

const abolishProject = async (p: any) => {
  uni.showModal({
    title: $t('admin.abolish'),
    content: $t('admin.abolishConfirm'),
    success: async (res: any) => {
      if (!res.confirm) return
      try {
        await request.put(`/api/v1/project/${p.id}/abolish`)
        uni.showToast({ title: $t('common.success'), icon: 'success' })
        load()
      } catch { /* toast 已提示 */ }
    },
  })
}

const statusText = (s: number) => {
  const map: Record<number, string> = {
    0: $t('project.draft'), 1: $t('project.published'), 2: $t('project.assigned'),
    3: $t('project.inProgress'), 4: $t('project.completed'), 5: $t('project.offline'),
    6: $t('project.withdrawn'), 7: $t('project.abolished'),
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
  margin-bottom: 12rpx;
}
.p-title {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--text-color, #333);
  flex: 1;
}
.p-status {
  font-size: 24rpx;
  color: var(--muted-color, #666);
}
.p-info {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  font-size: 24rpx;
  color: var(--muted-color, #666);
}
.p-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 16rpx;
}
.btn {
  flex: 1;
  font-size: 24rpx;
  border-radius: 10rpx;
  line-height: 56rpx;
  height: 56rpx;
  padding: 0;
  margin: 0;
}
.btn-change {
  background: #eef4ff;
  color: #2563eb;
}
.btn-warn {
  background: #fdf6ec;
  color: #e6a23c;
}
.btn-danger {
  background: #fef0f0;
  color: #f56c6c;
}
.empty {
  text-align: center;
  color: #999;
  padding: 60rpx 0;
  font-size: 26rpx;
}
</style>
