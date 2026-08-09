<template>
  <div class="settings-page">
    <el-card style="margin-bottom: 20px">
      <template #header>
        <div style="display: flex; align-items: center; justify-content: space-between;">
          <span>系统配置</span>
          <el-select v-model="filterCategory" placeholder="全部分类" clearable style="width: 160px">
            <el-option label="全部" value="" />
            <el-option label="主题" value="theme" />
            <el-option label="国际化" value="i18n" />
            <el-option label="多端" value="multiplatform" />
            <el-option label="版本" value="version" />
            <el-option label="演示数据" value="demo" />
            <el-option label="系统" value="system" />
          </el-select>
        </div>
      </template>
      <el-table :data="filteredConfigs" border stripe>
        <el-table-column prop="config_key" label="配置键" width="240" />
        <el-table-column prop="config_value" label="配置值" />
        <el-table-column prop="value_type" label="类型" width="80" />
        <el-table-column prop="description" label="说明" />
        <el-table-column prop="is_public" label="公开" width="70">
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
          <el-button @click="loadDemoStatus">刷新状态</el-button>
        </el-form-item>
      </el-form>
      <el-descriptions :column="4" border size="small" style="margin-top: 12px;">
        <el-descriptions-item label="状态">{{ demoStatus.demo_mode ? '已开启' : '已关闭' }}</el-descriptions-item>
        <el-descriptions-item label="用户数">{{ demoStatus.user_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="项目数">{{ demoStatus.project_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="订单数">{{ demoStatus.order_count || 0 }}</el-descriptions-item>
      </el-descriptions>
      <el-alert type="info" :closable="false" title="演示数据用于测试系统功能、演示交流、培训教程，生成与清理均写入审计日志。" style="margin-top: 12px;" />
    </el-card>

    <el-card style="margin-bottom: 20px">
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

    <el-card>
      <template #header>性能监控</template>
      <el-table :data="monitorStatsList" border stripe>
        <el-table-column prop="path" label="接口" />
        <el-table-column prop="count" label="请求次数" width="100" />
        <el-table-column prop="error_count" label="错误次数" width="100" />
        <el-table-column prop="avg_ms" label="平均耗时(ms)" width="120">
          <template #default="{ row }">{{ row.avg_ms?.toFixed(1) || '0' }}</template>
        </el-table-column>
        <el-table-column prop="error_rate" label="错误率(%)" width="100">
          <template #default="{ row }">{{ row.error_rate?.toFixed(1) || '0' }}%</template>
        </el-table-column>
      </el-table>
      <el-button style="margin-top: 15px" @click="loadMonitorStats">刷新统计</el-button>
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
import { ref, computed, onMounted } from 'vue'
import { api } from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const configs = ref<any[]>([])
const versions = ref<any[]>([])
const configDialog = ref(false)
const versionDialog = ref(false)
const demoMode = ref('demo')
const filterCategory = ref('')
const configForm = ref<any>({ config_key: '', config_value: '', value_type: 'string', description: '', is_public: false })
const versionForm = ref<any>({ version: '', build: 0, platform: 'all', update_url: '', release_notes: '', mandatory: false })
const demoStatus = ref<any>({ demo_mode: false, user_count: 0, project_count: 0, order_count: 0, dispute_count: 0 })
const monitorStats = ref<any>({})
const monitorStatsList = computed(() =>
  Object.entries(monitorStats.value).map(([path, stat]: [string, any]) => ({ path, ...stat }))
)

const filteredConfigs = computed(() => {
  if (!filterCategory.value) return configs.value
  return configs.value.filter(c => c.config_key.startsWith(filterCategory.value + '.'))
})

const loadConfigs = async () => {
  const res = await api.get('/api/v1/admin/config/list')
  configs.value = res.configs || []
}

const loadVersions = async () => {
  const res = await api.get('/api/v1/admin/version/list')
  versions.value = res.versions || []
}

const loadDemoStatus = async () => {
  try {
    const res = await api.get('/api/v1/admin/demo/status')
    demoStatus.value = res
  } catch { /* ignore */ }
}

const loadMonitorStats = async () => {
  try {
    const res = await api.get('/api/v1/admin/monitor/stats')
    monitorStats.value = res.stats || {}
  } catch { /* ignore */ }
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
  await loadDemoStatus()
}

const cleanDemo = async () => {
  await ElMessageBox.confirm('确认清理所有演示数据？', '提示')
  await api.post('/api/v1/admin/demo/clean')
  ElMessage.success('演示数据已清理')
  await loadDemoStatus()
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
  loadDemoStatus()
  loadMonitorStats()
})
</script>
