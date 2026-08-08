import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { setToken, clearToken } from '@/utils/request'

const mocks = vi.hoisted(() => ({
  request: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

vi.mock('@/utils/request', () => ({
  request: mocks.request,
  setToken: (t: string) => { (globalThis as any).uni.setStorageSync('token', t) },
  clearToken: () => { (globalThis as any).uni.removeStorageSync('token') },
}))

import { useUserStore } from './user'

describe('user store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mocks.request.get.mockReset()
    mocks.request.post.mockReset()
    ;(globalThis as any).uni.removeStorageSync('token')
  })

  it('login 保存 token 与用户信息', async () => {
    const store = useUserStore()
    mocks.request.post.mockResolvedValue({ token: 't1', user: { id: 7, user_type: 1 } })
    await store.login('138', '123456', 1)
    expect(store.token).toBe('t1')
    expect(store.user?.id).toBe(7)
    expect((globalThis as any).uni.getStorageSync('token')).toBe('t1')
    expect(mocks.request.post).toHaveBeenCalledWith('/api/v1/auth/login', {
      phone: '138', code: '123456', user_type: 1,
    })
  })

  it('loadUser 无 token 直接返回', async () => {
    const store = useUserStore()
    await store.loadUser()
    expect(mocks.request.get).not.toHaveBeenCalled()
  })

  it('loadUser 有 token 时拉取用户', async () => {
    setToken('t2')
    const store = useUserStore()
    mocks.request.get.mockResolvedValue({ user: { id: 9 } })
    await store.loadUser()
    expect(store.user?.id).toBe(9)
  })

  it('loadUser 失败时登出清空', async () => {
    setToken('bad')
    const store = useUserStore()
    mocks.request.get.mockRejectedValue(new Error('401'))
    await store.loadUser()
    expect(store.token).toBe('')
    expect(store.user).toBeNull()
    expect((globalThis as any).uni.getStorageSync('token')).toBe('')
  })

  it('logout 清空', () => {
    const store = useUserStore()
    store.login = vi.fn()
    store.user = { id: 1, phone: 'x', user_type: 1, company_name: '', credit_score: 100, status: 1 }
    store.token = 'x'
    clearToken()
    store.logout()
    expect(store.token).toBe('')
    expect(store.user).toBeNull()
  })
})