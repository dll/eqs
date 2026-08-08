<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.user_count }}</div>
            <div class="stat-label">总用户数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.project_count }}</div>
            <div class="stat-label">总项目数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.order_count }}</div>
            <div class="stat-label">总订单数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">¥{{ stats.settled_amount }}</div>
            <div class="stat-label">累计结算额</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top: 20px">
      <template #header>最近项目</template>
      <el-table :data="recentProjects" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="项目名称" />
        <el-table-column prop="project_type" label="类型" width="120" />
        <el-table-column prop="status" label="状态" width="100">
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
    0: '草稿', 1: '已发布', 2: '已接单', 3: '进行中', 4: '已完成'
  }
  return map[status] || '未知'
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