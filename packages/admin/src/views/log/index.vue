<template>
  <div>
    <el-card>
      <template #header>
        <div class="head">
          <span>{{ $t('log.title') }}</span>
          <div class="filters">
            <el-input
              v-model="filters.userId"
              :placeholder="$t('log.filterUser')"
              size="small"
              style="width: 140px"
              clearable
              @change="load(1)"
            />
            <el-input
              v-model="filters.action"
              :placeholder="$t('log.filterAction')"
              size="small"
              style="width: 160px"
              clearable
              @change="load(1)"
            />
            <el-button
              size="small"
              type="primary"
              @click="load(1)"
            >
              {{ $t('common.search') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="logs"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          :label="$t('log.id')"
          width="70"
        />
        <el-table-column
          prop="user_id"
          :label="$t('log.userId')"
          width="80"
        />
        <el-table-column
          prop="action"
          :label="$t('log.action')"
          width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="target_type"
          :label="$t('log.targetType')"
          width="110"
        />
        <el-table-column
          prop="target_id"
          :label="$t('log.targetId')"
          width="80"
        />
        <el-table-column
          prop="detail"
          :label="$t('log.detail')"
          show-overflow-tooltip
        />
        <el-table-column
          prop="ip"
          :label="$t('log.ip')"
          width="120"
        />
        <el-table-column
          prop="created_at"
          :label="$t('log.time')"
          width="170"
        >
          <template #default="{ row }">
            {{ fmtTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('log.restore')"
          width="110"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-if="row.target_type === 'config'"
              size="small"
              type="warning"
              plain
              @click="restore(row)"
            >
              {{ $t('log.restoreBtn') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pager"
        layout="total, prev, pager, next"
        :total="total"
        :page-size="size"
        :current-page="page"
        @current-change="load"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()
const logs = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = 50
const filters = reactive({ userId: '', action: '' })

const load = async (p: number) => {
  page.value = p
  try {
    const params: any = { page: p, size }
    if (filters.userId) params.user_id = filters.userId
    if (filters.action) params.action = filters.action
    const r = await api.get('/api/v1/admin/log/list', { params })
    logs.value = r.logs || []
    total.value = r.total || 0
  } catch { /* interceptor */ }
}

const restore = async (row: any) => {
  try {
    await api.post('/api/v1/admin/log/restore-config', { log_id: row.id })
    ElMessage.success($t('log.restoreSuccess'))
  } catch { /* interceptor */ }
}

const fmtTime = (s: string) => (s ? s.replace('T', ' ').slice(0, 19) : '')

onMounted(() => load(1))
</script>

<style scoped>
.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.filters {
  display: flex;
  gap: 8px;
}
.pager {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>
