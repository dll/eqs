<template>
  <div>
    <el-card>
      <template #header>信用评分</template>
      <el-table :data="users" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="company_name" label="公司名称" />
        <el-table-column prop="credit_score" label="信用分" width="100" />
        <el-table-column prop="level" label="等级" width="100">
          <template #default="{ row }">
            <el-tag :type="levelType(row.level)">{{ row.level }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const users = ref([
  { id: 1, company_name: '示例造价事务所', credit_score: 95, level: 'AAA' },
])

const levelType = (level: string) => {
  const map: Record<string, string> = {
    AAA: 'success', AA: 'success', A: 'info', B: 'warning', C: 'danger'
  }
  return map[level] || 'info'
}
</script>
