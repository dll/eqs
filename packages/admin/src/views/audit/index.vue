<template>
  <div>
    <el-card>
      <template #header>
        {{ $t('audit.title') }}
      </template>
      <el-table
        :data="auditList"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          :label="$t('audit.id')"
          width="80"
        />
        <el-table-column
          prop="supplier_id"
          :label="$t('audit.supplierId')"
          width="90"
        />
        <el-table-column
          prop="qualification_type"
          :label="$t('audit.qualType')"
          width="140"
        />
        <el-table-column
          prop="certificate_no"
          :label="$t('audit.certNo')"
        />
        <el-table-column
          prop="level"
          :label="$t('audit.level')"
          width="80"
        />
        <el-table-column
          prop="verification_status"
          :label="$t('audit.status')"
          width="100"
        >
          <template #default="{ row }">
            <el-tag :type="statusType(row.verification_status)">
              {{ statusText(row.verification_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('audit.actions')"
          width="180"
        >
          <template #default="{ row }">
            <template v-if="row.verification_status === 'pending'">
              <el-button
                type="success"
                size="small"
                @click="review(row, true)"
              >
                {{ $t('common.approve') }}
              </el-button>
              <el-button
                type="danger"
                size="small"
                @click="review(row, false)"
              >
                {{ $t('common.reject') }}
              </el-button>
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
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

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
    ElMessage.success(verified ? $t('common.passed') : $t('common.rejected'))
    load()
  } catch {
    // interceptor 已提示
  }
}

const statusText = (status: string) => {
  const map: Record<string, string> = {
    pending: $t('audit.status.pending'),
    approved: $t('audit.status.approved'),
    rejected: $t('audit.status.rejected'),
    expired: $t('audit.status.expired'),
  }
  return map[status] || status
}

const statusType = (status: string) => {
  const map: Record<string, string> = { pending: 'warning', approved: 'success', rejected: 'danger' }
  return map[status] || 'info'
}
</script>