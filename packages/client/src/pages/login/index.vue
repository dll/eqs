<template>
  <view class="container">
    <view class="lang-switch">
      <text class="lang-btn" :class="settingsStore.lang === 'zh-CN' ? 'active' : ''" @tap="settingsStore.setLang('zh-CN')">中文</text>
      <text class="lang-btn" :class="settingsStore.lang === 'en-US' ? 'active' : ''" @tap="settingsStore.setLang('en-US')">EN</text>
    </view>
    <view class="form">
      <view class="form-item">
        <input class="input" v-model="form.phone" :placeholder="$t('login.phone')" type="number" maxlength="11" />
      </view>
      <view class="form-item code-row">
        <input class="input code-input" v-model="form.code" :placeholder="$t('login.code')" type="number" maxlength="6" />
        <button class="code-btn" @tap="sendCode" :disabled="countdown > 0">
          {{ countdown > 0 ? `${countdown}s` : $t('login.getCode') }}
        </button>
      </view>
      <view class="form-item">
        <view class="role-select">
          <view :class="['role-item', form.userType === 1 ? 'active' : '']" @tap="form.userType = 1">{{ $t('login.client') }}</view>
          <view :class="['role-item', form.userType === 2 ? 'active' : '']" @tap="form.userType = 2">{{ $t('login.supplier') }}</view>
        </view>
      </view>
      <button class="submit-btn" @tap="login">{{ $t('common.login') }}</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useUserStore } from '@/store/user'
import { useSettingsStore } from '@/store/settings'
import { useI18n, usePageTitle } from '@/utils/i18n'
import { request } from '@/utils/request'

const userStore = useUserStore()
const settingsStore = useSettingsStore()
const { $t } = useI18n()
usePageTitle('page.login', { onLoad })

const form = ref({
  phone: '',
  code: '',
  userType: 1,
})

const countdown = ref(0)

const sendCode = async () => {
  if (!form.value.phone) {
    uni.showToast({ title: $t('login.phoneRequired'), icon: 'none' })
    return
  }
  try {
    await request.post('/api/v1/sms/send', { phone: form.value.phone })
    uni.showToast({ title: $t('login.codeSent'), icon: 'success' })
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch {
    // request 已提示
  }
}

const login = async () => {
  if (!form.value.phone || !form.value.code) {
    uni.showToast({ title: $t('login.fillAll'), icon: 'none' })
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

.lang-switch {
  display: flex;
  justify-content: flex-end;
  gap: 16rpx;
  margin-bottom: 40rpx;
}

.lang-btn {
  font-size: 26rpx;
  color: var(--muted-color);
  padding: 8rpx 20rpx;
  border-radius: 8rpx;
  background: var(--input-bg);
}

.lang-btn.active {
  background: var(--primary-color);
  color: #fff;
}

.form-item {
  margin-bottom: 30rpx;
}

.input {
  background: var(--input-bg);
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
  background: var(--primary-color);
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
  background: var(--input-bg);
  border-radius: 10rpx;
  font-size: 28rpx;
}

.role-item.active {
  background: var(--primary-color);
  color: #fff;
}

.submit-btn {
  background: var(--primary-color);
  color: #fff;
  margin-top: 40rpx;
  border-radius: 10rpx;
}
</style>
