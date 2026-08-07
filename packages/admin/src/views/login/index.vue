<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2>EQS 管理后台</h2>
      <el-form :model="form" @submit.prevent="login">
        <el-form-item>
          <el-input v-model="form.phone" placeholder="手机号" prefix-icon="Phone" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.code" placeholder="验证码" prefix-icon="Message" />
        </el-form-item>
        <el-button type="primary" @click="login" style="width: 100%">登录</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/store/user'

const router = useRouter()
const userStore = useUserStore()

const form = ref({
  phone: '',
  code: '',
})

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

h2 {
  text-align: center;
  margin-bottom: 30px;
}
</style>
