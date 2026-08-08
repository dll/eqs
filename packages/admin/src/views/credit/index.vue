<template>
  <div>
    <el-card>
      <template #header>信用评分（评价联动）</template>
      <el-table :data="users" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="company_name" label="公司名称" />
        <el-table-column prop="credit_score" label="信用分" width="100">
          <template #default="{ row }">
            <el-tag :type="scoreType(row.credit_score)">{{ row.credit_score }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="等级" width="100">
          <template #default="{ row }">
            <el-tag :type="levelType(row.credit_score)">{{ levelText(row.credit_score) }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/utils/request'

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

const levelText = (score: number) => {
  if (score >= 90) return 'AAA'
  if (score >= 80) return 'AA'
  if (score >= 70) return 'A'
  if (score >= 60) return 'B'
  return 'C'
}

const levelType = (score: number) => {
  const map: Record<string, string> = { AAA: 'success', AA: 'success', A: 'info', B: 'warning', C: 'danger' }
  return map[levelText(score)] || 'info'
}

const scoreType = (score: number) => {
  if (score >= 80) return 'success'
  if (score >= 60) return 'warning'
  return 'danger'
}
</script>