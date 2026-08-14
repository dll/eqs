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

      <!-- 服务超市增强：经营数据 -->
      <view
        v-if="stats"
        class="info-card stats-card"
      >
        <text class="qual-title">
          {{ $t('provider.business') }}
        </text>
        <view class="stats-row">
          <view class="stat-item">
            <text class="stat-num">
              {{ stats.orders_signed }}
            </text>
            <text class="stat-label">
              {{ $t('provider.ordersSigned') }}
            </text>
          </view>
          <view class="stat-item">
            <text class="stat-num">
              {{ stats.orders_completed }}
            </text>
            <text class="stat-label">
              {{ $t('provider.ordersCompleted') }}
            </text>
          </view>
          <view class="stat-item">
            <text class="stat-num">
              {{ stats.review_count }}
            </text>
            <text class="stat-label">
              {{ $t('provider.reviewCount') }}
            </text>
          </view>
        </view>

        <!-- 评分分布 -->
        <view
          v-if="stats.rating_dist"
          class="rating-dist"
        >
          <view
            v-for="star in [5, 4, 3, 2, 1]"
            :key="star"
            class="rating-row"
          >
            <text class="rating-label">
              {{ star }}★
            </text>
            <view class="rating-bar">
              <view
                class="rating-fill"
                :style="{ width: barWidth(star) + '%' }"
              />
            </view>
            <text class="rating-count">
              {{ stats.rating_dist[star] || 0 }}
            </text>
          </view>
        </view>
      </view>

      <!-- 近期评价 -->
      <view
        v-if="stats && stats.recent_reviews && stats.recent_reviews.length"
        class="info-card rev-card"
      >
        <text class="qual-title">
          {{ $t('provider.recentReviews') }}
        </text>
        <view
          v-for="(r, i) in stats.recent_reviews"
          :key="i"
          class="rev-item"
        >
          <text class="rev-star">
            {{ '★'.repeat(r.rating) }}{{ '☆'.repeat(5 - r.rating) }}
          </text>
          <text class="rev-content">
            {{ r.content || $t('provider.noComment') }}
          </text>
        </view>
      </view>

      <!-- V9 企业案例沉淀：服务案例 -->
      <view
        v-if="cases.length"
        class="info-card case-card"
      >
        <text class="qual-title">
          {{ $t('provider.cases') }}
        </text>
        <view
          v-for="cs in cases"
          :key="cs.id"
          class="case-item"
        >
          <text class="case-title">
            {{ cs.title }}
          </text>
          <text
            v-if="cs.description"
            class="case-desc"
          >
            {{ cs.description }}
          </text>
          <view
            v-if="cs.image_urls && cs.image_urls.length"
            class="case-imgs"
          >
            <image
              v-for="url in cs.image_urls"
              :key="url"
              class="case-img"
              :src="url"
              mode="aspectFill"
              @tap="previewCase(cs.image_urls)"
            />
          </view>
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
const stats = ref<any>(null)
const cases = ref<any[]>([])

onLoad((options) => {
  if (options?.id) {
    loadProvider(options.id)
    loadCases(options.id)
  }
})

// V9：服务案例（公开接口，image_urls 为签名公开预览链接）
const loadCases = async (id: string) => {
  try {
    const res = await request.get(`/api/v1/provider/${id}/cases`, { silent401: true })
    cases.value = (res && res.cases) || []
  } catch {
    cases.value = []
  }
}

// 预览案例成果图（签名公开链接，无需登录）
const previewCase = (urls: string[]) => {
  if (!urls || !urls.length) return
  uni.previewImage({ urls })
}

const loadProvider = async (id: string) => {
  try {
    const res = await request.get(`/api/v1/provider/${id}`, { silent401: true })
    provider.value = res.provider || null
    stats.value = res.stats || null
  } catch {
    provider.value = null
  }
}

const barWidth = (star: number) => {
  const total = stats.value?.review_count || 0
  if (!total) return 0
  return Math.round(((stats.value.rating_dist?.[star] || 0) / total) * 100)
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

.qual-card, .stats-card, .rev-card {
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

.stats-row {
  display: flex;
  justify-content: space-around;
  padding: 10rpx 0 20rpx;
}
.stat-item {
  text-align: center;
}
.stat-num {
  font-size: 34rpx;
  font-weight: 700;
  color: var(--primary-color);
  display: block;
}
.stat-label {
  font-size: 22rpx;
  color: var(--muted-color);
  margin-top: 6rpx;
}
.rating-dist {
  padding-top: 10rpx;
}
.rating-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 6rpx 0;
}
.rating-label {
  width: 50rpx;
  font-size: 24rpx;
  color: var(--muted-color);
}
.rating-bar {
  flex: 1;
  height: 14rpx;
  background: var(--input-bg);
  border-radius: 8rpx;
  overflow: hidden;
}
.rating-fill {
  height: 100%;
  background: linear-gradient(90deg, #f59e0b, #fbbf24);
  border-radius: 8rpx;
}
.rating-count {
  width: 40rpx;
  text-align: right;
  font-size: 22rpx;
  color: var(--muted-color);
}
.rev-item {
  padding: 12rpx 0;
  border-bottom: 1rpx solid var(--border-color);
}
.rev-star {
  color: #f59e0b;
  font-size: 26rpx;
  display: block;
  margin-bottom: 6rpx;
}
.rev-content {
  font-size: 26rpx;
  color: var(--text-color);
}
.case-card {
  margin-top: 20rpx;
}
.case-item {
  padding: 16rpx 0;
  border-bottom: 1rpx solid var(--border-color);
}
.case-title {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--text-color);
  display: block;
}
.case-desc {
  font-size: 26rpx;
  color: var(--muted-color);
  display: block;
  margin-top: 8rpx;
}
.case-imgs {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 12rpx;
}
.case-img {
  width: 160rpx;
  height: 120rpx;
  border-radius: 8rpx;
  background: var(--input-bg);
}
</style>