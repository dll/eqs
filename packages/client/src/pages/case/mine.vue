<template>
  <view class="container">
    <!-- 我的案例 -->
    <view class="section">
      <text class="section-title">
        {{ $t('case.myCases') }}
      </text>
      <view
        v-if="cases.length === 0"
        class="empty"
      >
        <text>{{ $t('case.empty') }}</text>
      </view>
      <view
        v-for="cs in cases"
        :key="cs.id"
        class="case-item"
      >
        <view class="case-info">
          <text class="case-title">
            {{ cs.title }}
          </text>
          <text
            v-if="cs.description"
            class="case-desc"
          >
            {{ cs.description }}
          </text>
        </view>
        <view class="case-actions">
          <text
            class="del-btn"
            @tap="removeCase(cs)"
          >
            {{ $t('case.delete') }}
          </text>
        </view>
      </view>
    </view>

    <!-- 从已完成订单沉淀案例 -->
    <view class="section">
      <text class="section-title">
        {{ $t('case.fromCompleted') }}
      </text>
      <view
        v-if="completedOrders.length === 0"
        class="empty"
      >
        <text>{{ $t('case.noCompleted') }}</text>
      </view>
      <view
        v-for="o in completedOrders"
        :key="o.id"
        class="case-item"
      >
        <view class="case-info">
          <text class="case-title">
            {{ o.project?.title || o.id }}
          </text>
          <text class="case-desc">
            {{ $t('order.amountPrefix', { amount: o.amount }) }}
          </text>
        </view>
        <view class="case-actions">
          <text
            class="create-btn"
            @tap="createFromOrder(o)"
          >
            {{ $t('case.publish') }}
          </text>
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
usePageTitle('page.caseMine', { onShow })

const cases = ref<any[]>([])
const completedOrders = ref<any[]>([])

onShow(() => {
  loadCases()
  loadCompletedOrders()
})

const loadCases = async () => {
  try {
    const res = await request.get('/api/v1/case/mine')
    cases.value = (res && res.cases) || []
  } catch {
    cases.value = []
  }
}

// 已完成订单（status=3）且尚未沉淀为案例
const loadCompletedOrders = async () => {
  try {
    const res = await request.get('/api/v1/order/list')
    const orders: any[] = (res && res.orders) || []
    const cased = new Set(cases.value.map(c => c.order_id))
    completedOrders.value = orders.filter(o => o.status === 3 && !cased.has(o.id))
  } catch {
    completedOrders.value = []
  }
}

const createFromOrder = async (o: any) => {
  try {
    await request.post('/api/v1/case/create', {
      order_id: o.id,
      title: o.project?.title || `订单 #${o.id} 成果案例`,
      description: '',
    })
    uni.showToast({ title: $t('case.published'), icon: 'success' })
    loadCases()
    loadCompletedOrders()
  } catch {
    // request 已提示
  }
}

const removeCase = (cs: any) => {
  uni.showModal({
    title: $t('case.deleteConfirmTitle'),
    content: $t('case.deleteConfirm'),
    success: async (res: any) => {
      if (!res.confirm) return
      try {
        await request.delete(`/api/v1/case/${cs.id}`)
        uni.showToast({ title: $t('case.deleted'), icon: 'success' })
        loadCases()
        loadCompletedOrders()
      } catch {
        // request 已提示
      }
    },
  })
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

.empty {
  padding: 30rpx 0;
  text-align: center;
  color: var(--muted-color);
  font-size: 26rpx;
}

.case-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16rpx 0;
  border-bottom: 1rpx solid var(--border-color);
}

.case-info {
  flex: 1;
  margin-right: 20rpx;
}

.case-title {
  font-size: 28rpx;
  color: var(--text-color);
  display: block;
}

.case-desc {
  font-size: 24rpx;
  color: var(--muted-color);
  margin-top: 6rpx;
  display: block;
}

.case-actions {
  flex-shrink: 0;
}

.del-btn {
  color: #ef4444;
  font-size: 26rpx;
  padding: 10rpx 16rpx;
}

.create-btn {
  color: var(--primary-color);
  font-size: 26rpx;
  padding: 10rpx 16rpx;
  border: 1rpx solid var(--primary-color);
  border-radius: 8rpx;
}
</style>
