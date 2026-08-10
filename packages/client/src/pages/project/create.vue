<template>
  <view class="container">
    <view class="form">
      <view class="form-item">
        <text class="label">
          {{ $t('project.type') }}
        </text>
        <picker
          :range="projectTypes"
          @change="onTypeChange"
        >
          <view class="picker">
            {{ form.projectType || $t('project.selectType') }}
          </view>
        </picker>
      </view>

      <view class="form-item">
        <text class="label">
          {{ $t('project.name') }}
        </text>
        <input
          v-model="form.title"
          class="input"
          :placeholder="$t('project.namePlaceholder')"
        >
      </view>

      <view class="form-item">
        <text class="label">
          {{ $t('project.budgetRange') }}
        </text>
        <view class="budget-row">
          <input
            v-model="form.budgetMin"
            class="input budget-input"
            :placeholder="$t('project.budgetMin')"
            type="digit"
          >
          <text class="budget-sep">
            -
          </text>
          <input
            v-model="form.budgetMax"
            class="input budget-input"
            :placeholder="$t('project.budgetMax')"
            type="digit"
          >
        </view>
      </view>

      <view class="form-item">
        <text class="label">
          {{ $t('project.location') }}
        </text>
        <input
          v-model="form.address"
          class="input"
          :placeholder="$t('project.addressPlaceholder')"
        >
      </view>

      <view class="form-item">
        <text class="label">
          {{ $t('project.desc') }}
        </text>
        <textarea
          v-model="form.description"
          class="textarea"
          :placeholder="$t('project.descPlaceholder')"
        />
      </view>

      <button
        class="submit-btn"
        @tap="submit"
      >
        {{ $t('project.create') }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'
import { PROJECT_TYPES, toTypeCode } from '@/lib/service'

const { $t } = useI18n()
usePageTitle('page.projectCreate', { onLoad })

const projectTypes = PROJECT_TYPES

const form = ref({
  projectType: '',
  title: '',
  budgetMin: '',
  budgetMax: '',
  address: '',
  description: '',
})

const onTypeChange = (e: any) => {
  form.value.projectType = projectTypes[e.detail.value]
}

const submit = async () => {
  if (!form.value.projectType || !form.value.title) {
    uni.showToast({ title: $t('project.required'), icon: 'none' })
    return
  }
  try {
    await request.post('/api/v1/project/create', {
      project_type: toTypeCode(form.value.projectType),
      service_type: toTypeCode(form.value.projectType),
      title: form.value.title,
      address: form.value.address,
      budget_min: Number(form.value.budgetMin) || 0,
      budget_max: Number(form.value.budgetMax) || 0,
      description: form.value.description,
      publish_scope: 'public',
    })
    uni.showToast({ title: $t('project.publishSuccess'), icon: 'success' })
    setTimeout(() => uni.navigateBack(), 1500)
  } catch {
    // request 已提示
  }
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.form-item {
  margin-bottom: 30rpx;
}

.label {
  display: block;
  font-size: 28rpx;
  color: var(--text-color);
  margin-bottom: 12rpx;
}

.picker {
  background: var(--input-bg);
  padding: 24rpx 30rpx;
  border-radius: 10rpx;
  font-size: 30rpx;
  color: var(--text-color);
}

.input {
  background: var(--input-bg);
  padding: 24rpx 30rpx;
  border-radius: 10rpx;
  font-size: 30rpx;
}

.budget-row {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.budget-input {
  flex: 1;
}

.budget-sep {
  color: var(--muted-color);
}

.textarea {
  background: var(--input-bg);
  padding: 24rpx 30rpx;
  border-radius: 10rpx;
  font-size: 30rpx;
  width: auto;
  min-height: 160rpx;
}

.submit-btn {
  background: var(--primary-color);
  color: #fff;
  margin-top: 40rpx;
  border-radius: 10rpx;
}
</style>