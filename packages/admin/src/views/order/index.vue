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
            <el-tag :type="row.status === 3 ? 'success' : row.status === 4 ? 'danger' : row.status === 6 ? 'info' : 'info'">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('order.actions')"
          width="140"
        >
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="viewDetail(row)"
            >
              {{ $t('order.detail') }}
            </el-button>
            <el-button
              v-if="row.status === 0"
              type="danger"
              size="small"
              @click="cancelOrder(row)"
            >
              {{ $t('order.cancel') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()
const router = useRouter()

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

const viewDetail = (row: any) => {
  // 跳转项目页并携带订单ID（后台无独立订单详情页，用项目维度查看）
  router.push({ path: '/project', query: { id: row.project_id, order: row.id } })
}

const cancelOrder = (row: any) => {
  ElMessageBox.confirm($t('order.cancelConfirm'), $t('order.cancel'), { type: 'warning' })
    .then(async () => {
      try {
        await api.post(`/api/v1/order/${row.id}/cancel`, { reason: $t('order.cancelReasonDefault') })
        ElMessage.success($t('order.cancelled'))
        load()
      } catch {
        // interceptor 已提示
      }
    })
    .catch(() => {})
}

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: $t('order.status.pending'), 1: $t('order.status.inProgress'), 2: $t('order.status.toAccept'), 3: $t('order.status.completed'), 4: $t('order.status.dispute'), 6: $t('order.status.cancelled')
  }
  return map[status] || $t('order.status.unknown')
}
</script>