<template>
  <div class="dashboard">
    <!-- 统计卡 -->
    <el-row :gutter="16">
      <el-col
        v-for="c in statCards"
        :key="c.label"
        :span="6"
      >
        <el-card
          shadow="hover"
          class="stat-card"
        >
          <div
            class="stat-icon"
            :style="{ background: c.bg }"
          >
            <el-icon :size="22">
              <component :is="c.icon" />
            </el-icon>
          </div>
          <div>
            <div class="stat-value">
              {{ c.value }}
            </div>
            <div class="stat-label">
              {{ c.label }}
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 甘特图：全部项目时间线 -->
    <el-card
      class="panel"
      shadow="hover"
    >
      <template #header>
        <div class="panel-head">
          <span>{{ $t('dashboard.ganttAll') }}</span>
          <div class="panel-actions">
            <el-radio-group
              v-model="ganttMode"
              size="small"
              @change="loadProgress"
            >
              <el-radio-button label="all">
                {{ $t('dashboard.allProjects') }}
              </el-radio-button>
              <el-radio-button label="single">
                {{ $t('dashboard.singleProject') }}
              </el-radio-button>
            </el-radio-group>
            <el-select
              v-if="ganttMode === 'single'"
              v-model="singleProjectId"
              size="small"
              style="width: 180px; margin-left: 8px"
              @change="loadProgress"
            >
              <el-option
                v-for="p in progressProjects"
                :key="p.id"
                :label="`#${p.id} ${p.title}`"
                :value="p.id"
              />
            </el-select>
          </div>
        </div>
      </template>
      <div class="chart-area">
        <GanttChart
          v-if="ganttData.length"
          :projects="ganttData"
          @select="openProject"
        />
        <el-empty
          v-else
          :description="$t('dashboard.noProjects')"
        />
      </div>
    </el-card>

    <el-row
      :gutter="16"
      style="margin-top: 16px"
    >
      <!-- 看板：进度分布 -->
      <el-col :span="12">
        <el-card
          class="panel"
          shadow="hover"
        >
          <template #header>
            {{ $t('dashboard.kanbanProgress') }}
          </template>
          <div class="chart-area">
            <KanbanChart
              v-if="kanbanData.length"
              :items="kanbanData"
            />
            <el-empty
              v-else
              :description="$t('dashboard.noProjects')"
            />
          </div>
        </el-card>
      </el-col>
      <!-- AI 分析 -->
      <el-col :span="12">
        <el-card
          class="panel"
          shadow="hover"
        >
          <template #header>
            <div class="panel-head">
              <span>🤖 {{ $t('dashboard.aiAnalysis') }}</span>
              <el-button
                size="small"
                type="primary"
                plain
                :loading="aiLoading"
                @click="runAI"
              >
                {{ $t('dashboard.aiRun') }}
              </el-button>
            </div>
          </template>
          <div
            v-if="aiLoading"
            class="ai-loading"
          >
            <el-skeleton
              :rows="4"
              animated
            />
          </div>
          <div
            v-else-if="aiResult"
            class="ai-body"
          >
            <div class="ai-summary">
              <el-tag
                :type="aiResult.riskTag"
                size="small"
              >
                {{ aiResult.riskLabel }}
              </el-tag>
              <p class="ai-text">
                {{ aiResult.summary }}
              </p>
            </div>
            <div
              v-if="aiResult.issues.length"
              class="ai-section"
            >
              <div class="ai-section-title">
                {{ $t('dashboard.aiIssues') }}
              </div>
              <ul class="ai-list">
                <li
                  v-for="(it, i) in aiResult.issues"
                  :key="i"
                >
                  {{ it }}
                </li>
              </ul>
            </div>
            <div
              v-if="aiResult.suggestions.length"
              class="ai-section"
            >
              <div class="ai-section-title">
                {{ $t('dashboard.aiSuggestions') }}
              </div>
              <ul class="ai-list success">
                <li
                  v-for="(it, i) in aiResult.suggestions"
                  :key="i"
                >
                  {{ it }}
                </li>
              </ul>
            </div>
          </div>
          <el-empty
            v-else
            :description="$t('dashboard.aiEmpty')"
          />
        </el-card>
      </el-col>
    </el-row>

    <!-- 运营看板（V10） -->
    <el-row
      :gutter="16"
      style="margin-top: 16px"
    >
      <el-col :span="12">
        <el-card
          class="panel"
          shadow="hover"
        >
          <template #header>
            {{ $t('dashboard.opsTitle') }}
          </template>
          <div class="ops-users">
            <div class="ops-user-item">
              <span class="ops-num">{{ ops?.users?.clients ?? 0 }}</span>
              <span class="ops-label">{{ $t('dashboard.opsClients') }}</span>
            </div>
            <div class="ops-user-item">
              <span class="ops-num">{{ ops?.users?.suppliers ?? 0 }}</span>
              <span class="ops-label">{{ $t('dashboard.opsSuppliers') }}</span>
            </div>
            <div class="ops-user-item">
              <span class="ops-num">{{ ops?.users?.experts ?? 0 }}</span>
              <span class="ops-label">{{ $t('dashboard.opsExperts') }}</span>
            </div>
            <div class="ops-user-item">
              <span class="ops-num">{{ ops?.active_suppliers_7d ?? 0 }}</span>
              <span class="ops-label">{{ $t('dashboard.opsActive7d') }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card
          class="panel"
          shadow="hover"
        >
          <template #header>
            {{ $t('dashboard.opsFunnel') }}
          </template>
          <div class="ops-funnel">
            <div class="funnel-row">
              <span class="funnel-label">{{ $t('dashboard.funnelPublished') }}</span>
              <el-progress
                :percentage="100"
                :stroke-width="12"
                color="#2563eb"
              />
              <span class="funnel-num">{{ ops?.funnel?.published ?? 0 }}</span>
            </div>
            <div class="funnel-row">
              <span class="funnel-label">{{ $t('dashboard.funnelWithBid') }}</span>
              <el-progress
                :percentage="funnelPct(ops?.funnel?.with_bid)"
                :stroke-width="12"
                color="#06b6d4"
              />
              <span class="funnel-num">{{ ops?.funnel?.with_bid ?? 0 }}</span>
            </div>
            <div class="funnel-row">
              <span class="funnel-label">{{ $t('dashboard.funnelCompleted') }}</span>
              <el-progress
                :percentage="funnelPct(ops?.funnel?.completed)"
                :stroke-width="12"
                color="#8b5cf6"
              />
              <span class="funnel-num">{{ ops?.funnel?.completed ?? 0 }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近项目 -->
    <el-card
      class="panel"
      shadow="hover"
      style="margin-top: 16px"
    >
      <template #header>
        {{ $t('dashboard.recentProjects') }}
      </template>
      <el-table
        :data="recentProjects"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          :label="$t('dashboard.columnId')"
          width="80"
        />
        <el-table-column
          prop="title"
          :label="$t('dashboard.columnTitle')"
        />
        <el-table-column
          prop="project_type"
          :label="$t('dashboard.columnType')"
          width="120"
        />
        <el-table-column
          prop="status"
          :label="$t('dashboard.columnStatus')"
          width="110"
        >
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="progress"
          :label="$t('dashboard.columnProgress')"
          width="160"
        >
          <template #default="{ row }">
            <el-progress
              :percentage="row.progress || 0"
              :stroke-width="10"
            />
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
const ops = ref<any>(null)
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

const loadOps = async () => {
  try {
    ops.value = await api.get('/api/v1/admin/operations-stats')
  } catch { /* interceptor */ }
}

// 漏斗占比（相对发布项目数）
const funnelPct = (n?: number) => {
  const base = ops.value?.funnel?.published || 0
  if (!base) return 0
  return Math.round(((n || 0) / base) * 100)
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
  loadOps()
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
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  box-shadow: 0 6px 16px rgba(37, 99, 235, .22);
}
.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--eqs-text, #303133);
  line-height: 1.2;
}
.stat-label {
  font-size: 13px;
  color: var(--eqs-text-muted, #909399);
}
.panel {
  margin-top: 16px;
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
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
  color: var(--eqs-text, #303133);
  line-height: 1.6;
}
.ai-section {
  margin-top: 12px;
}
.ai-section-title {
  font-weight: 600;
  color: var(--eqs-text-secondary, #606266);
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.ai-section-title::before {
  content: '';
  width: 4px;
  height: 14px;
  border-radius: 2px;
  background: var(--eqs-gradient-ai, linear-gradient(135deg, #8b5cf6, #06b6d4));
}
.ai-list {
  margin: 0;
  padding-left: 18px;
  color: var(--eqs-text-secondary, #606266);
  font-size: 13px;
  line-height: 1.8;
}
.ai-list.success {
  color: var(--eqs-success, #67C23A);
}
.ai-loading {
  padding: 8px 0;
}
.ops-users {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.ops-user-item {
  background: var(--eqs-gradient-soft, #f5f7fb);
  border-radius: 10px;
  padding: 14px;
  text-align: center;
}
.ops-num {
  display: block;
  font-size: 26px;
  font-weight: 700;
  color: var(--eqs-primary, #2563eb);
}
.ops-label {
  display: block;
  font-size: 12px;
  color: var(--eqs-text-secondary, #64748b);
  margin-top: 4px;
}
.ops-funnel .funnel-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}
.funnel-label {
  width: 90px;
  font-size: 13px;
  color: var(--eqs-text-secondary, #64748b);
  flex-shrink: 0;
}
.funnel-row .el-progress {
  flex: 1;
}
.funnel-num {
  width: 50px;
  text-align: right;
  font-weight: 600;
  color: var(--eqs-text, #1e293b);
}
</style>
