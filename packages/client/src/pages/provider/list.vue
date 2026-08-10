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

    <view class="provider-list">
      <view
        v-for="p in providers"
        :key="p.id"
        class="provider-card"
        @tap="goToDetail(p.id)"
      >
        <view class="provider-info">
          <text class="provider-name">
            {{ p.company_name }}
          </text>
          <text class="provider-score">
            {{ $t('provider.creditScorePrefix', { score: p.credit_score }) }}
          </text>
        </view>
      </view>
    </view>

    <view
      v-if="providers.length === 0"
      class="empty"
    >
      <text>{{ $t('common.empty') }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.providerList', { onShow })

const activeType = ref('')
interface ProviderInfo {
  id: number
  company_name: string
  credit_score: number
}
const providers = ref<ProviderInfo[]>([])

onShow(() => {
  loadProviders()
})

const loadProviders = async () => {
  try {
    const type = activeType.value
    const params = type ? `?type=${type}` : ''
    const res = await request.get(`/api/v1/provider/list${params}`, { silent401: true }).catch(() => ({ providers: [] }))
    providers.value = res.providers || []
  } catch {
    // request 已提示
  }
}

const filterByType = (type: string) => {
  activeType.value = type
  loadProviders()
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
}

.filter-item.active {
  color: var(--primary-color);
  font-weight: bold;
}

.provider-card {
  background: var(--card-bg);
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.provider-name {
  font-size: 30rpx;
  font-weight: bold;
  color: var(--text-color);
  display: block;
  margin-bottom: 10rpx;
}

.provider-score {
  font-size: 26rpx;
  color: var(--primary-color);
}

.empty {
  text-align: center;
  padding: 100rpx 0;
  color: var(--muted-color);
}
</style>