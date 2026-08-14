<template>
  <view class="container">
    <!-- 清单项 -->
    <view class="section">
      <text class="section-title">
        {{ $t('tool.estimateItems') }}
      </text>
      <view
        v-for="(it, i) in items"
        :key="i"
        class="item-row"
      >
        <input
          v-model="it.name"
          class="input item-name"
          :placeholder="$t('tool.itemName')"
        >
        <input
          v-model="it.unit"
          class="input item-unit"
          :placeholder="$t('tool.itemUnit')"
        >
        <input
          v-model="it.quantity"
          class="input item-num"
          :placeholder="$t('tool.itemQty')"
          type="digit"
        >
        <input
          v-model="it.unit_price"
          class="input item-num"
          :placeholder="$t('tool.itemPrice')"
          type="digit"
        >
        <text
          class="del-btn"
          @tap="items.splice(i, 1)"
        >
          ✕
        </text>
      </view>
      <button
        class="add-btn"
        @tap="items.push({ name: '', unit: '', quantity: '', unit_price: '' })"
      >
        {{ $t('tool.addItem') }}
      </button>
    </view>

    <!-- 费率（可调） -->
    <view class="section">
      <text class="section-title">
        {{ $t('tool.rates') }}
      </text>
      <view class="rate-row">
        <text class="rate-label">
          {{ $t('tool.measureRate') }}
        </text>
        <input
          v-model="rates.measure"
          class="input rate-input"
          type="digit"
        >
        <text class="rate-unit">
          %
        </text>
      </view>
      <view class="rate-row">
        <text class="rate-label">
          {{ $t('tool.overheadRate') }}
        </text>
        <input
          v-model="rates.overhead"
          class="input rate-input"
          type="digit"
        >
        <text class="rate-unit">
          %
        </text>
      </view>
      <view class="rate-row">
        <text class="rate-label">
          {{ $t('tool.taxRate') }}
        </text>
        <input
          v-model="rates.tax"
          class="input rate-input"
          type="digit"
        >
        <text class="rate-unit">
          %
        </text>
      </view>
    </view>

    <button
      class="calc-btn"
      @tap="calc"
    >
      {{ $t('tool.calculate') }}
    </button>

    <!-- 结果 -->
    <view
      v-if="result"
      class="section result-card"
    >
      <view class="result-row">
        <text class="result-label">
          {{ $t('tool.subtotal') }}
        </text>
        <text class="result-value">
          ¥{{ result.subtotal }}
        </text>
      </view>
      <view class="result-row">
        <text class="result-label">
          {{ $t('tool.measureFee') }}
        </text>
        <text class="result-value">
          ¥{{ result.measure_fee }}
        </text>
      </view>
      <view class="result-row">
        <text class="result-label">
          {{ $t('tool.overheadFee') }}
        </text>
        <text class="result-value">
          ¥{{ result.overhead_fee }}
        </text>
      </view>
      <view class="result-row">
        <text class="result-label">
          {{ $t('tool.tax') }}
        </text>
        <text class="result-value">
          ¥{{ result.tax }}
        </text>
      </view>
      <view class="result-row total">
        <text class="result-label">
          {{ $t('tool.total') }}
        </text>
        <text class="result-value">
          ¥{{ result.total }}
        </text>
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
usePageTitle('page.toolEstimate', { onLoad })

const items = ref<any[]>([{ name: '', unit: '', quantity: '', unit_price: '' }])
const rates = ref({ measure: '5', overhead: '2', tax: '9' })
const result = ref<any>(null)

const calc = async () => {
  const list = items.value
    .filter((it: any) => it.name && it.quantity !== '' && it.unit_price !== '')
    .map((it: any) => ({
      name: it.name,
      unit: it.unit || '-',
      quantity: Number(it.quantity) || 0,
      unit_price: Number(it.unit_price) || 0,
    }))
  if (list.length === 0) {
    uni.showToast({ title: $t('tool.needItem'), icon: 'none' })
    return
  }
  try {
    const res = await request.post('/api/v1/tools/cost-estimate', {
      items: list,
      rates: {
        measure_rate: Number(rates.value.measure) || 0,
        overhead_rate: Number(rates.value.overhead) || 0,
        tax_rate: Number(rates.value.tax) || 0,
      },
    })
    result.value = res.result || null
  } catch {
    // request 已提示
  }
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.section {
  background: var(--card-bg);
  border-radius: 10rpx;
  padding: 30rpx;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: bold;
  color: var(--text-color);
  display: block;
  margin-bottom: 16rpx;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 10rpx;
  margin-bottom: 12rpx;
}

.input {
  background: var(--input-bg);
  border-radius: 8rpx;
  padding: 16rpx 14rpx;
  font-size: 26rpx;
  color: var(--text-color);
}

.item-name {
  flex: 2;
  min-width: 0;
}

.item-unit {
  width: 90rpx;
}

.item-num {
  width: 130rpx;
}

.del-btn {
  color: #ef4444;
  font-size: 30rpx;
  padding: 10rpx;
}

.add-btn {
  margin-top: 10rpx;
  font-size: 26rpx;
  background: transparent;
  color: var(--primary-color);
  border: 1rpx dashed var(--primary-color);
  border-radius: 8rpx;
}

.rate-row {
  display: flex;
  align-items: center;
  margin-bottom: 12rpx;
}

.rate-label {
  flex: 1;
  font-size: 26rpx;
  color: var(--text-color);
}

.rate-input {
  width: 120rpx;
  text-align: center;
}

.rate-unit {
  width: 40rpx;
  text-align: center;
  color: var(--muted-color);
}

.calc-btn {
  background: var(--primary-color);
  color: #fff;
  border-radius: 10rpx;
  margin-bottom: 24rpx;
}

.result-card {
  border-left: 6rpx solid var(--primary-color);
}

.result-row {
  display: flex;
  justify-content: space-between;
  padding: 12rpx 0;
  font-size: 28rpx;
}

.result-label {
  color: var(--muted-color);
}

.result-value {
  color: var(--text-color);
  font-weight: 600;
}

.result-row.total {
  border-top: 1rpx solid var(--border-color);
  margin-top: 10rpx;
  padding-top: 20rpx;
}

.result-row.total .result-value {
  color: var(--primary-color);
  font-size: 34rpx;
}
</style>
