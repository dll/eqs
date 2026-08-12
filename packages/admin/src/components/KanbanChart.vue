<template>
  <div class="kanban-wrap">
    <!-- 进度状态统计 -->
    <div class="kanban-summary">
      <div
        v-for="s in scheduleStats"
        :key="s.state"
        class="ks-item"
        :style="{ background: s.bg }"
      >
        <span class="ks-num">{{ s.count }}</span>
        <span class="ks-label">{{ s.text }}</span>
      </div>
    </div>

    <!-- 按项目状态分列 -->
    <div class="kanban-columns">
      <div
        v-for="col in columns"
        :key="col.status"
        class="kanban-col"
      >
        <div class="kb-col-head">
          <span class="kb-col-title">{{ col.title }}</span>
          <el-tag
            size="small"
            :type="col.tag"
            effect="plain"
          >
            {{ col.items.length }}
          </el-tag>
        </div>
        <div class="kb-col-body">
          <div
            v-for="item in col.items"
            :key="item.id"
            class="kb-card"
          >
            <div class="kb-card-title">
              <span class="kb-card-id">#{{ item.id }}</span>
              {{ item.title }}
            </div>
            <el-progress
              :percentage="item.progress"
              :stroke-width="8"
              :color="progressColor(item)"
            />
            <div class="kb-card-foot">
              <el-tag
                size="small"
                :type="scheduleType(item.schedule_state)"
                effect="dark"
              >
                {{ scheduleText(item.schedule_state) }}
              </el-tag>
              <span
                v-if="item.end_date"
                class="kb-card-date"
              >
                {{ item.start_date }} ~ {{ item.end_date }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/utils/i18n'

export interface KanbanItem {
  id: number
  title: string
  progress: number
  status: number
  start_date?: string | null
  end_date?: string | null
  schedule_state?: string
}

const props = defineProps<{ items: KanbanItem[] }>()
const { $t } = useI18n()

const COLUMNS: { status: number; key: string; tag: string }[] = [
  { status: 0, key: 'project.status.draft', tag: 'info' },
  { status: 1, key: 'project.status.published', tag: 'primary' },
  { status: 2, key: 'project.status.assigned', tag: 'warning' },
  { status: 3, key: 'project.status.inProgress', tag: 'danger' },
  { status: 4, key: 'project.status.completed', tag: 'success' },
  { status: 6, key: 'project.status.withdrawn', tag: 'info' },
  { status: 7, key: 'project.status.abolished', tag: 'info' },
]

const columns = computed(() =>
  COLUMNS.map((c) => ({
    ...c,
    title: $t(c.key),
    items: props.items.filter((p) => p.status === c.status),
  })),
)

const scheduleStats = computed(() => {
  const count = (state?: string) => props.items.filter((p) => (p.schedule_state || 'on_time') === state).length
  return [
    { state: 'ahead', text: $t('dashboard.scheduleAhead'), count: count('ahead'), bg: 'linear-gradient(135deg, #67C23A, #95d475)' },
    { state: 'on_time', text: $t('dashboard.scheduleOnTime'), count: count('on_time'), bg: 'linear-gradient(135deg, #2563eb, #60a5fa)' },
    { state: 'late', text: $t('dashboard.scheduleLate'), count: count('late'), bg: 'linear-gradient(135deg, #f56c6c, #fab6b6)' },
  ]
})

const scheduleType = (state?: string): 'success' | 'primary' | 'danger' =>
  state === 'ahead' ? 'success' : state === 'late' ? 'danger' : 'primary'

const scheduleText = (state?: string): string => {
  if (state === 'ahead') return $t('dashboard.scheduleAhead')
  if (state === 'late') return $t('dashboard.scheduleLate')
  return $t('dashboard.scheduleOnTime')
}

const progressColor = (item: KanbanItem): string => {
  if (item.schedule_state === 'ahead') return '#67C23A'
  if (item.schedule_state === 'late') return '#F56C6C'
  return '#409EFF'
}
</script>

<style scoped>
.kanban-wrap {
  width: 100%;
  height: 100%;
  min-height: 340px;
  overflow-x: auto;
}
.kanban-summary {
  display: flex;
  gap: 12px;
  margin-bottom: 14px;
}
.ks-item {
  flex: 1;
  border-radius: 10px;
  padding: 10px 14px;
  color: #fff;
  display: flex;
  align-items: baseline;
  gap: 8px;
  box-shadow: 0 4px 12px rgba(37, 99, 235, .18);
}
.ks-num {
  font-size: 22px;
  font-weight: 700;
}
.ks-label {
  font-size: 12px;
  opacity: .95;
}
.kanban-columns {
  display: flex;
  gap: 12px;
}
.kanban-col {
  flex: 1;
  min-width: 150px;
  background: var(--eqs-gradient-soft, #f5f7fb);
  border-radius: 10px;
  padding: 10px;
  display: flex;
  flex-direction: column;
}
.kb-col-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
  padding: 0 4px;
}
.kb-col-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--eqs-text, #303133);
}
.kb-col-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 40px;
}
.kb-card {
  background: #fff;
  border-radius: 8px;
  padding: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, .06);
  transition: transform .15s;
}
.kb-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(37, 99, 235, .14);
}
.kb-card-title {
  font-size: 13px;
  color: var(--eqs-text, #303133);
  margin-bottom: 8px;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.kb-card-id {
  color: #909399;
  font-size: 11px;
  margin-right: 4px;
}
.kb-card-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8px;
  gap: 6px;
  flex-wrap: wrap;
}
.kb-card-date {
  font-size: 11px;
  color: #909399;
}
</style>