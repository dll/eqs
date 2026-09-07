<template>
  <div>
    <el-alert
      type="info"
      :closable="false"
      :title="$t('auditlog.title')"
      :description="$t('auditlog.hint')"
      style="margin-bottom: 12px"
    />
    <el-card>
      <div class="filters">
        <el-input
          v-model="filters.action"
          :placeholder="$t('auditlog.filterAction')"
          clearable
          style="width: 220px"
          @keyup.enter="load"
        />
        <el-button
          type="primary"
          @click="load"
        >{{ $t('common.search') }}</el-button>
        <el-button @click="reset">{{ $t('common.refresh') }}</el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="rows"
        style="width: 100%; margin-top: 8px"
      >
        <el-table-column
          :label="$t('auditlog.id')"
          prop="id"
          width="70"
        />
        <el-table-column
          :label="$t('auditlog.userId')"
          prop="user_id"
          width="90"
        />
        <el-table-column
          :label="$t('auditlog.action')"
          prop="action"
          min-width="160"
        />
        <el-table-column
          :label="$t('auditlog.targetType')"
          prop="target_type"
          width="150"
        />
        <el-table-column
          :label="$t('auditlog.targetId')"
          prop="target_id"
          width="90"
        />
        <el-table-column
          :label="$t('auditlog.ip')"
          prop="ip"
          width="120"
        />
        <el-table-column
          :label="$t('auditlog.createdAt')"
          width="170"
        >
          <template #default="{ row }">
            {{ fmt(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('auditlog.detail')">
          <template #default="{ row }">
            <span class="detail">{{ row.detail }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

const rows = ref<any[]>([])
const loading = ref(false)
const filters = reactive({ action: '' })

const load = async () => {
  loading.value = true
  try {
    const params: any = { page: 1, size: 100 }
    if (filters.action) params.action = filters.action
    const res = await api.get<{ logs: any[] }>('/api/v1/admin/audit/logs', params)
    rows.value = res.logs || []
  } catch {
    // interceptor
  } finally {
    loading.value = false
  }
}

const reset = () => {
  filters.action = ''
  load()
}

const fmt = (s?: string) => {
  if (!s) return '-'
  const d = new Date(s)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

onMounted(load)
</script>

<style scoped>
.filters {
  display: flex;
  gap: 8px;
}
.detail {
  font-size: 12px;
  color: #909399;
  font-family: monospace;
  word-break: break-all;
}
</style>
