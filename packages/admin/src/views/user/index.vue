<template>
  <div>
    <el-card>
      <template #header>用户管理</template>
      <el-table :data="users" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="phone" label="手机号" width="140" />
        <el-table-column prop="company_name" label="公司名称" />
        <el-table-column prop="user_type" label="类型" width="100">
          <template #default="{ row }">{{ userTypeText(row.user_type) }}</template>
        </el-table-column>
        <el-table-column prop="credit_score" label="信用分" width="80" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '正常' : '禁用' }}
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

const userTypeText = (t: number) => {
  const map: Record<number, string> = { 1: '甲方', 2: '服务方', 3: '管理员', 4: '评审专家' }
  return map[t] || '未知'
}
</script>