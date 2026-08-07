<template>
  <view class="container">
    <view class="form">
      <view class="form-item">
        <input class="input" v-model="form.phone" placeholder="请输入手机号" type="number" maxlength="11" />
      </view>
      <view class="form-item code-row">
        <input class="input code-input" v-model="form.code" placeholder="验证码" type="number" maxlength="6" />
        <button class="code-btn" @tap="sendCode" :disabled="countdown > 0">
          {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
        </button>
      </view>
      <view class="form-item">
        <view class="role-select">
          <view :class="['role-item', form.userType === 1 ? 'active' : '']" @tap="form.userType = 1">甲方</view>
          <view :class="['role-item', form.userType === 2 ? 'active' : '']" @tap="form.userType = 2">服务方</view>
        </view>
      </view>
      <button class="submit-btn" @tap="login">登录</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()

const form = ref({
  phone: '',
  code: '',
  userType: 1,
})

const countdown = ref(0)

const sendCode = () => {
  if (!form.value.phone) {
    uni.showToast({ title: '请输入手机号', icon: 'none' })
    return
  }
  countdown.value = 60
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) clearInterval(timer)
  }, 1000)
  // TODO: Call send SMS API
}

const login = async () => {
  if (!form.value.phone || !form.value.code) {
    uni.showToast({ title: '请填写完整信息', icon: 'none' })
    return
  }
  await userStore.login(form.value.phone, form.value.code, form.value.userType)
  uni.switchTab({ url: '/pages/index/index' })
}
</script>

<style scoped>
.container {
  padding: 60rpx 40rpx;
}

.form-item {
  margin-bottom: 30rpx;
}

.input {
  background: #f5f5f5;
  padding: 24rpx 30rpx;
  border-radius: 10rpx;
  font-size: 30rpx;
}

.code-row {
  display: flex;
  gap: 20rpx;
}

.code-input {
  flex: 1;
}

.code-btn {
  background: #1890ff;
  color: #fff;
  font-size: 26rpx;
  padding: 0 30rpx;
  border-radius: 10rpx;
  white-space: nowrap;
}

.role-select {
  display: flex;
  gap: 20rpx;
}

.role-item {
  flex: 1;
  text-align: center;
  padding: 20rpx;
  background: #f5f5f5;
  border-radius: 10rpx;
  font-size: 28rpx;
}

.role-item.active {
  background: #1890ff;
  color: #fff;
}

.submit-btn {
  background: #1890ff;
  color: #fff;
  margin-top: 40rpx;
  border-radius: 10rpx;
}
</style>
