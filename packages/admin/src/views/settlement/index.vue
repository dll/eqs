<template>
  <div>
    <el-card>
      <template #header>结算中心</template>
      <el-table :data="records" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="user_name" label="用户" />
        <el-table-column prop="amount" label="金额" width="120" />
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '成功' : '处理中' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const records = ref([
  { id: 1, user_name: '示例用户', amount: '¥10,000', type: '充值', status: 1 },
])
</script>
