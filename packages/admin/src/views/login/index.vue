<template>
  <div class="login-container eqs-grid-bg">
    <!-- 装饰光晕 -->
    <div class="glow glow-blue" />
    <div class="glow glow-ai" />
    <div class="glow glow-cyan" />

    <div class="login-card eqs-glass">
      <div class="brand">
        <div class="brand-badge">
          ⚡
        </div>
        <div class="brand-title">
          {{ $t('app.title') }}
        </div>
        <div class="brand-sub">
          Agile · AI · Engineering Service
        </div>
      </div>

      <el-form
        :model="form"
        class="login-form"
        @submit.prevent="login"
      >
        <el-form-item>
          <el-input
            v-model="form.phone"
            :placeholder="$t('login.phone')"
            prefix-icon="Phone"
            size="large"
          />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="form.code"
            :placeholder="$t('login.code')"
            prefix-icon="Message"
            size="large"
          />
        </el-form-item>
        <el-button
          type="primary"
          size="large"
          class="login-btn"
          @click="login"
        >
          {{ $t('login.submit') }}
        </el-button>
      </el-form>

      <div class="ai-tip">
        <span class="ai-dot" />
        AI 智能审核 · 敏捷交付闭环 · 工程全流程管控
      </div>

      <div class="lang-switch">
        <el-select
          :model-value="lang"
          size="small"
          style="width: 100px"
          @change="onLangChange"
        >
          <el-option
            label="中文"
            value="zh-CN"
          />
          <el-option
            label="EN"
            value="en-US"
          />
        </el-select>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/store/user'
import { useI18n } from '@/utils/i18n'

const router = useRouter()
const userStore = useUserStore()
const { $t, lang, setAdminLang } = useI18n()

const form = ref({
  phone: '',
  code: '',
})

const onLangChange = (v: string) => {
  setAdminLang(v)
  document.title = $t('login.title')
}

const login = async () => {
  await userStore.login(form.value.phone, form.value.code)
  router.push('/')
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    linear-gradient(135deg, rgba(15, 30, 61, .92) 0%, rgba(30, 58, 138, .88) 45%, rgba(14, 116, 144, .90) 100%),
    var(--eqs-bg);
  position: relative;
  overflow: hidden;
}

/* 浮动光晕 */
.glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: .5;
  pointer-events: none;
}
.glow-blue {
  width: 420px;
  height: 420px;
  background: #2563eb;
  top: -120px;
  left: -80px;
}
.glow-ai {
  width: 380px;
  height: 380px;
  background: #8b5cf6;
  bottom: -140px;
  right: -60px;
}
.glow-cyan {
  width: 300px;
  height: 300px;
  background: #06b6d4;
  top: 40%;
  left: 55%;
  opacity: .25;
}

.login-card {
  width: 400px;
  padding: 40px 36px;
  position: relative;
  z-index: 1;
}

.brand {
  text-align: center;
  margin-bottom: 30px;
}
.brand-badge {
  width: 56px;
  height: 56px;
  margin: 0 auto 14px;
  border-radius: 16px;
  background: var(--eqs-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #fff;
  box-shadow: 0 8px 24px rgba(37, 99, 235, .45);
}
.brand-title {
  font-size: 24px;
  font-weight: 800;
  color: var(--eqs-text);
  letter-spacing: 1px;
}
.brand-sub {
  margin-top: 6px;
  font-size: 12px;
  color: var(--eqs-text-secondary);
  letter-spacing: .6px;
}

.login-form {
  margin-top: 8px;
}
.login-form :deep(.el-input__wrapper) {
  border-radius: 10px;
  padding: 4px 14px;
}
.login-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  border-radius: 10px;
  margin-top: 6px;
}

.ai-tip {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 22px;
  font-size: 12px;
  color: var(--eqs-text-secondary);
}
.ai-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--eqs-ai);
  box-shadow: 0 0 8px var(--eqs-ai);
}

.lang-switch {
  margin-top: 18px;
  text-align: center;
}
</style>
