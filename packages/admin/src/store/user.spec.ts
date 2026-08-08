import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const mocks = vi.hoisted(() => ({
  api: { get: vi.fn(), post: vi.fn(), put: vi.fn() },
}))

vi.mock('@/utils/request', () => ({ api: mocks.api }))

import { useUserStore } from './user'

describe('admin user store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mocks.api.get.mockReset()
    mocks.api.post.mockReset()
    localStorage.clear()
  })

  it('login 保存 token 与用户（user_type=3）', async () => {
    const store = useUserStore()
    mocks.api.post.mockResolvedValue({ token: 't1', user: { id: 3, user_type: 3 } })
    await store.login('admin', '123456')
    expect(store.token).toBe('t1')
    expect(store.user?.id).toBe(3)
    expect(localStorage.getItem('token')).toBe('t1')
    expect(mocks.api.post).toHaveBeenCalledWith('/api/v1/auth/login', {
      phone: 'admin', code: '123456', user_type: 3,
    })
  })

  it('loadUser 无 token 直接返回', async () => {
    const store = useUserStore()
    await store.loadUser()
    expect(mocks.api.get).not.toHaveBeenCalled()
  })

  it('loadUser 有 token 拉取用户', async () => {
    localStorage.setItem('token', 't2')
    const store = useUserStore()
    mocks.api.get.mockResolvedValue({ user: { id: 9 } })
    await store.loadUser()
    expect(store.user?.id).toBe(9)
  })

  it('loadUser 失败登出', async () => {
    localStorage.setItem('token', 'bad')
    const store = useUserStore()
    mocks.api.get.mockRejectedValue(new Error('401'))
    await store.loadUser()
    expect(store.token).toBe('')
    expect(store.user).toBeNull()
    expect(localStorage.getItem('token')).toBeNull()
  })

  it('logout 清空本地与内存', () => {
    localStorage.setItem('token', 'x')
    const store = useUserStore()
    store.token = 'x'
    store.logout()
    expect(store.token).toBe('')
    expect(store.user).toBeNull()
    expect(localStorage.getItem('token')).toBeNull()
  })
})