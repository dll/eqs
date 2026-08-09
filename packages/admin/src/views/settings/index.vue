<template>
  <div class="settings-page">
    <el-card style="margin-bottom: 20px">
      <template #header>
        <div style="display: flex; align-items: center; justify-content: space-between;">
          <span>{{ $t('settings.configCenter') }}</span>
          <el-select v-model="filterCategory" :placeholder="$t('settings.categoryAll')" clearable style="width: 160px">
            <el-option :label="$t('settings.categoryAll')" value="" />
            <el-option :label="$t('settings.categoryTheme')" value="theme" />
            <el-option :label="$t('settings.categoryI18n')" value="i18n" />
            <el-option :label="$t('settings.categoryMulti')" value="multiplatform" />
            <el-option :label="$t('settings.categoryVersion')" value="version" />
            <el-option :label="$t('settings.categoryDemo')" value="demo" />
            <el-option :label="$t('settings.categorySystem')" value="system" />
          </el-select>
        </div>
      </template>
      <el-table :data="filteredConfigs" border stripe>
        <el-table-column prop="config_key" :label="$t('settings.configKey')" width="240" />
        <el-table-column prop="config_value" :label="$t('settings.configValue')" />
        <el-table-column prop="value_type" :label="$t('settings.valueType')" width="80" />
        <el-table-column prop="description" :label="$t('settings.description')" />
        <el-table-column prop="is_public" :label="$t('settings.isPublic')" width="70">
          <template #default="{ row }">{{ row.is_public ? $t('common.yes') : $t('common.no') }}</template>
        </el-table-column>
        <el-table-column :label="$t('settings.actions')" width="140">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="editConfig(row)">{{ $t('common.edit') }}</el-button>
            <el-button size="small" type="danger" @click="deleteConfig(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button type="primary" style="margin-top: 15px" @click="openConfigDialog()">{{ $t('settings.addConfig') }}</el-button>
    </el-card>

    <el-card style="margin-bottom: 20px">
      <template #header>{{ $t('settings.demo') }}</template>
      <el-form inline>
        <el-form-item :label="$t('settings.demoMode')">
          <el-select v-model="demoMode" style="width: 160px">
            <el-option :label="$t('settings.demoModeDemo')" value="demo" />
            <el-option :label="$t('settings.demoModeTest')" value="test" />
            <el-option :label="$t('settings.demoModeTraining')" value="training" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="success" @click="seedDemo">{{ $t('settings.demoSeed') }}</el-button>
          <el-button type="warning" @click="cleanDemo">{{ $t('settings.demoClean') }}</el-button>
          <el-button @click="loadDemoStatus">{{ $t('settings.demoRefresh') }}</el-button>
        </el-form-item>
      </el-form>
      <el-descriptions :column="4" border size="small" style="margin-top: 12px;">
        <el-descriptions-item :label="$t('settings.demoStatus')">{{ demoStatus.demo_mode ? $t('settings.demoEnabled') : $t('settings.demoDisabled') }}</el-descriptions-item>
        <el-descriptions-item :label="$t('settings.demoUsers')">{{ demoStatus.user_count || 0 }}</el-descriptions-item>
        <el-descriptions-item :label="$t('settings.demoProjects')">{{ demoStatus.project_count || 0 }}</el-descriptions-item>
        <el-descriptions-item :label="$t('settings.demoOrders')">{{ demoStatus.order_count || 0 }}</el-descriptions-item>
        <el-descriptions-item :label="$t('settings.demoDisputes')">{{ demoStatus.dispute_count || 0 }}</el-descriptions-item>
      </el-descriptions>
      <el-alert type="info" :closable="false" :title="$t('settings.demoTip')" style="margin-top: 12px;" />
    </el-card>

    <el-card style="margin-bottom: 20px">
      <template #header>{{ $t('settings.version') }}</template>
      <el-table :data="versions" border stripe>
        <el-table-column prop="version" :label="$t('settings.versionNo')" width="100" />
        <el-table-column prop="build" :label="$t('settings.buildNo')" width="90" />
        <el-table-column prop="platform" :label="$t('settings.platform')" width="110" />
        <el-table-column prop="release_notes" :label="$t('settings.releaseNotes')" />
        <el-table-column prop="mandatory" :label="$t('settings.mandatory')" width="80">
          <template #default="{ row }">{{ row.mandatory ? $t('common.yes') : $t('common.no') }}</template>
        </el-table-column>
      </el-table>
      <el-button type="primary" style="margin-top: 15px" @click="openVersionDialog">{{ $t('settings.publishVersion') }}</el-button>
    </el-card>

    <el-card>
      <template #header>
        <div style="display: flex; align-items: center; justify-content: space-between;">
          <span>{{ $t('settings.monitor') }}</span>
          <el-button size="small" @click="loadMonitorStats">{{ $t('common.refresh') }}</el-button>
        </div>
      </template>
      <el-table :data="monitorStatsList" border stripe>
        <el-table-column prop="path" :label="$t('settings.monitorPath')" />
        <el-table-column prop="count" :label="$t('settings.monitorCount')" width="100" />
        <el-table-column prop="error_count" :label="$t('settings.monitorErrors')" width="100" />
        <el-table-column prop="avg_ms" :label="$t('settings.monitorAvg')" width="120">
          <template #default="{ row }">{{ row.avg_ms?.toFixed(1) || '0' }}</template>
        </el-table-column>
        <el-table-column prop="p95_ms" :label="$t('settings.monitorP95')" width="120">
          <template #default="{ row }">{{ row.p95_ms?.toFixed(1) || '0' }}</template>
        </el-table-column>
        <el-table-column prop="error_rate" :label="$t('settings.monitorErrorRate')" width="110">
          <template #default="{ row }">{{ row.error_rate?.toFixed(1) || '0' }}%</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="configDialog" :title="$t('settings.addConfig')" width="500px">
      <el-form :model="configForm" label-width="100px">
        <el-form-item :label="$t('settings.configKey')"><el-input v-model="configForm.config_key" /></el-form-item>
        <el-form-item :label="$t('settings.configValue')"><el-input v-model="configForm.config_value" /></el-form-item>
        <el-form-item :label="$t('settings.valueType')">
          <el-select v-model="configForm.value_type" style="width: 200px">
            <el-option :label="$t('settings.typeString')" value="string" />
            <el-option :label="$t('settings.typeInt')" value="int" />
            <el-option :label="$t('settings.typeBool')" value="bool" />
            <el-option :label="$t('settings.typeJson')" value="json" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('settings.description')"><el-input v-model="configForm.description" /></el-form-item>
        <el-form-item :label="$t('settings.isPublic')"><el-switch v-model="configForm.is_public" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveConfig">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="versionDialog" :title="$t('settings.publishVersion')" width="500px">
      <el-form :model="versionForm" label-width="100px">
        <el-form-item :label="$t('settings.versionNo')"><el-input v-model="versionForm.version" :placeholder="$t('settings.placeholderVersion')" /></el-form-item>
        <el-form-item :label="$t('settings.buildNo')"><el-input-number v-model="versionForm.build" /></el-form-item>
        <el-form-item :label="$t('settings.platform')">
          <el-select v-model="versionForm.platform" style="width: 200px">
            <el-option :label="$t('settings.platformAll')" value="all" />
            <el-option :label="$t('settings.platformH5')" value="h5" />
            <el-option :label="$t('settings.platformMp')" value="mp-weixin" />
            <el-option :label="$t('settings.platformApp')" value="app" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('settings.updateUrl')"><el-input v-model="versionForm.update_url" /></el-form-item>
        <el-form-item :label="$t('settings.releaseNotes')"><el-input type="textarea" v-model="versionForm.release_notes" /></el-form-item>
        <el-form-item :label="$t('settings.forceUpdate')"><el-switch v-model="versionForm.mandatory" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="versionDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="publishVersion">{{ $t('settings.publishVersion') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

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
  ElMessage.success($t('settings.saved'))
  configDialog.value = false
  await loadConfigs()
}

const deleteConfig = async (row: any) => {
  await ElMessageBox.confirm($t('settings.confirmDelete', { key: row.config_key }), $t('common.prompt'))
  await api.delete(`/api/v1/admin/config/delete/${row.config_key}`)
  ElMessage.success($t('settings.deleted'))
  await loadConfigs()
}

const seedDemo = async () => {
  await api.post(`/api/v1/admin/demo/seed?mode=${demoMode.value}`)
  ElMessage.success($t('settings.demoSeedSuccess'))
  await loadDemoStatus()
}

const cleanDemo = async () => {
  await ElMessageBox.confirm($t('settings.demoCleanConfirm'), $t('common.prompt'))
  await api.post('/api/v1/admin/demo/clean')
  ElMessage.success($t('settings.demoCleanSuccess'))
  await loadDemoStatus()
}

const openVersionDialog = () => {
  versionDialog.value = true
}

const publishVersion = async () => {
  await api.post('/api/v1/admin/version/publish', versionForm.value)
  ElMessage.success($t('settings.versionPublished'))
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
