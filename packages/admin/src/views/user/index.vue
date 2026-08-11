<template>
  <div>
    <el-card>
      <template #header>
        {{ $t('user.title') }}
      </template>
      <el-table
        :data="users"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          :label="$t('user.id')"
          width="80"
        />
        <el-table-column
          prop="phone"
          :label="$t('user.phone')"
          width="140"
        />
        <el-table-column
          prop="company_name"
          :label="$t('user.company')"
        />
        <el-table-column
          prop="user_type"
          :label="$t('user.type')"
          width="100"
        >
          <template #default="{ row }">
            {{ userTypeText(row.user_type) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="credit_score"
          :label="$t('user.creditScore')"
          width="80"
        />
        <el-table-column
          prop="status"
          :label="$t('user.status')"
          width="80"
        >
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? $t('user.status.active') : $t('user.status.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('user.actions')"
          width="110"
        >
          <template #default="{ row }">
            <el-button
              :type="row.status === 1 ? 'danger' : 'success'"
              size="small"
              @click="toggleStatus(row)"
            >
              {{ row.status === 1 ? $t('user.disable') : $t('user.enable') }}
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
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

const users = ref<any[]>([])

const load = async () => {
  try {
    const res = await api.get<{ users: any[] }>('/api/v1/admin/users')
    users.value = res.users || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

const toggleStatus = (row: any) => {
  const next = row.status === 1 ? 0 : 1
  const action = next === 1 ? $t('user.enable') : $t('user.disable')
  ElMessageBox.confirm(`${action} #${row.id}?`, $t('user.actions'), { type: 'warning' })
    .then(async () => {
      try {
        await api.put(`/api/v1/admin/users/${row.id}/status`, { status: next })
        ElMessage.success($t('common.passed'))
        load()
      } catch {
        // interceptor 已提示
      }
    })
    .catch(() => {})
}

const userTypeText = (t: number) => {
  const map: Record<number, string> = { 1: $t('role.client'), 2: $t('role.supplier'), 3: $t('role.admin'), 4: $t('role.expert') }
  return map[t] || $t('role.unknown')
}
</script>