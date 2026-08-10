<template>
  <view class="container">
    <view
      v-if="provider"
      class="provider-info"
    >
      <view class="info-card">
        <text class="provider-name">
          {{ provider.company_name }}
        </text>
        <text class="provider-score">
          {{ $t('provider.creditScorePrefix', { score: provider.credit_score }) }}
        </text>
      </view>

      <view
        v-if="provider.qualifications && provider.qualifications.length"
        class="info-card qual-card"
      >
        <text class="qual-title">
          {{ $t('provider.qualifications') }}
        </text>
        <view
          v-for="(q, i) in provider.qualifications"
          :key="i"
          class="qual-item"
        >
          <text class="qual-type">
            {{ q.qualification_type }}
          </text>
          <text class="qual-level">
            {{ q.level }}
          </text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.providerDetail', { onLoad })
const provider = ref<any>(null)

onLoad((options) => {
  if (options?.id) {
    loadProvider(options.id)
  }
})

const loadProvider = async (id: string) => {
  try {
    const res = await request.get(`/api/v1/provider/${id}`, { silent401: true })
    provider.value = res.provider || null
  } catch {
    provider.value = null
  }
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.info-card {
  background: var(--card-bg);
  border-radius: 10rpx;
  padding: 30rpx;
}

.provider-name {
  font-size: 36rpx;
  font-weight: bold;
  color: var(--text-color);
  display: block;
  margin-bottom: 15rpx;
}

.provider-score {
  font-size: 28rpx;
  color: var(--primary-color);
}

.qual-card {
  margin-top: 20rpx;
}

.qual-title {
  font-size: 30rpx;
  font-weight: bold;
  color: var(--text-color);
  display: block;
  margin-bottom: 15rpx;
}

.qual-item {
  display: flex;
  justify-content: space-between;
  padding: 10rpx 0;
  font-size: 28rpx;
}

.qual-type {
  color: var(--text-color);
}

.qual-level {
  color: var(--primary-color);
}
</style>