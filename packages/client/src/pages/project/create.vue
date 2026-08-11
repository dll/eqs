<template>
  <view class="container">
    <view class="form">
      <view class="form-item">
        <text class="label">
          {{ $t('project.type') }}
        </text>
        <picker
          :range="projectTypes"
          @change="onTypeChange"
        >
          <view class="picker">
            {{ form.projectType || $t('project.selectType') }}
          </view>
        </picker>
      </view>

      <view class="form-item">
        <text class="label">
          {{ $t('project.name') }}
        </text>
        <input
          v-model="form.title"
          class="input"
          :placeholder="$t('project.namePlaceholder')"
        >
      </view>

      <!-- 智能服务清单（按所选类型自动生成） -->
      <view
        v-if="checklist.length"
        class="form-item"
      >
        <text class="label">
          {{ $t('project.smartChecklist') }}
        </text>
        <view class="checklist-box">
          <view
            v-for="(item, i) in checklist"
            :key="i"
            class="checklist-item"
          >
            <text class="cl-dot">
              ✓
            </text>
            <text class="cl-text">
              {{ item }}
            </text>
          </view>
        </view>
      </view>

      <view class="form-item">
        <text class="label">
          {{ $t('project.budgetRange') }}
        </text>
        <view class="budget-row">
          <input
            v-model="form.budgetMin"
            class="input budget-input"
            :placeholder="$t('project.budgetMin')"
            type="digit"
          >
          <text class="budget-sep">
            -
          </text>
          <input
            v-model="form.budgetMax"
            class="input budget-input"
            :placeholder="$t('project.budgetMax')"
            type="digit"
          >
        </view>
      </view>

      <view class="form-item">
        <text class="label">
          {{ $t('project.location') }}
        </text>
        <input
          v-model="form.address"
          class="input"
          :placeholder="$t('project.addressPlaceholder')"
        >
      </view>

      <view class="form-item">
        <text class="label">
          {{ $t('project.desc') }}
        </text>
        <textarea
          v-model="form.description"
          class="textarea"
          :placeholder="$t('project.descPlaceholder')"
        />
      </view>

      <view class="form-item">
        <text class="label">
          {{ $t('project.attachment') }}
        </text>
        <button
          class="upload-btn"
          :disabled="uploading"
          @tap="chooseAndUpload"
        >
          {{ uploading ? $t('project.uploading') : $t('project.uploadAttachment') }}
        </button>
        <view
          v-if="attachments.length"
          class="att-list"
        >
          <view
            v-for="(a, i) in attachments"
            :key="a.file_id"
            class="att-item"
          >
            <text class="att-name">
              {{ a.original_name }}
            </text>
            <text
              class="att-del"
              @tap="attachments.splice(i, 1)"
            >
              ✕
            </text>
          </view>
        </view>
      </view>

      <button
        class="submit-btn"
        @tap="submit"
      >
        {{ $t('project.create') }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { useI18n, usePageTitle } from '@/utils/i18n'
import { PROJECT_TYPES, toTypeCode } from '@/lib/service'

const { $t } = useI18n()
usePageTitle('page.projectCreate', { onLoad })

const projectTypes = PROJECT_TYPES
const editId = ref(0)

const form = ref({
  projectType: '',
  title: '',
  budgetMin: '',
  budgetMax: '',
  address: '',
  description: '',
})

// V10：附件上传（批文/CAD/PDF，≤50MB，先登记后随项目提交）
const attachments = ref<any[]>([])
const uploading = ref(false)

const chooseAndUpload = async () => {
  if (uploading.value) return
  uploading.value = true
  try {
    const token = uni.getStorageSync('token')
    const res: any = await new Promise((resolve, reject) => {
      uni.chooseImage({
        count: 1,
        sizeType: ['original'],
        success: (chooseRes: any) => {
          const filePath = chooseRes.tempFilePaths[0]
          uni.uploadFile({
            url: `/api/v1/project/upload`,
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
      attachments.value.push({ file_id: res.file_id, original_name: res.original_name })
    } else {
      uni.showToast({ title: $t('project.uploadFail'), icon: 'none' })
    }
  } catch {
    uni.showToast({ title: $t('project.uploadFail'), icon: 'none' })
  } finally {
    uploading.value = false
  }
}

const onTypeChange = (e: any) => {
  form.value.projectType = projectTypes[e.detail.value]
  loadChecklist()
}

// V10：智能服务清单（按服务类型自动生成，AC-01）
const checklist = ref<string[]>([])
const loadChecklist = async () => {
  const code = toTypeCode(form.value.projectType)
  if (!code) {
    checklist.value = []
    return
  }
  try {
    const res = await request.get(`/api/v1/project/checklist?service_type=${code}`)
    checklist.value = (res && res.checklist) || []
  } catch {
    checklist.value = []
  }
}

// 编辑模式：加载项目详情回填表单
onLoad((options) => {
  if (options?.id) {
    editId.value = Number(options.id)
    loadProject(editId.value)
  }
})

const loadProject = async (id: number) => {
  try {
    const res = await request.get(`/api/v1/project/${id}`)
    const p = res.project
    form.value = {
      projectType: p.service_type || p.project_type,
      title: p.title,
      budgetMin: String(p.budget_min ?? ''),
      budgetMax: String(p.budget_max ?? ''),
      address: p.address || '',
      description: p.description || '',
    }
  } catch {
    // request 已提示
  }
}

const submit = async () => {
  if (!form.value.projectType || !form.value.title) {
    uni.showToast({ title: $t('project.required'), icon: 'none' })
    return
  }
  const payload = {
    project_type: toTypeCode(form.value.projectType),
    service_type: toTypeCode(form.value.projectType),
    title: form.value.title,
    address: form.value.address,
    budget_min: Number(form.value.budgetMin) || 0,
    budget_max: Number(form.value.budgetMax) || 0,
    description: form.value.description,
    publish_scope: 'public',
  }
  try {
    if (editId.value) {
      await request.put(`/api/v1/project/${editId.value}`, payload)
      uni.showToast({ title: $t('project.edited'), icon: 'success' })
    } else {
      await request.post('/api/v1/project/create', payload)
      uni.showToast({ title: $t('project.publishSuccess'), icon: 'success' })
    }
    setTimeout(() => uni.navigateBack(), 1500)
  } catch {
    // request 已提示
  }
}
</script>

<style scoped>
.container {
  padding: 20rpx;
}

.form-item {
  margin-bottom: 30rpx;
}

.label {
  display: block;
  font-size: 28rpx;
  color: var(--text-color);
  margin-bottom: 12rpx;
}

.picker {
  background: var(--input-bg);
  padding: 24rpx 30rpx;
  border-radius: 10rpx;
  font-size: 30rpx;
  color: var(--text-color);
}

.input {
  background: var(--input-bg);
  padding: 24rpx 30rpx;
  border-radius: 10rpx;
  font-size: 30rpx;
}

.budget-row {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.budget-input {
  flex: 1;
}

.budget-sep {
  color: var(--muted-color);
}

.textarea {
  background: var(--input-bg);
  padding: 24rpx 30rpx;
  border-radius: 10rpx;
  font-size: 30rpx;
  width: auto;
  min-height: 160rpx;
}

.submit-btn {
  background: var(--primary-color);
  color: #fff;
  margin-top: 40rpx;
  border-radius: 10rpx;
}
</style>