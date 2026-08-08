<template>
  <div>
    <el-card>
      <template #header>纠纷仲裁（专家评审+平台调解）</template>
      <el-table :data="disputes" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="order_id" label="订单ID" width="80" />
        <el-table-column prop="reason" label="纠纷原因" show-overflow-tooltip />
        <el-table-column prop="claim" label="诉求" show-overflow-tooltip />
        <el-table-column prop="initiator_id" label="发起人ID" width="90" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resolution_type" label="结案方式" width="100">
          <template #default="{ row }">{{ row.resolution_type || '-' }}</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/utils/request'

const disputes = ref<any[]>([])

const load = async () => {
  try {
    const res = await api.get<{ disputes: any[] }>('/api/v1/admin/disputes')
    disputes.value = res.disputes || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

const statusText = (status: string) => {
  const map: Record<string, string> = {
    evidence: '举证中', review: '评审中', mediation: '调解中', reconsideration: '复议中', closed: '已结案'
  }
  return map[status] || status
}

const statusType = (status: string) => {
  if (status === 'closed') return 'success'
  if (status === 'review' || status === 'mediation') return 'warning'
  return 'info'
}
</script>