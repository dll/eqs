<template>
  <div>
    <el-card>
      <template #header>资质审核</template>
      <el-table :data="auditList" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="company_name" label="公司名称" />
        <el-table-column prop="type" label="资质类型" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? 'warning' : row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 0 ? '待审核' : row.status === 1 ? '已通过' : '已拒绝' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button type="success" size="small" @click="approve(row)" v-if="row.status === 0">通过</el-button>
            <el-button type="danger" size="small" @click="reject(row)" v-if="row.status === 0">拒绝</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const auditList = ref([
  { id: 1, company_name: '示例造价事务所', type: '造价', status: 0 },
])

const approve = (row: any) => {
  row.status = 1
}

const reject = (row: any) => {
  row.status = 2
}
</script>
