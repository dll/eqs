<template>
  <view class="container">
    <view class="form">
      <view class="form-item">
        <text class="label">服务类型</text>
        <picker :range="projectTypes" @change="onTypeChange">
          <view class="picker">{{ form.projectType || '请选择服务类型' }}</view>
        </picker>
      </view>

      <view class="form-item">
        <text class="label">项目名称</text>
        <input class="input" v-model="form.title" placeholder="请输入项目名称" />
      </view>

      <view class="form-item">
        <text class="label">预算范围（元）</text>
        <view class="budget-row">
          <input class="input budget-input" v-model="form.budgetMin" placeholder="最低" type="digit" />
          <text class="budget-sep">-</text>
          <input class="input budget-input" v-model="form.budgetMax" placeholder="最高" type="digit" />
        </view>
      </view>

      <view class="form-item">
        <text class="label">项目地点</text>
        <input class="input" v-model="form.address" placeholder="请输入工程地址" />
      </view>

      <view class="form-item">
        <text class="label">项目描述</text>
        <textarea class="textarea" v-model="form.description" placeholder="请输入项目描述" />
      </view>

      <button class="submit-btn" @tap="submit">发布项目</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { request } from '@/utils/request'

const projectTypes = ['造价咨询', '工程监理', '地质勘察', '工程设计']
const typeCode: Record<string, string> = { 造价咨询: 'cost', 工程监理: 'supervision', 地质勘察: 'geotech', 工程设计: 'design' }

const form = ref({
  projectType: '',
  title: '',
  budgetMin: '',
  budgetMax: '',
  address: '',
  description: '',
})

const onTypeChange = (e: any) => {
  form.value.projectType = projectTypes[e.detail.value]
}

const submit = async () => {
  if (!form.value.projectType || !form.value.title) {
    uni.showToast({ title: '请填写必要信息', icon: 'none' })
    return
  }
  try {
    await request.post('/api/v1/project/create', {
      project_type: typeCode[form.value.projectType] || 'cost',
      service_type: typeCode[form.value.projectType] || 'cost',
      title: form.value.title,
      address: form.value.address,
      budget_min: Number(form.value.budgetMin) || 0,
      budget_max: Number(form.value.budgetMax) || 0,
      description: form.value.description,
      publish_scope: 'public',
    })
    uni.showToast({ title: '发布成功', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 1500)
  } catch {
    // request 已提示
  }
}
</script>