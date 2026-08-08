<template>
  <div class="settings-page">
    <el-card style="margin-bottom: 20px">
      <template #header>系统配置</template>
      <el-table :data="configs" border stripe>
        <el-table-column prop="config_key" label="配置键" width="220" />
        <el-table-column prop="config_value" label="配置值" />
        <el-table-column prop="value_type" label="类型" width="100" />
        <el-table-column prop="description" label="说明" />
        <el-table-column prop="is_public" label="公开" width="80">
          <template #default="{ row }">{{ row.is_public ? '是' : '否' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="editConfig(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteConfig(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button type="primary" style="margin-top: 15px" @click="openConfigDialog()">新增配置</el-button>
    </el-card>

    <el-card style="margin-bottom: 20px">
      <template #header>演示数据管理</template>
      <el-form inline>
        <el-form-item label="模式">
          <el-select v-model="demoMode" style="width: 160px">
            <el-option label="演示交流" value="demo" />
            <el-option label="功能测试" value="test" />
            <el-option label="培训教程" value="training" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="success" @click="seedDemo">生成演示数据</el-button>
          <el-button type="warning" @click="cleanDemo">清理演示数据</el-button>
        </el-form-item>
      </el-form>
      <el-alert type="info" :closable="false" title="演示数据用于测试系统功能、演示交流、培训教程，生成与清理均写入审计日志。" />
    </el-card>

    <el-card>
      <template #header>版本管理</template>
      <el-table :data="versions" border stripe>
        <el-table-column prop="version" label="版本号" width="100" />
        <el-table-column prop="build" label="构建号" width="90" />
        <el-table-column prop="platform" label="平台" width="110" />
        <el-table-column prop="release_notes" label="更新说明" />
        <el-table-column prop="mandatory" label="强制" width="80">
          <template #default="{ row }">{{ row.mandatory ? '是' : '否' }}</template>
        </el-table-column>
      </el-table>
      <el-button type="primary" style="margin-top: 15px" @click="openVersionDialog">发布新版本</el-button>
    </el-card>

    <el-dialog v-model="configDialog" title="配置项" width="500px">
      <el-form :model="configForm" label-width="100px">
        <el-form-item label="配置键"><el-input v-model="configForm.config_key" /></el-form-item>
        <el-form-item label="配置值"><el-input v-model="configForm.config_value" /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="configForm.value_type" style="width: 200px">
            <el-option label="字符串" value="string" />
            <el-option label="整数" value="int" />
            <el-option label="布尔" value="bool" />
            <el-option label="JSON" value="json" />
          </el-select>
        </el-form-item>
        <el-form-item label="说明"><el-input v-model="configForm.description" /></el-form-item>
        <el-form-item label="公开"><el-switch v-model="configForm.is_public" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configDialog = false">取消</el-button>
        <el-button type="primary" @click="saveConfig">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="versionDialog" title="发布新版本" width="500px">
      <el-form :model="versionForm" label-width="100px">
        <el-form-item label="版本号"><el-input v-model="versionForm.version" placeholder="如 1.1.0" /></el-form-item>
        <el-form-item label="构建号"><el-input-number v-model="versionForm.build" /></el-form-item>
        <el-form-item label="平台">
          <el-select v-model="versionForm.platform" style="width: 200px">
            <el-option label="全部" value="all" />
            <el-option label="H5" value="h5" />
            <el-option label="小程序" value="mp-weixin" />
            <el-option label="App" value="app" />
          </el-select>
        </el-form-item>
        <el-form-item label="更新地址"><el-input v-model="versionForm.update_url" /></el-form-item>
        <el-form-item label="更新说明"><el-input type="textarea" v-model="versionForm.release_notes" /></el-form-item>
        <el-form-item label="强制更新"><el-switch v-model="versionForm.mandatory" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="versionDialog = false">取消</el-button>
        <el-button type="primary" @click="publishVersion">发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const configs = ref<any[]>([])
const versions = ref<any[]>([])
const configDialog = ref(false)
const versionDialog = ref(false)
const demoMode = ref('demo')
const configForm = ref<any>({ config_key: '', config_value: '', value_type: 'string', description: '', is_public: false })
const versionForm = ref<any>({ version: '', build: 0, platform: 'all', update_url: '', release_notes: '', mandatory: false })

const loadConfigs = async () => {
  const res = await api.get('/api/v1/admin/config/list')
  configs.value = res.configs || []
}

const loadVersions = async () => {
  const res = await api.get('/api/v1/admin/version/list')
  versions.value = res.versions || []
}

const openConfigDialog = (row?: any) => {
  configForm.value = row
    ? { ...row }
    : { config_key: '', config_value: '', value_type: 'string', description: '', is_public: false }
  configDialog.value = true
}

const editConfig = (row: any) => openConfigDialog(row)

const saveConfig = async () => {
  await api.post('/api/v1/admin/config/upsert', configForm.value)
  ElMessage.success('配置已保存')
  configDialog.value = false
  await loadConfigs()
}

const deleteConfig = async (row: any) => {
  await ElMessageBox.confirm(`确认删除配置 ${row.config_key}？`, '提示')
  await api.delete(`/api/v1/admin/config/delete/${row.config_key}`)
  ElMessage.success('已删除')
  await loadConfigs()
}

const seedDemo = async () => {
  await api.post(`/api/v1/admin/demo/seed?mode=${demoMode.value}`)
  ElMessage.success('演示数据已生成')
}

const cleanDemo = async () => {
  await ElMessageBox.confirm('确认清理所有演示数据？', '提示')
  await api.post('/api/v1/admin/demo/clean')
  ElMessage.success('演示数据已清理')
}

const openVersionDialog = () => {
  versionDialog.value = true
}

const publishVersion = async () => {
  await api.post('/api/v1/admin/version/publish', versionForm.value)
  ElMessage.success('版本已发布')
  versionDialog.value = false
  await loadVersions()
}

onMounted(() => {
  loadConfigs()
  loadVersions()
})
</script>
