<template>
  <div>
    <el-card>
      <template #header>用户管理</template>
      <el-table :data="users" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="phone" label="手机号" width="140" />
        <el-table-column prop="company_name" label="公司名称" />
        <el-table-column prop="user_type" label="类型" width="100">
          <template #default="{ row }">
            {{ row.user_type === 1 ? '甲方' : row.user_type === 2 ? '服务方' : '管理员' }}
          </template>
        </el-table-column>
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
import { ref } from 'vue'

const users = ref([
  { id: 1, phone: '138****8888', company_name: '示例甲方', user_type: 1, status: 1 },
  { id: 2, phone: '139****9999', company_name: '示例服务方', user_type: 2, status: 1 },
])
</script>
