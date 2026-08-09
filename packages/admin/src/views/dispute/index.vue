<template>
  <div>
    <el-card>
      <template #header>{{ $t('dispute.title') }}</template>
      <el-table :data="disputes" style="width: 100%">
        <el-table-column prop="id" :label="$t('dispute.id')" width="80" />
        <el-table-column prop="order_id" :label="$t('dispute.orderId')" width="80" />
        <el-table-column prop="reason" :label="$t('dispute.reason')" show-overflow-tooltip />
        <el-table-column prop="claim" :label="$t('dispute.claim')" show-overflow-tooltip />
        <el-table-column prop="initiator_id" :label="$t('dispute.initiatorId')" width="90" />
        <el-table-column prop="status" :label="$t('dispute.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resolution_type" :label="$t('dispute.resolution')" width="100">
          <template #default="{ row }">{{ row.resolution_type || '-' }}</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

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
    evidence: $t('dispute.status.evidence'),
    review: $t('dispute.status.review'),
    mediation: $t('dispute.status.mediation'),
    reconsideration: $t('dispute.status.reconsideration'),
    closed: $t('dispute.status.closed'),
  }
  return map[status] || status
}

const statusType = (status: string) => {
  if (status === 'closed') return 'success'
  if (status === 'review' || status === 'mediation') return 'warning'
  return 'info'
}
</script>