<template>
  <div>
    <el-card>
      <template #header>系统配置</template>
      <el-form :model="settings" label-width="120px">
        <el-form-item label="佣金比例">
          <el-input-number v-model="settings.commission" :min="0" :max="100" />
          <span style="margin-left: 10px">%</span>
        </el-form-item>
        <el-form-item label="平台名称">
          <el-input v-model="settings.platformName" style="width: 300px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="save">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const settings = ref({
  commission: 5,
  platformName: '工程快捷服务',
})

const save = () => {
  // TODO: Save settings
}
</script>
