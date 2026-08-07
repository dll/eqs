<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">1,234</div>
            <div class="stat-label">总用户数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">567</div>
            <div class="stat-label">总项目数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">¥890万</div>
            <div class="stat-label">总交易额</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">98.5%</div>
            <div class="stat-label">系统可用性</div>
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
import { ref } from 'vue'

const recentProjects = ref([
  { id: 1, title: '示例项目', project_type: '造价', status: 1 },
])

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: '草稿', 1: '已发布', 2: '已接单', 3: '进行中', 4: '已完成'
  }
  return map[status] || '未知'
}

const statusType = (status: number) => {
  const map: Record<number, string> = {
    0: 'info', 1: 'primary', 2: 'warning', 3: 'success', 4: 'success'
  }
  return map[status] || 'info'
}
</script>

<style scoped>
.stat-card {
  text-align: center;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 10px;
}

.stat-label {
  color: #909399;
}
</style>
