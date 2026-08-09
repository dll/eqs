<template>
  <div>
    <el-card>
      <template #header>{{ $t('settlement.title') }}</template>
      <el-table :data="records" style="width: 100%">
        <el-table-column prop="id" :label="$t('settlement.id')" width="80" />
        <el-table-column prop="user_id" :label="$t('settlement.userId')" width="80" />
        <el-table-column prop="order_id" :label="$t('settlement.orderId')" width="80" />
        <el-table-column prop="milestone_id" :label="$t('settlement.milestoneId')" width="80" />
        <el-table-column :label="$t('settlement.amount')" width="120">
          <template #default="{ row }">¥{{ row.amount }}</template>
        </el-table-column>
        <el-table-column prop="type" :label="$t('settlement.type')" width="110">
          <template #default="{ row }">{{ typeText(row.type) }}</template>
        </el-table-column>
        <el-table-column prop="channel" :label="$t('settlement.channel')" width="90" />
        <el-table-column prop="status" :label="$t('settlement.status')" width="90">
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
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

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
    payment: $t('settlement.type.payment'),
    settlement: $t('settlement.type.settlement'),
    refund: $t('settlement.type.refund'),
    freeze: $t('settlement.type.freeze'),
    unfreeze: $t('settlement.type.unfreeze'),
  }
  return map[t] || t
}

const statusText = (s: number) => {
  const map: Record<number, string> = { 0: $t('settlement.status.processing'), 1: $t('settlement.status.success'), 2: $t('settlement.status.failed') }
  return map[s] || $t('settlement.status.unknown')
}
</script>