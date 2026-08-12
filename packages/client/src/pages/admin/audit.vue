<template>
  <view class="container">
    <view
      v-for="q in list"
      :key="q.id"
      class="audit-card"
    >
      <view class="a-head">
        <text class="a-title">
          #{{ q.id }} {{ q.qualification_type }}
        </text>
        <text
          class="a-status"
          :class="'st-' + q.verification_status"
        >
          {{ statusText(q.verification_status) }}
        </text>
      </view>
      <view class="a-info">
        <text>{{ $t('admin.supplierId') }}：{{ q.supplier_id }}</text>
        <text>{{ $t('admin.certNo') }}：{{ q.certificate_no }}</text>
        <text>{{ $t('admin.level') }}：{{ q.level }}｜{{ $t('admin.authority') }}：{{ q.issuing_authority || '-' }}</text>
        <text>{{ $t('admin.validTo') }}：{{ q.valid_to ? String(q.valid_to).slice(0, 10) : '-' }}</text>
        <text
          v-if="q.review_comment"
          class="a-comment"
        >
          {{ q.review_comment }}
        </text>
      </view>
      <view
        v-if="q.verification_status === 'pending'"
        class="a-actions"
      >
        <button
          class="btn btn-pass"
          @tap="review(q, true)"
        >
          {{ $t('common.approve') }}
        </button>
        <button
          class="btn btn-reject"
          @tap="review(q, false)"
        >
          {{ $t('common.reject') }}
        </button>
      </view>
    </view>

    <view
      v-if="!list.length"
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

const userStore = useUserStore()
const { $t } = useI18n()
usePageTitle('page.adminAudit', { onShow })

const list = ref<any[]>([])

const load = async () => {
  try {
    const res = await request.get('/api/v1/admin/qualifications', { silent401: true })
    list.value = res.qualifications || []
  } catch {
    list.value = []
  }
}

onShow(() => {
  if (userStore.user?.user_type !== 3) {
    uni.reLaunch({ url: '/pages/index/index' })
    return
  }
  load()
})

const review = async (q: any, verified: boolean) => {
  try {
    await request.post(`/api/v1/qualification/${q.id}/review`, { verified, comment: '' })
    uni.showToast({ title: verified ? $t('common.approve') : $t('common.reject'), icon: 'success' })
    load()
  } catch { /* toast 已提示 */ }
}

const statusText = (status: string) => {
  const map: Record<string, string> = {
    pending: $t('audit.status.pending'),
    approved: $t('audit.status.approved'),
    rejected: $t('audit.status.rejected'),
    expired: $t('audit.status.expired'),
  }
  return map[status] || status
}
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.audit-card {
  background: var(--card-bg, #fff);
  border-radius: 14rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.a-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}
.a-title {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--text-color, #333);
}
.a-status {
  font-size: 24rpx;
}
.st-pending {
  color: #e6a23c;
}
.st-approved {
  color: #67c23a;
}
.st-rejected {
  color: #f56c6c;
}
.a-info {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  font-size: 24rpx;
  color: var(--muted-color, #666);
}
.a-comment {
  color: #e6a23c;
}
.a-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 16rpx;
}
.btn {
  flex: 1;
  font-size: 26rpx;
  border-radius: 10rpx;
  line-height: 60rpx;
  height: 60rpx;
  padding: 0;
  margin: 0;
}
.btn-pass {
  background: #f0f9eb;
  color: #67c23a;
}
.btn-reject {
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
