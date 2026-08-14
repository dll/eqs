<template>
  <view class="page">
    <!-- 品牌头图 -->
    <view class="hero">
      <view class="hero-glow hero-glow-ai" />
      <view class="hero-glow hero-glow-cyan" />
      <view class="hero-badge">
        ⚡
      </view>
      <text class="hero-title">
        {{ $t('app.title') }}
      </text>
      <text class="hero-sub">
        Agile · AI · Engineering Service
      </text>
    </view>

    <view class="container">
      <view class="lang-switch">
        <text
          class="lang-btn"
          :class="settingsStore.lang === 'zh-CN' ? 'active' : ''"
          @tap="settingsStore.setLang('zh-CN')"
        >
          中文
        </text>
        <text
          class="lang-btn"
          :class="settingsStore.lang === 'en-US' ? 'active' : ''"
          @tap="settingsStore.setLang('en-US')"
        >
          EN
        </text>
      </view>

      <view class="form">
        <view class="form-item">
          <input
            v-model="form.phone"
            class="input"
            :placeholder="$t('login.phone')"
            type="number"
            maxlength="11"
          >
        </view>
        <view class="form-item code-row">
          <input
            v-model="form.code"
            class="input code-input"
            :placeholder="$t('login.code')"
            type="number"
            maxlength="6"
          >
          <button
            class="code-btn"
            :disabled="countdown > 0"
            @tap="sendCode"
          >
            {{ countdown > 0 ? `${countdown}s` : $t('login.getCode') }}
          </button>
        </view>
        <view class="form-item">
          <view class="role-select">
            <view
              :class="['role-item', form.userType === 1 ? 'active' : '']"
              @tap="form.userType = 1"
            >
              {{ $t('login.client') }}
            </view>
            <view
              :class="['role-item', form.userType === 2 ? 'active' : '']"
              @tap="form.userType = 2"
            >
              {{ $t('login.supplier') }}
            </view>
            <view
              :class="['role-item', form.userType === 3 ? 'active' : '']"
              @tap="form.userType = 3"
            >
              {{ $t('login.admin') }}
            </view>
          </view>
        </view>
        <button
          class="submit-btn"
          @tap="login"
        >
          {{ $t('common.login') }}
        </button>
        <text class="ai-tip">
          AI 智能审核 · 敏捷交付闭环
        </text>
      </view>
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
.page {
  min-height: 100vh;
  background: var(--bg-color);
}

/* 品牌头图 */
.hero {
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #2563eb 0%, #06b6d4 100%);
  padding: 80rpx 40rpx 60rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.hero-glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(40rpx);
  opacity: .45;
}
.hero-glow-ai {
  width: 260rpx;
  height: 260rpx;
  background: #8b5cf6;
  top: -80rpx;
  right: -40rpx;
}
.hero-glow-cyan {
  width: 220rpx;
  height: 220rpx;
  background: #06b6d4;
  bottom: -70rpx;
  left: -40rpx;
}
.hero-badge {
  width: 96rpx;
  height: 96rpx;
  border-radius: 28rpx;
  background: rgba(255, 255, 255, .2);
  backdrop-filter: blur(8rpx);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48rpx;
  color: #fff;
  position: relative;
  z-index: 1;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, .15);
}
.hero-title {
  color: #fff;
  font-size: 44rpx;
  font-weight: 800;
  margin-top: 24rpx;
  position: relative;
  z-index: 1;
  letter-spacing: 2rpx;
}
.hero-sub {
  color: rgba(255, 255, 255, .85);
  font-size: 22rpx;
  margin-top: 10rpx;
  letter-spacing: 1rpx;
  position: relative;
  z-index: 1;
}

.container {
  padding: 40rpx;
  margin-top: -20rpx;
  position: relative;
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
  padding: 8rpx 24rpx;
  border-radius: 20rpx;
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
  background: var(--card-bg);
  padding: 24rpx 30rpx;
  border-radius: 14rpx;
  font-size: 30rpx;
  color: var(--text-color);
  border: 1rpx solid var(--border-color);
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
  border-radius: 14rpx;
  white-space: nowrap;
}

.role-select {
  display: flex;
  gap: 20rpx;
}

.role-item {
  flex: 1;
  text-align: center;
  padding: 22rpx;
  background: var(--card-bg);
  border-radius: 14rpx;
  font-size: 28rpx;
  color: var(--muted-color);
  border: 1rpx solid var(--border-color);
}

.role-item.active {
  background: linear-gradient(135deg, #2563eb, #06b6d4);
  color: #fff;
  border-color: transparent;
  box-shadow: 0 6rpx 18rpx rgba(37, 99, 235, .3);
}

.submit-btn {
  background: linear-gradient(135deg, #2563eb, #06b6d4);
  color: #fff;
  margin-top: 40rpx;
  border-radius: 14rpx;
  font-size: 32rpx;
  font-weight: 600;
  box-shadow: 0 8rpx 24rpx rgba(37, 99, 235, .3);
}

.ai-tip {
  display: block;
  text-align: center;
  margin-top: 24rpx;
  font-size: 22rpx;
  color: var(--muted-color);
}
</style>
