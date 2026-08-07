import { defineStore } from 'pinia'
import { ref } from 'vue'
import { request } from '@/utils/request'

interface AdminUser {
  id: number
  phone: string
  user_type: number
  company_name: string
}

export const useUserStore = defineStore('user', () => {
  const user = ref<AdminUser | null>(null)
  const token = ref('')

  const login = async (phone: string, code: string) => {
    const res = await request.post('/api/v1/auth/login', { phone, code, user_type: 3 })
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
  }

  const logout = () => {
    user.value = null
    token.value = ''
    localStorage.removeItem('token')
  }

  return { user, token, login, logout }
})
