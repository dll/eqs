<template>
  <view class="container">
    <view class="card">
      <view class="card-title">
        {{ $t('qual.title') }}
      </view>
      <view class="form-item">
        <text class="label">
          {{ $t('qual.type') }}
        </text>
        <input
          v-model="form.qualification_type"
          class="input"
          :placeholder="$t('qual.typePh')"
        >
      </view>
      <view class="form-item">
        <text class="label">
          {{ $t('qual.certNo') }}
        </text>
        <input
          v-model="form.certificate_no"
          class="input"
          :placeholder="$t('qual.certNoPh')"
        >
      </view>
      <view class="form-item">
        <text class="label">
          {{ $t('qual.level') }}
        </text>
        <input
          v-model="form.level"
          class="input"
          :placeholder="$t('qual.levelPh')"
        >
      </view>
      <view class="form-item">
        <text class="label">
          {{ $t('qual.scope') }}
        </text>
        <input
          v-model="form.scope"
          class="input"
          :placeholder="$t('qual.scopePh')"
        >
      </view>
      <view class="form-item">
        <text class="label">
          {{ $t('qual.authority') }}
        </text>
        <input
          v-model="form.issuing_authority"
          class="input"
          :placeholder="$t('qual.authorityPh')"
        >
      </view>
      <view class="form-item">
        <text class="label">
          {{ $t('qual.issueDate') }}
        </text>
        <picker
          mode="date"
          :value="form.issue_date || ''"
          @change="onIssueDate"
        >
          <view class="input picker-value">
            {{ form.issue_date || $t('qual.issueDatePh') }}
          </view>
        </picker>
      </view>
      <view class="form-item">
        <text class="label">
          {{ $t('qual.validTo') }}
        </text>
        <picker
          mode="date"
          :value="form.valid_to || ''"
          @change="onValidTo"
        >
          <view class="input picker-value">
            {{ form.valid_to || $t('qual.validToPh') }}
          </view>
        </picker>
      </view>
      <view class="form-item">
        <text class="label">
          {{ $t('qual.upload') }}
        </text>
        <button
          class="upload-btn"
          :disabled="uploading"
          @tap="chooseAndUpload"
        >
          {{ uploading ? $t('qual.uploading') : (form.evidence_file_id ? $t('qual.uploaded') : $t('qual.upload')) }}
        </button>
      </view>
      <button
        class="submit-btn"
        @tap="submit"
      >
        {{ $t('qual.submit') }}
      </button>
      <text
        v-if="submitMsg"
        class="submit-msg"
      >
        {{ submitMsg }}
      </text>
    </view>

    <view class="card">
      <view class="card-title">
        {{ $t('qual.myList') }}
      </view>
      <view
        v-if="qualList.length"
        class="q-item"
      >
        <view
          v-for="q in qualList"
          :key="q.id"
          class="q-row"
        >
          <view class="q-main">
            <text class="q-type">
              {{ q.qualification_type }}
            </text>
            <text class="q-detail">
              {{ q.level || '' }} · {{ q.issuing_authority || '' }}
            </text>
            <text class="q-detail">
              {{ $t('qual.certNo') }}：{{ q.certificate_no || '-' }}
            </text>
            <text class="q-detail">
              {{ $t('qual.validTo') }}：{{ q.valid_to ? String(q.valid_to).slice(0, 10) : '-' }}
            </text>
            <text
              v-if="q.review_comment"
              class="q-detail q-comment"
            >
              {{ $t('qual.comment') }}{{ q.review_comment }}
            </text>
          </view>
          <text
            class="q-status"
            :class="'s-' + q.verification_status"
          >
            {{ statusText(q.verification_status) }}
          </text>
          <view class="q-ops">
            <text
              v-if="q.evidence_file_id"
              class="q-op"
              @tap.stop="previewFile(q)"
            >
              预览
            </text>
            <text
              class="q-op"
              @tap.stop="viewQual(q)"
            >
              详情
            </text>
            <text
              v-if="q.verification_status !== 'approved'"
              class="q-op"
              @tap.stop="editQual(q)"
            >
              编辑
            </text>
            <text
              v-if="q.verification_status !== 'approved'"
              class="q-op danger"
              @tap.stop="removeQual(q)"
            >
              删除
            </text>
          </view>
        </view>
      </view>
      <text
        v-else
        class="empty"
      >
        {{ $t('qual.empty') }}
      </text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useUserStore } from '@/store/user'
import { useI18n, usePageTitle } from '@/utils/i18n'

const { $t } = useI18n()
usePageTitle('page.qual', { onShow })
const userStore = useUserStore()

// initialForm 资质提交表单初始值（reset 复用，避免字段漂移）
const initialForm = () => ({
  qualification_type: '',
  certificate_no: '',
  level: '',
  scope: '',
  issuing_authority: '',
  issue_date: '',
  valid_from: '',
  valid_to: '',
  evidence_file_id: 0,
})
const form = ref(initialForm())
const qualList = ref<any[]>([])
const submitMsg = ref('')
const uploading = ref(false)

const supplierId = () => userStore.user?.id

onShow(() => loadList())

const loadList = async () => {
  const sid = supplierId()
  if (!sid) return
  try {
    const res = await request.get(`/api/v1/supplier/${sid}/qualifications`, { silent401: true }).catch(() => ({ qualifications: [] }))
    qualList.value = res.qualifications || []
  } catch {
    qualList.value = []
  }
}

const onIssueDate = (e: any) => {
  form.value.issue_date = e.detail.value
  // 有效期起默认同签发日，可后续调整
  if (!form.value.valid_from) form.value.valid_from = e.detail.value
}

const onValidTo = (e: any) => {
  form.value.valid_to = e.detail.value
}

// 选择并上传扫描件（uni.uploadFile → /qualification/upload）
const chooseAndUpload = async () => {
  if (uploading.value) return
  const sid = supplierId()
  if (!sid) {
    submitMsg.value = $t('qual.notSupplier')
    return
  }
  uploading.value = true
  try {
    const token = uni.getStorageSync('token')
    const res: any = await new Promise((resolve, reject) => {
      uni.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        success: (chooseRes: any) => {
          const filePath = chooseRes.tempFilePaths[0]
          uni.uploadFile({
            url: `/api/v1/qualification/upload`,
            filePath,
            name: 'file',
            header: { Authorization: `Bearer ${token}` },
            success: (up: any) => {
              try {
                resolve(JSON.parse(up.data))
              } catch {
                reject(new Error('bad response'))
              }
            },
            fail: reject,
          })
        },
        fail: reject,
      })
    })
    if (res && res.file_id) {
      form.value.evidence_file_id = res.file_id
      submitMsg.value = $t('qual.uploaded')
    } else {
      submitMsg.value = $t('qual.uploadFail')
    }
  } catch {
    submitMsg.value = $t('qual.uploadFail')
  } finally {
    uploading.value = false
  }
}

const submit = async () => {
  const sid = supplierId()
  if (!sid) {
    submitMsg.value = $t('qual.notSupplier')
    return
  }
  if (!form.value.qualification_type || !form.value.certificate_no || !form.value.valid_to) {
    submitMsg.value = $t('qual.required')
    return
  }
  const payload: any = { ...form.value }
  if (!payload.evidence_file_id) delete payload.evidence_file_id
  try {
    if (editingID.value) {
      await request.put(`/api/v1/qualification/${editingID.value}`, payload)
      submitMsg.value = '资质已更新并重新进入审核'
    } else {
      await request.post(`/api/v1/supplier/${sid}/qualifications`, payload)
      submitMsg.value = $t('qual.submitted')
    }
    editingID.value = 0
    form.value = initialForm()
    loadList()
  } catch {
    submitMsg.value = $t('qual.fail')
  }
}

const statusText = (s: string) => {
  const map: Record<string, string> = {
    pending: $t('qual.pending'),
    approved: $t('qual.approved'),
    rejected: $t('qual.rejected'),
    expired: $t('qual.expired'),
  }
  return map[s] || s
}

// V9：编辑资质（回填表单后 PUT）
const editingID = ref(0)

// V9：查看资质详情（弹窗展示完整信息与审核意见）
const detailQual = ref<any>(null)

// V10：文件在线预览（下载后按类型打开：图片 previewImage / PDF openDocument）
const previewFile = (q: any) => {
  if (!q.evidence_file_id) return
  uni.showLoading({ title: '加载中...' })
  const token = uni.getStorageSync('token')
  uni.downloadFile({
    url: `/api/v1/file/${q.evidence_file_id}/preview`,
    header: { Authorization: `Bearer ${token}` },
    success: (res: any) => {
      uni.hideLoading()
      if (res.statusCode !== 200) {
        uni.showToast({ title: '预览失败', icon: 'none' })
        return
      }
      const isImage = /\.(jpg|jpeg|png)$/i.test(res.tempFilePath)
      if (isImage) {
        uni.previewImage({ urls: [res.tempFilePath] })
      } else {
        uni.openDocument({ filePath: res.tempFilePath, showMenu: true, fail: () => uni.showToast({ title: '请下载后查看', icon: 'none' }) })
      }
    },
    fail: () => {
      uni.hideLoading()
      uni.showToast({ title: '预览失败', icon: 'none' })
    },
  })
}
const viewQual = (q: any) => {
  detailQual.value = q
  uni.showModal({
    title: '资质详情',
    content: [
      `类型：${q.qualification_type || '-'}`,
      `证书号：${q.certificate_no || '-'}`,
      `等级：${q.level || '-'}`,
      `发证机关：${q.issuing_authority || '-'}`,
      `有效期至：${q.valid_to ? String(q.valid_to).slice(0, 10) : '-'}`,
      `状态：${statusText(q.verification_status)}`,
      q.review_comment ? `审核意见：${q.review_comment}` : '',
    ].filter(Boolean).join('\n'),
    showCancel: false,
    confirmText: '关闭',
  })
}
const editQual = (q: any) => {
  editingID.value = q.id
  form.value = {
    qualification_type: q.qualification_type || '',
    certificate_no: q.certificate_no || '',
    level: q.level || '',
    scope: q.scope || '',
    issuing_authority: q.issuing_authority || '',
    issue_date: q.issue_date ? String(q.issue_date).slice(0, 10) : '',
    valid_from: q.valid_from ? String(q.valid_from).slice(0, 10) : '',
    valid_to: q.valid_to ? String(q.valid_to).slice(0, 10) : '',
    evidence_file_id: q.evidence_file_id || 0,
  }
  uni.pageScrollTo({ scrollTop: 0, duration: 200 })
}

// V9：删除资质（待审核/已驳回可删）
const removeQual = (q: any) => {
  uni.showModal({
    title: '删除资质',
    content: '确认删除该资质？删除后不可恢复。',
    success: async (r: any) => {
      if (!r.confirm) return
      try {
        await request.delete(`/api/v1/qualification/${q.id}`)
        uni.showToast({ title: '已删除', icon: 'none' })
        loadList()
      } catch {
        // request 已提示
      }
    },
  })
}
</script>

<style scoped>
.container {
  padding: 20rpx;
  background: var(--bg-color, #f5f5f5);
  min-height: 100vh;
}
.card {
  background: var(--card-bg, #fff);
  border-radius: 10rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}
.card-title {
  font-size: 32rpx;
  font-weight: bold;
  color: var(--text-color, #333);
  margin-bottom: 20rpx;
}
.form-item {
  margin-bottom: 20rpx;
}
.label {
  display: block;
  font-size: 26rpx;
  color: var(--muted-color, #666);
  margin-bottom: 8rpx;
}
.input {
  background: var(--input-bg, #f5f5f5);
  border-radius: 8rpx;
  padding: 16rpx 20rpx;
  font-size: 28rpx;
  color: var(--text-color, #333);
}
.submit-btn {
  background: var(--primary-color, #409eff);
  color: #fff;
  border-radius: 10rpx;
  font-size: 30rpx;
  margin-top: 10rpx;
}
.upload-btn {
  background: var(--input-bg, #f5f5f5);
  color: var(--primary-color, #409eff);
  border: 1rpx solid var(--primary-color, #409eff);
  border-radius: 8rpx;
  font-size: 28rpx;
  padding: 14rpx 0;
}
.picker-value {
  line-height: 44rpx;
  color: var(--text-color, #333);
}
.submit-msg {
  display: block;
  text-align: center;
  margin-top: 16rpx;
  font-size: 26rpx;
  color: var(--primary-color);
}
.q-item .q-row {
  display: flex;
  justify-content: space-between;
  padding: 14rpx 0;
  border-bottom: 1rpx solid var(--border-color, #eee);
  font-size: 28rpx;
}
.q-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}
.q-type {
  color: var(--text-color, #333);
  font-weight: bold;
}
.q-detail {
  font-size: 24rpx;
  color: var(--muted-color, #888);
}
.q-comment {
  color: #e6a23c;
}
.q-status {
  font-size: 24rpx;
  padding: 4rpx 14rpx;
  border-radius: 20rpx;
  color: #fff;
}
.s-pending { background: #e6a23c; }
.s-approved { background: #67c23a; }
.s-rejected { background: #f56c6c; }
.s-expired { background: #909399; }
.empty {
  font-size: 26rpx;
  color: var(--muted-color, #999);
}
</style>
