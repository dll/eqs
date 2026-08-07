<template>
  <view class="container">
    <view class="provider-info" v-if="provider">
      <view class="info-card">
        <text class="provider-name">{{ provider.company_name }}</text>
        <text class="provider-score">信用分：{{ provider.credit_score }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const provider = ref<any>(null)

onLoad((options) => {
  if (options?.id) {
    provider.value = { company_name: '示例服务商', credit_score: 95 }
  }
})
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.info-card {
  background: #fff;
  border-radius: 10rpx;
  padding: 30rpx;
}

.provider-name {
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 15rpx;
}

.provider-score {
  font-size: 28rpx;
  color: #1890ff;
}
</style>
