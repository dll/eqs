<template>
  <div>
    <el-card>
      <template #header>结算中心（经持牌机构，非自建资金池）</template>
      <el-table :data="records" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="user_id" label="用户ID" width="80" />
        <el-table-column prop="order_id" label="订单ID" width="80" />
        <el-table-column prop="milestone_id" label="节点ID" width="80" />
        <el-table-column label="金额" width="120">
          <template #default="{ row }">¥{{ row.amount }}</template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="110">
          <template #default="{ row }">{{ typeText(row.type) }}</template>
        </el-table-column>
        <el-table-column prop="channel" label="通道" width="90" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : row.status === 2 ? 'danger' : 'info'">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/utils/request'

const records = ref<any[]>([])

const load = async () => {
  try {
    const res = await api.get<{ transactions: any[] }>('/api/v1/admin/transactions')
    records.value = res.transactions || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

const typeText = (t: string) => {
  const map: Record<string, string> = {
    payment: '支付', settlement: '结算', refund: '退款', freeze: '冻结', unfreeze: '解冻'
  }
  return map[t] || t
}

const statusText = (s: number) => {
  const map: Record<number, string> = { 0: '处理中', 1: '成功', 2: '失败' }
  return map[s] || '未知'
}
</script>