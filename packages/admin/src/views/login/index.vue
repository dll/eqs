<template>
  <div class="login-container">
    <el-card class="login-card">
      <div class="login-head">
        <h2>{{ $t('app.title') }}</h2>
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
      <el-form
        :model="form"
        @submit.prevent="login"
      >
        <el-form-item>
          <el-input
            v-model="form.phone"
            :placeholder="$t('login.phone')"
            prefix-icon="Phone"
          />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="form.code"
            :placeholder="$t('login.code')"
            prefix-icon="Message"
          />
        </el-form-item>
        <el-button
          type="primary"
          style="width: 100%"
          @click="login"
        >
          {{ $t('login.submit') }}
        </el-button>
      </el-form>
    </el-card>
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
  // 语言切换后重新渲染标题
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
  background: #f0f2f5;
}

.login-card {
  width: 400px;
}

.login-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 30px;
}

h2 {
  margin: 0;
}
</style>