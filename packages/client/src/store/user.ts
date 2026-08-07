import { defineStore } from 'pinia'
import { ref } from 'vue'
import { request } from '@/utils/request'

interface User {
  id: number
  phone: string
  user_type: number
  company_name: string
  credit_score: number
}

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const token = ref('')

  const login = async (phone: string, code: string, userType: number) => {
    const res = await request.post('/api/v1/auth/login', { phone, code, user_type: userType })
    token.value = res.data.token
    user.value = res.data.user
    uni.setStorageSync('token', res.data.token)
  }

  const logout = () => {
    user.value = null
    token.value = ''
    uni.removeStorageSync('token')
  }

  const loadUser = async () => {
    const storedToken = uni.getStorageSync('token')
    if (storedToken) {
      token.value = storedToken
      try {
        const res = await request.get('/api/v1/user/info')
        user.value = res.data.user
      } catch {
        logout()
      }
    }
  }

  return { user, token, login, logout, loadUser }
})
