<template>
  <div>
    <el-card>
      <template #header>
        {{ $t('order.title') }}
      </template>
      <el-table
        :data="orders"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          :label="$t('order.id')"
          width="80"
        />
        <el-table-column
          :label="$t('order.relatedProject')"
          width="200"
        >
          <template #default="{ row }">
            {{ row.project?.title || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="project_id"
          :label="$t('order.projectId')"
          width="80"
        />
        <el-table-column
          :label="$t('order.supplier')"
          width="140"
        >
          <template #default="{ row }">
            {{ row.supplier?.company_name || ('#' + row.supplier_id) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('order.amount')"
          width="120"
        >
          <template #default="{ row }">
            ¥{{ row.amount }}
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          :label="$t('order.status')"
          width="100"
        >
          <template #default="{ row }">
            <el-tag :type="row.status === 3 ? 'success' : row.status === 4 ? 'danger' : 'info'">
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
    0: $t('order.status.pending'), 1: $t('order.status.inProgress'), 2: $t('order.status.toAccept'), 3: $t('order.status.completed'), 4: $t('order.status.dispute')
  }
  return map[status] || $t('order.status.unknown')
}
</script>