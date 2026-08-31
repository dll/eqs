import { defineStore } from 'pinia'
import { ref } from 'vue'
import { request, setToken, clearToken } from '@/utils/request'

interface User {
  id: number
  phone: string
  user_type: number
  company_name: string
  credit_score: number
  status: number
}

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const token = ref('')

  const login = async (phone: string, code: string, userType: number) => {
    const res = await request.post('/api/v1/auth/login', { phone, code, user_type: userType })
    token.value = res.token
    user.value = res.user
    setToken(res.token)
  }

  // V11：微信小程序登录。code 由调用方（条件编译中的 uni.login/uni.getUserProfile 流程）取得，
  // 此处仅负责交换 token 并落库，H5 不调用本方法。
  const wechatLogin = async (code: string, userType = 1) => {
    const res = await request.post('/api/v1/auth/wechat-login', { code, user_type: userType })
    token.value = res.token
    user.value = res.user
    setToken(res.token)
  }

  const logout = () => {
    user.value = null
    token.value = ''
    clearToken()
  }

  const loadUser = async () => {
    const storedToken = uni.getStorageSync('token')
    if (!storedToken) return
    token.value = storedToken
    try {
      const res = await request.get('/api/v1/user/info')
      user.value = res.user
    } catch {
      logout()
    }
  }

  return { user, token, login, wechatLogin, logout, loadUser }
})