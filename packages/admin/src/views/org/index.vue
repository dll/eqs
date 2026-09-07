<template>
  <div>
    <el-alert
      type="success"
      :closable="false"
      :title="$t('org.title')"
      :description="$t('org.hint')"
      style="margin-bottom: 12px"
    />
    <el-card>
      <template #header>
        <div class="bar">
          <span>{{ $t('org.title') }}</span>
          <el-button
            type="primary"
            @click="openCreate"
          >
            <el-icon style="margin-right:4px"><Plus /></el-icon>
            {{ $t('org.create') }}
          </el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="orgs"
        style="width:100%"
      >
        <el-table-column :label="$t('org.id')" prop="id" width="70" />
        <el-table-column :label="$t('org.name')" prop="name" min-width="160" />
        <el-table-column :label="$t('org.type')" width="110">
          <template #default="{ row }">
            {{ $t('org.type.' + (row.type || 'other')) }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('org.members')" width="140">
          <template #default="{ row }">
            <el-button link type="primary" @click="openMembers(row)">{{ $t('org.members') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 新建组织 -->
      <el-dialog v-model="createOpen" :title="$t('org.create')" width="440px">
        <el-form label-width="100px">
          <el-form-item :label="$t('org.name')">
            <el-input v-model="form.name" />
          </el-form-item>
          <el-form-item :label="$t('org.type')">
            <el-select v-model="form.type" style="width:100%">
              <el-option v-for="ty in ['client','supplier','internal','other']" :key="ty" :label="$t('org.type.'+ty)" :value="ty" />
            </el-select>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="createOpen=false">{{ $t('common.cancel') }}</el-button>
          <el-button type="primary" @click="create">{{ $t('common.save') }}</el-button>
        </template>
      </el-dialog>

      <!-- 成员 -->
      <el-dialog v-model="memberOpen" :title="$t('org.members')" width="520px">
        <el-table :data="members" size="small" style="width:100%">
          <el-table-column prop="user_id" :label="$t('org.userId')" width="90" />
          <el-table-column prop="phone" :label="$t('org.userPhone')" />
          <el-table-column prop="role" :label="$t('org.role')" width="110" />
        </el-table>
        <template #footer>
          <el-button link type="primary" @click="memberOpen=false">{{ $t('common.cancel') }}</el-button>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()
const orgs = ref<any[]>([])
const members = ref<any[]>([])
const loading = ref(false)
const createOpen = ref(false)
const memberOpen = ref(false)
const current = ref<any>(null)
const form = reactive({ name: '', type: 'client' })

const load = async () => {
  loading.value = true
  try {
    const res = await api.get<{ organizations: any[] }>('/api/v1/org/mine')
    orgs.value = res.organizations || []
  } catch { /* i */ } finally { loading.value = false }
}
const openCreate = () => { form.name = ''; form.type = 'client'; createOpen.value = true }
const create = async () => {
  if (!form.name) { ElMessage.warning($t('org.name')); return }
  try { await api.post('/api/v1/org/create', form); ElMessage.success($t('common.save')); createOpen.value = false; load() } catch { /* i */ }
}
const openMembers = async (row: any) => {
  current.value = row
  try { const res = await api.get<{ members: any[] }>(`/api/v1/org/${row.id}/members`); members.value = res.members || []; memberOpen.value = true } catch { /* i */ }
}
onMounted(load)
</script>

<style scoped>
.bar { display: flex; align-items: center; justify-content: space-between; }
</style>
