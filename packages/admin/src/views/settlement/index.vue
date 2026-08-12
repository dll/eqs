<template>
  <div>
    <el-card>
      <template #header>
        {{ $t('settlement.title') }}
      </template>
      <el-table
        :data="records"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          :label="$t('settlement.id')"
          width="80"
        />
        <el-table-column
          prop="user_id"
          :label="$t('settlement.userId')"
          width="80"
        />
        <el-table-column
          prop="order_id"
          :label="$t('settlement.orderId')"
          width="80"
        />
        <el-table-column
          prop="milestone_id"
          :label="$t('settlement.milestoneId')"
          width="80"
        />
        <el-table-column
          :label="$t('settlement.amount')"
          width="120"
        >
          <template #default="{ row }">
            ¥{{ row.amount }}
          </template>
        </el-table-column>
        <el-table-column
          prop="type"
          :label="$t('settlement.type')"
          width="110"
        >
          <template #default="{ row }">
            {{ typeText(row.type) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="channel"
          :label="$t('settlement.channel')"
          width="90"
        />
        <el-table-column
          prop="status"
          :label="$t('settlement.status')"
          width="90"
        >
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : row.status === 2 ? 'danger' : 'info'">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 平台佣金（V10：收取操作） -->
    <el-card
      style="margin-top: 16px"
    >
      <template #header>
        {{ $t('settlement.commission') }}
      </template>
      <el-table
        :data="commissions"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          label="ID"
          width="80"
        />
        <el-table-column
          prop="user_id"
          :label="$t('settlement.userId')"
          width="90"
        />
        <el-table-column
          prop="order_id"
          :label="$t('settlement.orderId')"
          width="90"
        />
        <el-table-column
          :label="$t('settlement.amount')"
          width="130"
        >
          <template #default="{ row }">
            ¥{{ row.commission ?? row.amount }}
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          :label="$t('settlement.status')"
          width="110"
        >
          <template #default="{ row }">
            <el-tag :type="row.status === 'collected' ? 'success' : 'warning'">
              {{ commissionStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('settlement.actions')"
          width="110"
        >
          <template #default="{ row }">
            <el-button
              v-if="row.status !== 'collected'"
              type="primary"
              size="small"
              @click="collect(row)"
            >
              {{ $t('settlement.collect') }}
            </el-button>
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

const records = ref<any[]>([])
const commissions = ref<any[]>([])

const load = async () => {
  try {
    const res = await api.get<{ transactions: any[] }>('/api/v1/admin/transactions')
    records.value = res.transactions || []
  } catch {
    // interceptor 已提示
  }
  try {
    const cres = await api.get<{ commissions: any[] }>('/api/v1/admin/commission/list')
    commissions.value = cres.commissions || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

// V10：收取平台佣金
const collect = async (row: any) => {
  try {
    await api.post(`/api/v1/admin/commission/${row.id}/collect`)
    ElMessage.success($t('settlement.collected'))
    load()
  } catch {
    // interceptor 已提示
  }
}

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

const commissionStatusText = (s: string) => {
  const map: Record<string, string> = {
    pending: $t('settlement.commission.pending'),
    collected: $t('settlement.commission.collected'),
  }
  return map[s] || s
}
</script>