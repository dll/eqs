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
          width="170"
        >
          <template #default="{ row }">
            <el-button
              type="info"
              size="small"
              @click="viewDetail(row)"
            >
              {{ $t('user.detail') }}
            </el-button>
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

    <!-- 用户详情 -->
    <el-dialog
      v-model="detailVisible"
      :title="$t('user.detail')"
      width="520px"
    >
      <template v-if="detail">
        <el-descriptions
          :column="2"
          border
        >
          <el-descriptions-item :label="$t('user.id')">
            #{{ detail.user.id }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('user.phone')">
            {{ detail.user.phone }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('user.type')">
            {{ userTypeText(detail.user.user_type) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('user.company')">
            {{ detail.user.company_name || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('user.creditScore')">
            {{ detail.user.credit_score }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('user.status')">
            <el-tag :type="detail.user.status === 1 ? 'success' : 'danger'">
              {{ detail.user.status === 1 ? $t('user.status.active') : $t('user.status.disabled') }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
        <el-divider />
        <div class="stat-grid">
          <div class="stat-item">
            <div class="stat-num">
              {{ detail.stats.projects }}
            </div>
            <div class="stat-label">
              {{ $t('user.statProjects') }}
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-num">
              {{ detail.stats.orders_as_owner }}
            </div>
            <div class="stat-label">
              {{ $t('user.statOrdersOwner') }}
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-num">
              {{ detail.stats.orders_as_supplier }}
            </div>
            <div class="stat-label">
              {{ $t('user.statOrdersSupplier') }}
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-num">
              {{ detail.stats.qualifications }}
            </div>
            <div class="stat-label">
              {{ $t('user.statQualifications') }}
            </div>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

const users = ref<any[]>([])
const detailVisible = ref(false)
const detail = ref<any>(null)

const load = async () => {
  try {
    const res = await api.get<{ users: any[] }>('/api/v1/admin/users')
    users.value = res.users || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

const viewDetail = async (row: any) => {
  try {
    detail.value = await api.get(`/api/v1/admin/users/${row.id}`)
    detailVisible.value = true
  } catch {
    // interceptor 已提示
  }
}

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

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.stat-item {
  background: var(--eqs-gradient-soft, #f5f7fb);
  border-radius: 10px;
  padding: 14px;
  text-align: center;
}
.stat-num {
  font-size: 24px;
  font-weight: 700;
  color: var(--eqs-primary, #2563eb);
}
.stat-label {
  font-size: 12px;
  color: var(--eqs-text-secondary, #64748b);
  margin-top: 4px;
}
</style>