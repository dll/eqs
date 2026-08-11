<template>
  <view class="container">
    <view
      v-if="templates.length"
      class="tpl-list"
    >
      <view
        v-for="t in templates"
        :key="t.id"
        class="tpl-card"
        @tap="viewDetail(t)"
      >
        <view class="t-head">
          <text class="t-name">
            {{ t.name }}
          </text>
          <text class="t-type">
            {{ t.service_type || '-' }}
          </text>
        </view>
        <text class="t-desc">
          {{ t.description || '标准交付模板' }}
        </text>
        <text class="t-version">
          v{{ t.version }} · {{ t.status }}
        </text>
      </view>
    </view>
    <view
      v-else
      class="empty"
    >
      暂无可用模板
    </view>

    <!-- 模板详情 + 清单校验 -->
    <view
      v-if="detail"
      class="detail-mask"
      @tap="closeDetail"
    >
      <view
        class="detail-card"
        @tap.stop
      >
        <text class="d-title">
          {{ detail.name }}
        </text>
        <text class="d-desc">
          {{ detail.description || '' }}
        </text>
        <text class="d-label">
          必交清单
        </text>
        <view
          v-for="(item, i) in checklistItems"
          :key="i"
          class="check-item"
        >
          <text class="check-name">
            {{ item }}
          </text>
          <text
            class="check-toggle"
            :class="checked[item] ? 'on' : ''"
            @tap="toggle(item)"
          >
            {{ checked[item] ? '✓ 已备齐' : '○ 未备齐' }}
          </text>
        </view>
        <view class="d-actions">
          <button
            class="d-btn"
            @tap="closeDetail"
          >
            关闭
          </button>
          <button
            class="d-btn primary"
            @tap="validate"
          >
            校验清单
          </button>
        </view>
        <text
          v-if="validateMsg"
          class="d-result"
        >
          {{ validateMsg }}
        </text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { usePageTitle } from '@/utils/i18n'

usePageTitle('page.templates', { onShow })

const templates = ref<any[]>([])
const detail = ref<any>(null)
const checked = ref<Record<string, boolean>>({})
const validateMsg = ref('')

const checklistItems = computed(() => {
  if (!detail.value?.checklist) return []
  try {
    const list = typeof detail.value.checklist === 'string' ? JSON.parse(detail.value.checklist) : detail.value.checklist
    return Array.isArray(list) ? list : []
  } catch {
    return []
  }
})

const load = async () => {
  try {
    const res = await request.get('/api/v1/delivery-templates', { silent401: true }).catch(() => ({ templates: [] }))
    templates.value = res.templates || []
  } catch {
    templates.value = []
  }
}

onShow(load)

const viewDetail = async (t: any) => {
  try {
    const res = await request.get(`/api/v1/delivery-templates/${t.id}`)
    detail.value = res.template
    checked.value = {}
    validateMsg.value = ''
  } catch {
    // request 已提示
  }
}

const closeDetail = () => {
  detail.value = null
}

const toggle = (item: string) => {
  checked.value[item] = !checked.value[item]
}

const validate = async () => {
  try {
    const res = await request.post(`/api/v1/delivery-templates/${detail.value.id}/validate`, {
      checklist_result: checked.value,
    })
    validateMsg.value = (res as any)?.message || '校验完成'
  } catch {
    // request 已提示
  }
}
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.tpl-card {
  background: var(--card-bg, #fff);
  border-radius: 14rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.t-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10rpx;
}
.t-name {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--text-color, #333);
  flex: 1;
}
.t-type {
  font-size: 22rpx;
  color: var(--primary-color);
  background: var(--input-bg, #f5f5f5);
  padding: 4rpx 14rpx;
  border-radius: 20rpx;
}
.t-desc {
  font-size: 24rpx;
  color: var(--muted-color, #666);
  display: block;
  margin-bottom: 8rpx;
}
.t-version {
  font-size: 22rpx;
  color: var(--muted-color, #999);
}
.empty {
  text-align: center;
  color: var(--muted-color, #999);
  padding: 120rpx 0;
  font-size: 26rpx;
}
.detail-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, .5);
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
}
.detail-card {
  background: var(--card-bg, #fff);
  border-radius: 16rpx;
  padding: 32rpx;
  width: 82%;
  max-height: 80vh;
  overflow-y: auto;
}
.d-title {
  font-size: 34rpx;
  font-weight: 700;
  color: var(--text-color, #333);
  display: block;
  margin-bottom: 8rpx;
}
.d-desc {
  font-size: 26rpx;
  color: var(--muted-color, #666);
  display: block;
  margin-bottom: 20rpx;
}
.d-label {
  font-size: 28rpx;
  font-weight: 600;
  color: var(--text-color, #333);
  margin-bottom: 12rpx;
  display: block;
}
.check-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12rpx 0;
  border-bottom: 1rpx solid var(--border-color, #eee);
}
.check-name {
  font-size: 26rpx;
  color: var(--text-color, #333);
  flex: 1;
}
.check-toggle {
  font-size: 24rpx;
  color: var(--muted-color, #999);
}
.check-toggle.on {
  color: #10b981;
}
.d-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 24rpx;
}
.d-btn {
  flex: 1;
  background: var(--input-bg, #f5f5f5);
  color: var(--text-color, #333);
  border-radius: 10rpx;
  font-size: 28rpx;
}
.d-btn.primary {
  background: var(--primary-color);
  color: #fff;
}
.d-result {
  display: block;
  text-align: center;
  margin-top: 16rpx;
  font-size: 26rpx;
  color: var(--primary-color);
}
</style>
