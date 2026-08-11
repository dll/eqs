<template>
  <div>
    <el-alert
      type="info"
      :closable="false"
      :title="$t('audit.gateHint')"
      style="margin-bottom: 12px"
    />
    <el-card>
      <template #header>
        {{ $t('audit.title') }}
      </template>
      <el-table
        :data="auditList"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          :label="$t('audit.id')"
          width="60"
        />
        <el-table-column
          prop="supplier_id"
          :label="$t('audit.supplierId')"
          width="80"
        />
        <el-table-column
          prop="qualification_type"
          :label="$t('audit.qualType')"
          width="120"
        />
        <el-table-column
          prop="certificate_no"
          :label="$t('audit.certNo')"
          width="130"
        />
        <el-table-column
          prop="level"
          :label="$t('audit.level')"
          width="70"
        />
        <el-table-column
          prop="issuing_authority"
          :label="$t('audit.authority')"
          width="140"
        />
        <el-table-column
          prop="valid_to"
          :label="$t('audit.validTo')"
          width="110"
        >
          <template #default="{ row }">
            {{ row.valid_to ? String(row.valid_to).slice(0, 10) : '-' }}
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('audit.status')"
          width="90"
        >
          <template #default="{ row }">
            <el-tag :type="statusType(row.verification_status)">
              {{ statusText(row.verification_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('audit.actions')"
          width="230"
        >
          <template #default="{ row }">
            <template v-if="row.verification_status === 'pending'">
              <el-input
                v-model="comments[row.id]"
                :placeholder="$t('audit.commentPh')"
                size="small"
                style="width: 130px; margin-right: 6px"
              />
              <el-button
                type="success"
                size="small"
                @click="review(row, true)"
              >
                {{ $t('common.approve') }}
              </el-button>
              <el-button
                type="danger"
                size="small"
                @click="review(row, false)"
              >
                {{ $t('common.reject') }}
              </el-button>
            </template>
            <span
              v-else-if="row.review_comment"
              class="comment-text"
            >
              {{ row.review_comment }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '@/utils/request'
import { useI18n } from '@/utils/i18n'

const { $t } = useI18n()

const auditList = ref<any[]>([])
const comments = ref<Record<number, string>>({})

const load = async () => {
  try {
    const res = await api.get<{ qualifications: any[] }>('/api/v1/admin/qualifications')
    auditList.value = res.qualifications || []
  } catch {
    // interceptor 已提示
  }
}

onMounted(load)

const review = async (row: any, verified: boolean) => {
  try {
    const res = await api.post(`/api/v1/qualification/${row.id}/review`, {
      verified,
      comment: comments.value[row.id] || '',
    })
    const ai = (res as any)?.ai_suggestion
    if (ai?.suggestion) {
      ElMessage.info(`${verified ? $t('common.passed') : $t('common.rejected')}。${$t('audit.aiSuggestion')}: ${ai.suggestion}`)
    } else {
      ElMessage.success(verified ? $t('common.passed') : $t('common.rejected'))
    }
    delete comments.value[row.id]
    load()
  } catch {
    // interceptor 已提示
  }
}

const statusText = (status: string) => {
  const map: Record<string, string> = {
    pending: $t('audit.status.pending'),
    approved: $t('audit.status.approved'),
    rejected: $t('audit.status.rejected'),
    expired: $t('audit.status.expired'),
  }
  return map[status] || status
}

const statusType = (status: string) => {
  const map: Record<string, string> = { pending: 'warning', approved: 'success', rejected: 'danger' }
  return map[status] || 'info'
}
</script>

<style scoped>
.comment-text {
  font-size: 12px;
  color: #909399;
}
</style>