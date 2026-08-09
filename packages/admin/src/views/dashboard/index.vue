<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.user_count }}</div>
            <div class="stat-label">{{ $t('dashboard.totalUsers') }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.project_count }}</div>
            <div class="stat-label">{{ $t('dashboard.totalProjects') }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.order_count }}</div>
            <div class="stat-label">{{ $t('dashboard.totalOrders') }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">¥{{ stats.settled_amount }}</div>
            <div class="stat-label">{{ $t('dashboard.totalSettled') }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top: 20px">
      <template #header>{{ $t('dashboard.recentProjects') }}</template>
      <el-table :data="recentProjects" style="width: 100%">
        <el-table-column prop="id" :label="$t('dashboard.columnId')" width="80" />
        <el-table-column prop="title" :label="$t('dashboard.columnTitle')" />
        <el-table-column prop="project_type" :label="$t('dashboard.columnType')" width="120" />
        <el-table-column prop="status" :label="$t('dashboard.columnStatus')" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

const stats = ref({ user_count: 0, project_count: 0, order_count: 0, dispute_count: 0, settled_amount: 0 })
const recentProjects = ref<any[]>([])

const load = async () => {
  try {
    const s = await api.get('/api/v1/admin/stats')
    stats.value = s
    const p = await api.get<{ projects: any[] }>('/api/v1/project/list')
    recentProjects.value = (p.projects || []).slice(0, 8)
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

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
  text-align: center;
  padding: 20px 0;
}

.stat-value {
  font-size: 34px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  margin-top: 8px;
  color: #909399;
  font-size: 14px;
}
</style>