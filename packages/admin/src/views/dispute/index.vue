<template>
  <div>
    <el-card>
      <template #header>{{ $t('dispute.title') }}</template>
      <el-table :data="disputes" style="width: 100%">
        <el-table-column prop="id" :label="$t('dispute.id')" width="80" />
        <el-table-column prop="order_id" :label="$t('dispute.orderId')" width="80" />
        <el-table-column prop="reason" :label="$t('dispute.reason')" show-overflow-tooltip />
        <el-table-column prop="claim" :label="$t('dispute.claim')" show-overflow-tooltip />
        <el-table-column prop="initiator_id" :label="$t('dispute.initiatorId')" width="90" />
        <el-table-column prop="status" :label="$t('dispute.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resolution_type" :label="$t('dispute.resolution')" width="100">
          <template #default="{ row }">{{ row.resolution_type || '-' }}</template>
        </el-table-column>
        <el-table-column :label="$t('dispute.actions')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'evidence' || row.status === 'review' || row.status === 'mediation'"
              size="small"
              type="primary"
              plain
              @click="openExpertDialog(row)"
            >
              {{ $t('dispute.assignExpert') }}
            </el-button>
            <el-button
              v-if="row.status !== 'closed'"
              size="small"
              type="success"
              plain
              @click="openCloseDialog(row)"
            >
              {{ $t('dispute.close') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="expertDialog.visible" :title="$t('dispute.assignExpert')" width="420px">
      <el-form label-width="100px">
        <el-form-item :label="$t('dispute.disputeId')">
          <span>{{ expertDialog.disputeId }}</span>
        </el-form-item>
        <el-form-item :label="$t('dispute.expertUserId')">
          <el-input v-model="expertDialog.expertId" type="number" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="expertDialog.visible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="expertDialog.loading" @click="submitExpert">
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="closeDialog.visible" :title="$t('dispute.close')" width="420px">
      <el-form label-width="100px">
        <el-form-item :label="$t('dispute.resolution')">
          <el-select v-model="closeDialog.resolutionType" placeholder="settlement / agreement / award">
            <el-option label="settlement" value="settlement" />
            <el-option label="agreement" value="agreement" />
            <el-option label="award" value="award" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('dispute.settleAmount')">
          <el-input v-model="closeDialog.settleAmount" type="number" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeDialog.visible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="success" :loading="closeDialog.loading" @click="submitClose">
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

const disputes = ref<any[]>([])

const expertDialog = reactive({
  visible: false,
  disputeId: 0,
  expertId: '',
  loading: false,
})

const closeDialog = reactive({
  visible: false,
  disputeId: 0,
  resolutionType: 'settlement',
  settleAmount: '',
  loading: false,
})

const load = async () => {
  try {
    const res = await api.get<{ disputes: any[] }>('/api/v1/admin/disputes')
    disputes.value = res.disputes || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

const openExpertDialog = (row: any) => {
  expertDialog.disputeId = row.id
  expertDialog.expertId = ''
  expertDialog.visible = true
}

const submitExpert = async () => {
  const expertUserId = Number(expertDialog.expertId)
  if (!expertUserId) {
    ElMessage.warning($t('dispute.expertRequired'))
    return
  }
  expertDialog.loading = true
  try {
    await api.post(`/api/v1/dispute/${expertDialog.disputeId}/expert`, { expert_user_id: expertUserId })
    ElMessage.success($t('dispute.expertAssigned'))
    expertDialog.visible = false
    load()
  } catch {
    // interceptor 已提示
  } finally {
    expertDialog.loading = false
  }
}

const openCloseDialog = (row: any) => {
  closeDialog.disputeId = row.id
  closeDialog.resolutionType = 'settlement'
  closeDialog.settleAmount = ''
  closeDialog.visible = true
}

const submitClose = async () => {
  closeDialog.loading = true
  try {
    await api.post(`/api/v1/dispute/${closeDialog.disputeId}/close`, {
      resolution_type: closeDialog.resolutionType,
      settle_amount: Number(closeDialog.settleAmount) || 0,
    })
    ElMessage.success($t('dispute.closed'))
    closeDialog.visible = false
    load()
  } catch {
    // interceptor 已提示
  } finally {
    closeDialog.loading = false
  }
}

const statusText = (status: string) => {
  const map: Record<string, string> = {
    evidence: $t('dispute.status.evidence'),
    review: $t('dispute.status.review'),
    mediation: $t('dispute.status.mediation'),
    reconsideration: $t('dispute.status.reconsideration'),
    closed: $t('dispute.status.closed'),
  }
  return map[status] || status
}

const statusType = (status: string) => {
  if (status === 'closed') return 'success'
  if (status === 'review' || status === 'mediation') return 'warning'
  return 'info'
}
</script>
