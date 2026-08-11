<template>
  <view class="container">
    <view class="card">
      <view class="card-title">{{ $t('qual.title') }}</view>
      <view class="form-item">
        <text class="label">{{ $t('qual.type') }}</text>
        <input v-model="form.qualification_type" class="input" :placeholder="$t('qual.typePh')" />
      </view>
      <view class="form-item">
        <text class="label">{{ $t('qual.certNo') }}</text>
        <input v-model="form.certificate_no" class="input" :placeholder="$t('qual.certNoPh')" />
      </view>
      <view class="form-item">
        <text class="label">{{ $t('qual.level') }}</text>
        <input v-model="form.level" class="input" :placeholder="$t('qual.levelPh')" />
      </view>
      <view class="form-item">
        <text class="label">{{ $t('qual.scope') }}</text>
        <input v-model="form.scope" class="input" :placeholder="$t('qual.scopePh')" />
      </view>
      <button
        class="submit-btn"
        @tap="submit"
      >
        {{ $t('qual.submit') }}
      </button>
      <text
        v-if="submitMsg"
        class="submit-msg"
      >
        {{ submitMsg }}
      </text>
    </view>

    <view class="card">
      <view class="card-title">{{ $t('qual.myList') }}</view>
      <view
        v-if="qualList.length"
        class="q-item"
      >
        <view
          v-for="q in qualList"
          :key="q.id"
          class="q-row"
        >
          <text class="q-type">
            {{ q.qualification_type }}
          </text>
          <text class="q-status" :class="'s-' + q.verification_status">
            {{ statusText(q.verification_status) }}
          </text>
        </view>
      </view>
      <text
        v-else
        class="empty"
      >
        {{ $t('qual.empty') }}
      </text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useUserStore } from '@/store/user'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.qual', { onShow })
const userStore = useUserStore()

const form = ref({ qualification_type: '', certificate_no: '', level: '', scope: '' })
const qualList = ref<any[]>([])
const submitMsg = ref('')

const supplierId = () => userStore.user?.id

onShow(() => loadList())

const loadList = async () => {
  const sid = supplierId()
  if (!sid) return
  try {
    const res = await request.get(`/api/v1/supplier/${sid}/qualifications`, { silent401: true }).catch(() => ({ qualifications: [] }))
    qualList.value = res.qualifications || []
  } catch {
    qualList.value = []
  }
}

const submit = async () => {
  const sid = supplierId()
  if (!sid) {
    submitMsg.value = $t('qual.notSupplier')
    return
  }
  if (!form.value.qualification_type || !form.value.certificate_no) {
    submitMsg.value = $t('qual.required')
    return
  }
  try {
    await request.post(`/api/v1/supplier/${sid}/qualifications`, form.value)
    submitMsg.value = $t('qual.submitted')
    form.value = { qualification_type: '', certificate_no: '', level: '', scope: '' }
    loadList()
  } catch {
    submitMsg.value = $t('qual.fail')
  }
}

const statusText = (s: string) => {
  const map: Record<string, string> = {
    pending: $t('qual.pending'),
    approved: $t('qual.approved'),
    rejected: $t('qual.rejected'),
    expired: $t('qual.expired'),
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
.card {
  background: var(--card-bg, #fff);
  border-radius: 10rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.card-title {
  font-size: 32rpx;
  font-weight: bold;
  color: var(--text-color, #333);
  margin-bottom: 20rpx;
}
.form-item {
  margin-bottom: 20rpx;
}
.label {
  display: block;
  font-size: 26rpx;
  color: var(--muted-color, #666);
  margin-bottom: 8rpx;
}
.input {
  background: var(--input-bg, #f5f5f5);
  border-radius: 8rpx;
  padding: 16rpx 20rpx;
  font-size: 28rpx;
  color: var(--text-color, #333);
}
.submit-btn {
  background: var(--primary-color, #409eff);
  color: #fff;
  border-radius: 10rpx;
  font-size: 30rpx;
  margin-top: 10rpx;
}
.submit-msg {
  display: block;
  text-align: center;
  margin-top: 16rpx;
  font-size: 26rpx;
  color: var(--primary-color);
}
.q-item .q-row {
  display: flex;
  justify-content: space-between;
  padding: 14rpx 0;
  border-bottom: 1rpx solid var(--border-color, #eee);
  font-size: 28rpx;
}
.q-type {
  color: var(--text-color, #333);
}
.q-status {
  font-size: 24rpx;
  padding: 4rpx 14rpx;
  border-radius: 20rpx;
  color: #fff;
}
.s-pending { background: #e6a23c; }
.s-approved { background: #67c23a; }
.s-rejected { background: #f56c6c; }
.s-expired { background: #909399; }
.empty {
  font-size: 26rpx;
  color: var(--muted-color, #999);
}
</style>
