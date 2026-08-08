<template>
  <div>
    <el-card>
      <template #header>订单管理</template>
      <el-table :data="orders" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="关联项目" width="200">
          <template #default="{ row }">{{ row.project?.title || '-' }}</template>
        </el-table-column>
        <el-table-column prop="project_id" label="项目ID" width="80" />
        <el-table-column label="服务方" width="140">
          <template #default="{ row }">{{ row.supplier?.company_name || ('#' + row.supplier_id) }}</template>
        </el-table-column>
        <el-table-column label="金额" width="120">
          <template #default="{ row }">¥{{ row.amount }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 3 ? 'success' : row.status === 4 ? 'danger' : 'info'">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/utils/request'

const orders = ref<any[]>([])

const load = async () => {
  try {
    const res = await api.get<{ orders: any[] }>('/api/v1/admin/orders')
    orders.value = res.orders || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: '待签约', 1: '进行中', 2: '待验收', 3: '已完成', 4: '纠纷中'
  }
  return map[status] || '未知'
}
</script>