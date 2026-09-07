<template>
  <div>
    <el-alert
      type="success"
      :closable="false"
      :title="$t('archive.title')"
      :description="$t('archive.hint')"
      style="margin-bottom: 12px"
    />
    <el-card>
      <template #header>
        <div class="bar">
          <span>{{ $t('archive.title') }}</span>
          <div class="bar-actions">
            <el-button
              type="primary"
              @click="openNew"
            >
              <el-icon style="margin-right: 4px"><Plus /></el-icon>
              {{ $t('archive.new') }}
            </el-button>
            <el-button
              @click="load"
            >{{ $t('common.refresh') }}</el-button>
          </div>
        </div>
      </template>

      <!-- 筛选 -->
      <div class="filters">
        <el-select
          v-model="filters.phase"
          :placeholder="$t('archive.filterPhase')"
          clearable
          style="width: 150px"
          @change="load"
        >
          <el-option
            v-for="p in phaseList"
            :key="p"
            :label="$t('archive.phase.' + p)"
            :value="p"
          />
        </el-select>
        <el-select
          v-model="filters.source"
          :placeholder="$t('archive.filterSource')"
          clearable
          style="width: 150px"
          @change="load"
        >
          <el-option
            v-for="s in ['manual', 'batch', 'migrate']"
            :key="s"
            :label="$t('archive.source.' + s)"
            :value="s"
          />
        </el-select>
        <el-input
          v-model="filters.q"
          :placeholder="$t('archive.filterQ')"
          clearable
          style="width: 260px"
          @keyup.enter="load"
        />
        <el-button
          type="primary"
          plain
          @click="load"
        >{{ $t('common.search') }}</el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="rows"
        style="width: 100%; margin-top: 8px"
      >
        <el-table-column
          :label="$t('archive.id')"
          prop="id"
          width="70"
        />
        <el-table-column
          :label="$t('archive.titleCol')"
          prop="title"
          min-width="180"
        />
        <el-table-column
          :label="$t('archive.ownerOrg')"
          prop="owner_org"
          width="130"
        />
        <el-table-column
          :label="$t('archive.serviceType')"
          width="96"
        >
          <template #default="{ row }">
            {{ serviceTypeText(row.service_type) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('archive.phase')"
          width="90"
        >
          <template #default="{ row }">
            <el-tag :type="phaseType(row.status_phase)">
              {{ $t('archive.phase.' + row.status_phase) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('archive.source')"
          width="76"
        >
          <template #default="{ row }">
            {{ $t('archive.source.' + row.source) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('archive.createdAt')"
          prop="created_at"
          width="150"
        >
          <template #default="{ row }">
            {{ fmt(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('archive.actions')"
          width="140"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              @click="openEdit(row)"
            >{{ $t('common.edit') }}</el-button>
            <el-button
              link
              type="danger"
              @click="remove(row)"
            >{{ $t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑 -->
    <el-dialog
      v-model="dialog"
      :title="editing ? $t('archive.editTitle') : $t('archive.newTitle')"
      width="560px"
    >
      <el-form
        ref="formRef"
        :model="form"
        label-width="110px"
      >
        <el-form-item
          :label="$t('archive.titleCol')"
          required
        >
          <el-input
            v-model="form.title"
            :placeholder="$t('archive.phTitle')"
          />
        </el-form-item>
        <el-form-item :label="$t('archive.ownerOrg')">
          <el-input v-model="form.owner_org" />
        </el-form-item>
        <el-form-item :label="$t('archive.serviceType')">
          <el-select
            v-model="form.service_type"
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="s in ['now','supervision','geotech','design']"
              :key="s"
              :label="$t('archive.serviceType.' + s)"
              :value="s"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('archive.phase')">
          <el-select
            v-model="form.status_phase"
            style="width: 100%"
          >
            <el-option
              v-for="p in phaseList"
              :key="p"
              :label="$t('archive.phase.' + p)"
              :value="p"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('archive.amount')">
          <el-input-number
            v-model="form.amount"
            :min="0"
            :precision="2"
            :controls="false"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('archive.phDesc')">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="2"
          />
        </el-form-item>
        <el-form-item :label="$t('archive.phAddr')">
          <el-input v-model="form.address" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          @click="save"
        >{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

const phaseList = ['preparing', 'in_progress', 'delivered', 'settled', 'liquidation']

const rows = ref<any[]>([])
const loading = ref(false)
const filters = reactive<{ phase?: string; source?: string; q?: string }>({})
const dialog = ref(false)
const editing = ref<any>(null)

const form = reactive<any>({
  title: '',
  owner_org: '',
  service_type: 'now',
  status_phase: 'preparing',
  amount: 0,
  description: '',
  address: '',
})

const load = async () => {
  loading.value = true
  try {
    const params: any = { page: 1, size: 100 }
    if (filters.phase) params.phase = filters.phase
    if (filters.source) params.source = filters.source
    if (filters.q) params.q = filters.q
    const res = await api.get<{ archived_projects: any[] }>('/api/v1/admin/archive/projects', params)
    rows.value = res.archived_projects || []
  } catch {
    // interceptor
  } finally {
    loading.value = false
  }
}

const serviceTypeText = (v?: string) => $t(`archive.serviceType.${v || 'other'}`)
const phaseType = (p: string) =>
  ({ preparing: 'info', in_progress: 'warning', delivered: 'success', settled: 'success', liquidation: 'danger' } as Record<string, string>)[p] || 'info'
const fmt = (s?: string) => (s ? String(s).slice(0, 10) : '-')

const openNew = () => {
  editing.value = null
  Object.assign(form, { title: '', owner_org: '', service_type: 'now', status_phase: 'preparing', amount: 0, description: '', address: '' })
  dialog.value = true
}
const openEdit = (row: any) => {
  editing.value = row
  Object.assign(form, {
    title: row.title, owner_org: row.owner_org, service_type: row.service_type || 'now',
    status_phase: row.status_phase, amount: row.amount || 0, description: row.description || '', address: row.address || '',
  })
  dialog.value = true
}

const save = async () => {
  if (!form.title) {
    ElMessage.warning($t('archive.phTitle'))
    return
  }
  const payload = { ...form }
  try {
    if (editing.value) {
      await api.put(`/api/v1/admin/archive/projects/${editing.value.id}`, payload)
    } else {
      await api.post('/api/v1/admin/archive/projects', payload)
    }
    ElMessage.success($t('common.save'))
    dialog.value = false
    load()
  } catch {
    // interceptor
  }
}

const remove = async (row: any) => {
  try {
    await ElMessageBox.confirm($t('archive.deleteMsg'), $t('common.prompt'), { type: 'warning' })
  } catch {
    return
  }
  try {
    await api.delete(`/api/v1/admin/archive/projects/${row.id}`)
    ElMessage.success($t('archive.deleted'))
    load()
  } catch {
    // interceptor
  }
}

onMounted(load)
</script>

<style scoped>
.bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.filters {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
