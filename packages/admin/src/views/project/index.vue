<template>
  <div>
    <el-card>
      <template #header>项目管理</template>
      <el-table :data="projects" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="项目名称" />
        <el-table-column prop="project_type" label="类型" width="120" />
        <el-table-column prop="budget" label="预算" width="150" />
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
import { ref } from 'vue'

const projects = ref([
  { id: 1, title: '示例项目', project_type: '造价', budget: '10-50万', status: 1 },
])

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: '草稿', 1: '已发布', 2: '已接单', 3: '进行中', 4: '已完成'
  }
  return map[status] || '未知'
}
</script>
