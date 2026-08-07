<template>
  <div>
    <el-card>
      <template #header>纠纷仲裁</template>
      <el-table :data="disputes" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="order_id" label="订单ID" width="80" />
        <el-table-column prop="reason" label="纠纷原因" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? 'warning' : 'success'">
              {{ row.status === 0 ? '待处理' : '已裁决' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const disputes = ref([
  { id: 1, order_id: 100, reason: '交付物质量不达标', status: 0 },
])
</script>
