<template>
  <view class="container">
    <view class="form">
      <view class="form-item">
        <text class="label">项目类型</text>
        <picker :range="projectTypes" @change="onTypeChange">
          <view class="picker">{{ form.projectType || '请选择项目类型' }}</view>
        </picker>
      </view>

      <view class="form-item">
        <text class="label">项目名称</text>
        <input class="input" v-model="form.title" placeholder="请输入项目名称" />
      </view>

      <view class="form-item">
        <text class="label">预算范围（万元）</text>
        <view class="budget-row">
          <input class="input budget-input" v-model="form.budgetMin" placeholder="最低" type="digit" />
          <text class="budget-sep">-</text>
          <input class="input budget-input" v-model="form.budgetMax" placeholder="最高" type="digit" />
        </view>
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

const projectTypes = ['造价咨询', '工程监理', '地质勘察', '工程设计']

const form = ref({
  projectType: '',
  title: '',
  budgetMin: '',
  budgetMax: '',
  description: '',
})

const onTypeChange = (e: any) => {
  form.value.projectType = projectTypes[e.detail.value]
}

const submit = () => {
  if (!form.value.projectType || !form.value.title) {
    uni.showToast({ title: '请填写必要信息', icon: 'none' })
    return
  }
  // TODO: Call API to create project
  uni.showToast({ title: '发布成功', icon: 'success' })
  setTimeout(() => uni.navigateBack(), 1500)
}
</script>

<style scoped>
.container {
  padding: 30rpx;
}

.form-item {
  margin-bottom: 30rpx;
}

.label {
  font-size: 28rpx;
  color: #333;
  display: block;
  margin-bottom: 15rpx;
}

.input {
  background: #f5f5f5;
  padding: 20rpx;
  border-radius: 10rpx;
  font-size: 28rpx;
}

.picker {
  background: #f5f5f5;
  padding: 20rpx;
  border-radius: 10rpx;
  font-size: 28rpx;
  color: #666;
}

.budget-row {
  display: flex;
  align-items: center;
  gap: 15rpx;
}

.budget-input {
  flex: 1;
}

.budget-sep {
  font-size: 28rpx;
  color: #999;
}

.textarea {
  background: #f5f5f5;
  padding: 20rpx;
  border-radius: 10rpx;
  font-size: 28rpx;
  height: 200rpx;
}

.submit-btn {
  background: #1890ff;
  color: #fff;
  margin-top: 40rpx;
  border-radius: 10rpx;
}
</style>
