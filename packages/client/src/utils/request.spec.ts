import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setToken, clearToken, request } from './request'

// 捕获 uni.request 的调用配置，让测试可注入成功/失败
let lastOptions: any = null

beforeEach(() => {
  lastOptions = null
  const uniAny: any = (globalThis as any).uni
  uniAny.request = (opts: any) => {
    lastOptions = opts
  }
})

describe('request 封装', () => {
  it('发起成功请求并返回解析后的数据体', async () => {
    ;(globalThis as any).uni.request = (opts: any) => {
      opts.success({ statusCode: 200, data: { token: 'abc', user: { id: 1 } } })
    }
    const res = await request.post('/api/v1/auth/login', { phone: 'x', code: 'y' })
    expect(res.token).toBe('abc')
    expect(res.user.id).toBe(1)
  })

  it('设置 token 后写入 header', async () => {
    setToken('tok-123')
    ;(globalThis as any).uni.request = (opts: any) => {
      lastOptions = opts
      opts.success({ statusCode: 200, data: { ok: true } })
    }
    await request.get('/api/v1/user/info')
    expect(lastOptions.header.Authorization).toBe('Bearer tok-123')
  })

  it('400 错误被拒绝并携带 message', async () => {
    ;(globalThis as any).uni.request = (opts: any) => {
      opts.success({ statusCode: 400, data: { message: '参数错误' } })
    }
    await expect(request.post('/bad')).rejects.toMatchObject({ message: '参数错误' })
  })

  it('401 清除 token 并跳转登录页', async () => {
    setToken('expired')
    ;(globalThis as any).uni.request = (opts: any) => {
      opts.success({ statusCode: 401, data: { message: '未登录' } })
    }
    const reLaunch = vi.fn()
    ;(globalThis as any).uni.reLaunch = reLaunch
    await expect(request.get('/api/v1/user/info')).rejects.toBeTruthy()
    expect(clearToken)
    expect((globalThis as any).uni.getStorageSync('token')).toBe('')
    expect(reLaunch).toHaveBeenCalledWith({ url: '/pages/login/index' })
  })

  it('clearToken 清空本地 token', () => {
    setToken('x')
    clearToken()
    expect((globalThis as any).uni.getStorageSync('token')).toBe('')
  })
})