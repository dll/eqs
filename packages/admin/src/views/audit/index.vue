<template>
  <div>
    <el-card>
      <template #header>资质审核</template>
      <el-table :data="auditList" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="supplier_id" label="服务方ID" width="90" />
        <el-table-column prop="qualification_type" label="资质类型" width="140" />
        <el-table-column prop="certificate_no" label="证书编号" />
        <el-table-column prop="level" label="等级" width="80" />
        <el-table-column prop="verification_status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.verification_status)">{{ statusText(row.verification_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <template v-if="row.verification_status === 'pending'">
              <el-button type="success" size="small" @click="review(row, true)">通过</el-button>
              <el-button type="danger" size="small" @click="review(row, false)">拒绝</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '@/utils/request'

const auditList = ref<any[]>([])

const load = async () => {
  try {
    const res = await api.get<{ qualifications: any[] }>('/api/v1/admin/qualifications')
    auditList.value = res.qualifications || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

const review = async (row: any, verified: boolean) => {
  try {
    await api.post(`/api/v1/qualification/${row.id}/review`, { verified })
    ElMessage.success(verified ? '已通过' : '已拒绝')
    load()
  } catch {
    // interceptor 已提示
  }
}

const statusText = (status: string) => {
  const map: Record<string, string> = { pending: '待审核', approved: '已通过', rejected: '已拒绝', expired: '已过期' }
  return map[status] || status
}

const statusType = (status: string) => {
  const map: Record<string, string> = { pending: 'warning', approved: 'success', rejected: 'danger' }
  return map[status] || 'info'
}
</script>