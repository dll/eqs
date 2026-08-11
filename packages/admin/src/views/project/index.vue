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
            <el-tag>{{ statusText(row.status) }}</el-tag>
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

const projects = ref<any[]>([])

const load = async () => {
  try {
    const res = await api.get<{ projects: any[] }>('/api/v1/project/list')
    projects.value = res.projects || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

const statusText = (status: number) => {
  const map: Record<number, string> = {
    0: $t('project.status.draft'), 1: $t('project.status.published'), 2: $t('project.status.assigned'), 3: $t('project.status.inProgress'), 4: $t('project.status.completed')
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