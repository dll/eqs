<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>项目管理</span>
        </div>
      </template>
      <el-table :data="projects" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="项目名称" />
        <el-table-column prop="service_type" label="类型" width="120" />
        <el-table-column prop="user_id" label="业主ID" width="80" />
        <el-table-column label="预算" width="150">
          <template #default="{ row }">¥{{ row.budget_min }} - ¥{{ row.budget_max }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag>{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/utils/request'

const projects = ref<any[]>([])

const load = async () => {
  try {
    const res = await api.get<{ projects: any[] }>('/api/v1/project/list')
    projects.value = res.projects || []
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
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>