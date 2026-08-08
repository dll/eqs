<template>
  <view class="container">
    <view class="filter-bar">
      <view :class="['filter-item', activeType === '' ? 'active' : '']" @tap="filterByType('')">全部</view>
      <view :class="['filter-item', activeType === '造价' ? 'active' : '']" @tap="filterByType('造价')">造价</view>
      <view :class="['filter-item', activeType === '监理' ? 'active' : '']" @tap="filterByType('监理')">监理</view>
      <view :class="['filter-item', activeType === '地勘' ? 'active' : '']" @tap="filterByType('地勘')">地勘</view>
      <view :class="['filter-item', activeType === '设计' ? 'active' : '']" @tap="filterByType('设计')">设计</view>
    </view>

    <view class="provider-list">
      <view class="provider-card" v-for="p in providers" :key="p.id" @tap="goToDetail(p.id)">
        <view class="provider-info">
          <text class="provider-name">{{ p.company_name }}</text>
          <text class="provider-score">信用分：{{ p.credit_score }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const activeType = ref('')
interface ProviderInfo {
  id: number
  company_name: string
  credit_score: number
}
const providers = ref<ProviderInfo[]>([])

const filterByType = (type: string) => {
  activeType.value = type
}

const goToDetail = (id: number) => {
  uni.navigateTo({ url: `/pages/provider/detail?id=${id}` })
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.filter-bar {
  display: flex;
  background: #fff;
  padding: 20rpx;
  border-radius: 10rpx;
  margin-bottom: 20rpx;
}

.filter-item {
  flex: 1;
  text-align: center;
  font-size: 28rpx;
  color: #666;
}

.filter-item.active {
  color: #1890ff;
  font-weight: bold;
}

.provider-card {
  background: #fff;
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.provider-name {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 10rpx;
}

.provider-score {
  font-size: 26rpx;
  color: #1890ff;
}
</style>
