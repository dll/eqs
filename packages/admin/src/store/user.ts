import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/utils/request'

interface AdminUser {
  id: number
  phone: string
  user_type: number
  company_name: string
  credit_score: number
  status: number
}

export const useUserStore = defineStore('user', () => {
  const user = ref<AdminUser | null>(null)
  const token = ref('')

  const login = async (phone: string, code: string) => {
    const res = await api.post<{ token: string; user: AdminUser; isNew: boolean }>('/api/v1/auth/login', {
      phone,
      code,
      user_type: 3,
    })
    token.value = res.token
    user.value = res.user
    localStorage.setItem('token', res.token)
  }

  const logout = () => {
    user.value = null
    token.value = ''
    localStorage.removeItem('token')
  }

  const loadUser = async () => {
    const t = localStorage.getItem('token')
    if (!t) return
    token.value = t
    try {
      const res = await api.get<{ user: AdminUser }>('/api/v1/user/info')
      user.value = res.user
    } catch {
      logout()
    }
  }

  return { user, token, login, logout, loadUser }
})