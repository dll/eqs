<template>
  <div>
    <el-card>
      <template #header>订单管理</template>
      <el-table :data="orders" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="project_title" label="关联项目" />
        <el-table-column prop="amount" label="金额" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 3 ? 'success' : 'info'">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const orders = ref([
  { id: 1, project_title: '示例项目', amount: '¥50,000', status: 1 },
])

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: '待签约', 1: '进行中', 2: '待验收', 3: '已完成', 4: '纠纷中'
  }
  return map[status] || '未知'
}
</script>
