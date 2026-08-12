<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('project.title') }}</span>
        </div>
      </template>
      <el-table
        :data="projects"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          :label="$t('project.id')"
          width="80"
        />
        <el-table-column
          prop="title"
          :label="$t('project.name')"
        />
        <el-table-column
          prop="service_type"
          :label="$t('project.type')"
          width="120"
        />
        <el-table-column
          prop="user_id"
          :label="$t('project.ownerId')"
          width="80"
        />
        <el-table-column
          :label="$t('project.budget')"
          width="150"
        >
          <template #default="{ row }">
            ¥{{ row.budget_min }} - ¥{{ row.budget_max }}
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          :label="$t('project.status')"
          width="100"
        >
          <template #default="{ row }">
            <el-tag :type="row.status === 5 ? 'info' : ''">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('project.actions')"
          width="220"
        >
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="openEdit(row)"
            >
              {{ $t('project.change') }}
            </el-button>
            <el-button
              type="warning"
              size="small"
              @click="withdrawProject(row)"
            >
              {{ $t('project.withdraw') }}
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="abolishProject(row)"
            >
              {{ $t('project.abolish') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 编辑项目 -->
    <el-dialog
      v-model="editVisible"
      :title="$t('project.change')"
      width="480px"
    >
      <el-form
        :model="editForm"
        label-width="90px"
      >
        <el-form-item :label="$t('project.name')">
          <el-input v-model="editForm.title" />
        </el-form-item>
        <el-form-item :label="$t('project.desc')">
          <el-input
            v-model="editForm.description"
            type="textarea"
            :rows="3"
          />
        </el-form-item>
        <el-form-item :label="$t('project.address')">
          <el-input v-model="editForm.address" />
        </el-form-item>
        <el-form-item :label="$t('project.budget')">
          <el-input-number
            v-model="editForm.budget_min"
            :min="0"
            :step="1000"
          />
          <span style="margin: 0 8px">-</span>
          <el-input-number
            v-model="editForm.budget_max"
            :min="0"
            :step="1000"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">
          {{ $t('common.cancel') }}
        </el-button>
        <el-button
          type="primary"
          @click="saveEdit"
        >
          {{ $t('common.confirm') }}
        </el-button>
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

const projects = ref<any[]>([])
const editVisible = ref(false)
const editForm = ref<any>({})

const load = async () => {
  try {
    const res = await api.get<{ projects: any[] }>('/api/v1/project/list')
    projects.value = res.projects || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

const openEdit = (row: any) => {
  editForm.value = {
    id: row.id,
    title: row.title,
    description: row.description,
    address: row.address,
    budget_min: row.budget_min,
    budget_max: row.budget_max,
  }
  editVisible.value = true
}

const saveEdit = async () => {
  try {
    const res = await api.put(`/api/v1/project/${editForm.value.id}`, {
      title: editForm.value.title,
      description: editForm.value.description,
      address: editForm.value.address,
      budget_min: editForm.value.budget_min,
      budget_max: editForm.value.budget_max,
    })
    editVisible.value = false
    ElMessage.success((res as any)?.locked ? $t('project.lockedHint') : $t('common.passed'))
    load()
  } catch {
    // interceptor 已提示
  }
}

const withdrawProject = (row: any) => {
  ElMessageBox.confirm($t('project.withdrawConfirm'), $t('project.withdraw'), { type: 'warning' })
    .then(async () => {
      try {
        await api.put(`/api/v1/project/${row.id}/withdraw`)
        ElMessage.success($t('project.withdrawn'))
        load()
      } catch {
        // interceptor 已提示
      }
    })
    .catch(() => {})
}

const abolishProject = (row: any) => {
  ElMessageBox.confirm($t('project.abolishConfirm'), $t('project.abolish'), { type: 'warning' })
    .then(async () => {
      try {
        await api.put(`/api/v1/project/${row.id}/abolish`)
        ElMessage.success($t('project.abolished'))
        load()
      } catch {
        // interceptor 已提示
      }
    })
    .catch(() => {})
}

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: $t('project.status.draft'), 1: $t('project.status.published'), 2: $t('project.status.assigned'), 3: $t('project.status.inProgress'), 4: $t('project.status.completed'), 5: $t('project.status.offline'), 6: $t('project.status.withdrawn'), 7: $t('project.status.abolished')
  }
  return map[status] || $t('project.status.unknown')
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>