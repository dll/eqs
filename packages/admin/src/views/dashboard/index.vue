<template>
  <div class="dashboard">
    <!-- 统计卡 -->
    <el-row :gutter="16">
      <el-col v-for="c in statCards" :key="c.label" :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" :style="{ background: c.bg }">
            <el-icon :size="22"><component :is="c.icon" /></el-icon>
          </div>
          <div>
            <div class="stat-value">{{ c.value }}</div>
            <div class="stat-label">{{ c.label }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 甘特图：全部项目时间线 -->
    <el-card class="panel" shadow="hover">
      <template #header>
        <div class="panel-head">
          <span>{{ $t('dashboard.ganttAll') }}</span>
          <div class="panel-actions">
            <el-radio-group v-model="ganttMode" size="small" @change="loadProgress">
              <el-radio-button label="all">{{ $t('dashboard.allProjects') }}</el-radio-button>
              <el-radio-button label="single">{{ $t('dashboard.singleProject') }}</el-radio-button>
            </el-radio-group>
            <el-select v-if="ganttMode === 'single'" v-model="singleProjectId" size="small" style="width: 180px; margin-left: 8px" @change="loadProgress">
              <el-option v-for="p in progressProjects" :key="p.id" :label="`#${p.id} ${p.title}`" :value="p.id" />
            </el-select>
          </div>
        </div>
      </template>
      <div class="chart-area">
        <GanttChart v-if="ganttData.length" :projects="ganttData" @select="openProject" />
        <el-empty v-else :description="$t('dashboard.noProjects')" />
      </div>
    </el-card>

    <el-row :gutter="16" style="margin-top: 16px">
      <!-- 看板：进度分布 -->
      <el-col :span="12">
        <el-card class="panel" shadow="hover">
          <template #header>{{ $t('dashboard.kanbanProgress') }}</template>
          <div class="chart-area">
            <KanbanChart v-if="kanbanData.length" :items="kanbanData" />
            <el-empty v-else :description="$t('dashboard.noProjects')" />
          </div>
        </el-card>
      </el-col>
      <!-- AI 分析 -->
      <el-col :span="12">
        <el-card class="panel" shadow="hover">
          <template #header>
            <div class="panel-head">
              <span>🤖 {{ $t('dashboard.aiAnalysis') }}</span>
              <el-button size="small" type="primary" plain :loading="aiLoading" @click="runAI">
                {{ $t('dashboard.aiRun') }}
              </el-button>
            </div>
          </template>
          <div v-if="aiLoading" class="ai-loading">
            <el-skeleton :rows="4" animated />
          </div>
          <div v-else-if="aiResult" class="ai-body">
            <div class="ai-summary">
              <el-tag :type="aiResult.riskTag" size="small">{{ aiResult.riskLabel }}</el-tag>
              <p class="ai-text">{{ aiResult.summary }}</p>
            </div>
            <div v-if="aiResult.issues.length" class="ai-section">
              <div class="ai-section-title">{{ $t('dashboard.aiIssues') }}</div>
              <ul class="ai-list">
                <li v-for="(it, i) in aiResult.issues" :key="i">{{ it }}</li>
              </ul>
            </div>
            <div v-if="aiResult.suggestions.length" class="ai-section">
              <div class="ai-section-title">{{ $t('dashboard.aiSuggestions') }}</div>
              <ul class="ai-list success">
                <li v-for="(it, i) in aiResult.suggestions" :key="i">{{ it }}</li>
              </ul>
            </div>
          </div>
          <el-empty v-else :description="$t('dashboard.aiEmpty')" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近项目 -->
    <el-card class="panel" shadow="hover" style="margin-top: 16px">
      <template #header>{{ $t('dashboard.recentProjects') }}</template>
      <el-table :data="recentProjects" style="width: 100%">
        <el-table-column prop="id" :label="$t('dashboard.columnId')" width="80" />
        <el-table-column prop="title" :label="$t('dashboard.columnTitle')" />
        <el-table-column prop="project_type" :label="$t('dashboard.columnType')" width="120" />
        <el-table-column prop="status" :label="$t('dashboard.columnStatus')" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="progress" :label="$t('dashboard.columnProgress')" width="160">
          <template #default="{ row }">
            <el-progress :percentage="row.progress || 0" :stroke-width="10" />
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { User, Files, Tickets, Wallet } from '@element-plus/icons-vue'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'
import GanttChart, { type GanttItem } from '@/components/GanttChart.vue'
import KanbanChart, { type KanbanItem } from '@/components/KanbanChart.vue'

const { $t } = useI18n()
const router = useRouter()

const stats = ref({ user_count: 0, project_count: 0, order_count: 0, dispute_count: 0, settled_amount: 0 })
const recentProjects = ref<any[]>([])
const progressProjects = ref<any[]>([])
const ganttData = ref<GanttItem[]>([])
const kanbanData = ref<KanbanItem[]>([])
const ganttMode = ref('all')
const singleProjectId = ref<number>()
const aiResult = ref<any>(null)
const aiLoading = ref(false)

const statCards = computed(() => [
  { label: $t('dashboard.totalUsers'), value: stats.value.user_count, icon: User, bg: '#409EFF' },
  { label: $t('dashboard.totalProjects'), value: stats.value.project_count, icon: Files, bg: '#67C23A' },
  { label: $t('dashboard.totalOrders'), value: stats.value.order_count, icon: Tickets, bg: '#E6A23C' },
  { label: $t('dashboard.totalSettled'), value: `¥${stats.value.settled_amount}`, icon: Wallet, bg: '#F56C6C' },
])

let timer: ReturnType<typeof setInterval> | null = null

const loadStats = async () => {
  try {
    stats.value = await api.get('/api/v1/admin/stats')
  } catch { /* interceptor */ }
}

const loadProgress = async () => {
  try {
    if (ganttMode.value === 'single' && singleProjectId.value) {
      const r = await api.get<{ project: GanttItem }>(`/api/v1/project/${singleProjectId.value}/progress`)
      ganttData.value = [r.project]
    } else {
      const r = await api.get<{ projects: any[] }>('/api/v1/admin/project-progress')
      progressProjects.value = r.projects || []
      ganttData.value = r.projects || []
      kanbanData.value = (r.projects || []).map((p) => ({ id: p.id, title: p.title, progress: p.progress, status: p.status }))
      // 最近项目合并进度
      const pmap = new Map((r.projects || []).map((p) => [p.id, p.progress]))
      recentProjects.value = (recentProjects.value || []).map((p) => ({ ...p, progress: pmap.get(p.id) ?? 0 }))
    }
  } catch { /* interceptor */ }
}

const loadRecent = async () => {
  try {
    const p = await api.get<{ projects: any[] }>('/api/v1/project/list')
    recentProjects.value = (p.projects || []).slice(0, 8)
  } catch { /* interceptor */ }
}

const runAI = async () => {
  aiLoading.value = true
  try {
    const r = await api.post('/api/v1/admin/ai/project-analysis')
    const items = r.items || []
    const issues: string[] = []
    const suggestions: string[] = []
    items.forEach((it: any) => {
      ;(it.analysis.issues || []).forEach((x: string) => issues.push(`#${it.project_id} ${it.title}：${x}`))
      ;(it.analysis.suggestions || []).forEach((x: string) => suggestions.push(`#${it.project_id} ${it.title}：${x}`))
    })
    const riskLevels = items.map((it: any) => it.analysis.risk_level)
    const risk = riskLevels.includes('high') ? 'high' : riskLevels.includes('medium') ? 'medium' : 'low'
    aiResult.value = {
      summary: r.summary || (items.length ? `已分析 ${items.length} 个项目` : '暂无项目'),
      issues: issues.slice(0, 6),
      suggestions: suggestions.slice(0, 6),
      riskTag: risk === 'high' ? 'danger' : risk === 'medium' ? 'warning' : 'success',
      riskLabel: risk === 'high' ? $t('dashboard.riskHigh') : risk === 'medium' ? $t('dashboard.riskMedium') : $t('dashboard.riskLow'),
    }
  } catch { aiResult.value = null } finally {
    aiLoading.value = false
  }
}

const openProject = (id: number) => {
  router.push(`/project?id=${id}`)
}

const loadAll = () => {
  loadStats()
  loadProgress()
  loadRecent()
}

onMounted(() => {
  loadAll()
  timer = setInterval(loadProgress, 30000) // 实时更新 30s
})
onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: $t('project.status.draft'), 1: $t('project.status.published'), 2: $t('project.status.assigned'), 3: $t('project.status.inProgress'), 4: $t('project.status.completed')
  }
  return map[status] || $t('project.status.unknown')
}
const statusType = (status: number) => {
  const map: Record<number, string> = { 1: 'success', 3: 'primary', 4: 'success' }
  return map[status] || 'info'
}
</script>

<style scoped>
.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 6px 0;
}
.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}
.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
}
.stat-label {
  font-size: 13px;
  color: #909399;
}
.panel {
  margin-top: 16px;
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.panel-actions {
  display: flex;
  align-items: center;
}
.chart-area {
  height: 360px;
}
.ai-body .ai-summary {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}
.ai-text {
  margin: 0;
  color: #303133;
  line-height: 1.6;
}
.ai-section {
  margin-top: 12px;
}
.ai-section-title {
  font-weight: 600;
  color: #606266;
  margin-bottom: 6px;
}
.ai-list {
  margin: 0;
  padding-left: 18px;
  color: #606266;
  font-size: 13px;
  line-height: 1.8;
}
.ai-list.success {
  color: #67C23A;
}
</style>
