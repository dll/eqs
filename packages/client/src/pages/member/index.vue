<template>
  <view class="container">
    <!-- 当前会员状态 -->
    <view class="member-card">
      <text class="member-title">
        {{ $t('member.current') }}
      </text>
      <text class="member-level">
        {{ myInfo.level_name || $t('member.free') }}
      </text>
      <text
        v-if="myInfo.active"
        class="member-expire"
      >
        {{ $t('member.expireAt', { time: fmtDate(myInfo.expire_at) }) }}
      </text>
      <text
        v-else
        class="member-expire"
      >
        {{ $t('member.notActive') }}
      </text>
    </view>

    <!-- 等级权益 -->
    <view class="levels">
      <view
        v-for="lv in levels"
        :key="lv.level"
        class="level-card"
      >
        <text class="level-name">
          {{ lv.name }}
        </text>
        <text class="level-price">
          {{ lv.price_per_month > 0 ? `¥${lv.price_per_month}/${$t('member.month')}` : $t('member.free') }}
        </text>
        <view class="benefits">
          <text
            v-for="(b, i) in lv.benefits"
            :key="i"
            class="benefit"
          >
            ✓ {{ b }}
          </text>
        </view>
        <view
          v-if="lv.level !== 'free'"
          class="buy-row"
        >
          <button
            class="buy-btn"
            :disabled="buying"
            @tap="upgrade(lv.level, 1)"
          >
            {{ $t('member.buyMonth1') }}
          </button>
          <button
            class="buy-btn primary"
            :disabled="buying"
            @tap="upgrade(lv.level, 12)"
          >
            {{ $t('member.buyMonth12') }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.member', { onShow })

const levels = ref<any[]>([])
const myInfo = ref<any>({})
const buying = ref(false)

onShow(() => {
  loadLevels()
  loadInfo()
})

const loadLevels = async () => {
  try {
    const res = await request.get('/api/v1/member/levels', { silent401: true })
    levels.value = (res && res.levels) || []
  } catch {
    levels.value = []
  }
}

const loadInfo = async () => {
  try {
    const res = await request.get('/api/v1/member/info', { silent401: true })
    myInfo.value = res || {}
  } catch {
    myInfo.value = {}
  }
}

const upgrade = async (level: string, months: number) => {
  buying.value = true
  try {
    await request.post('/api/v1/member/upgrade', { level, months })
    uni.showToast({ title: $t('member.upgraded'), icon: 'success' })
    loadInfo()
  } catch {
    // request 已提示
  } finally {
    buying.value = false
  }
}

const fmtDate = (s: string) => (s ? s.replace('T', ' ').slice(0, 10) : '')
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.member-card {
  background: linear-gradient(135deg, #8b5cf6, #2563eb);
  border-radius: 16rpx;
  padding: 40rpx 30rpx;
  color: #fff;
  margin-bottom: 24rpx;
}

.member-title {
  font-size: 24rpx;
  opacity: 0.85;
  display: block;
}

.member-level {
  font-size: 44rpx;
  font-weight: bold;
  display: block;
  margin: 12rpx 0;
}

.member-expire {
  font-size: 26rpx;
  opacity: 0.9;
}

.level-card {
  background: var(--card-bg);
  border-radius: 14rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
  border: 1rpx solid var(--border-color);
}

.level-name {
  font-size: 34rpx;
  font-weight: bold;
  color: var(--text-color);
  display: block;
}

.level-price {
  font-size: 28rpx;
  color: var(--primary-color);
  margin-top: 8rpx;
  display: block;
}

.benefits {
  margin: 16rpx 0;
}

.benefit {
  display: block;
  font-size: 26rpx;
  color: var(--muted-color);
  padding: 6rpx 0;
}

.buy-row {
  display: flex;
  gap: 16rpx;
  margin-top: 10rpx;
}

.buy-btn {
  flex: 1;
  font-size: 26rpx;
  border-radius: 10rpx;
  background: #f3f4f6;
  color: var(--text-color);
  line-height: 64rpx;
  height: 64rpx;
  padding: 0;
}

.buy-btn.primary {
  background: var(--primary-color);
  color: #fff;
}
</style>
