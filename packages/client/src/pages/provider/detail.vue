<template>
  <view class="container">
    <view class="provider-info" v-if="provider">
      <view class="info-card">
        <text class="provider-name">{{ provider.company_name }}</text>
        <text class="provider-score">{{ $t('provider.creditScorePrefix', { score: provider.credit_score }) }}</text>
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
    provider.value = { company_name: $t('provider.sample'), credit_score: 95 }
  }
})
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
</style>