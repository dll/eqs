import { describe, it, expect, vi, beforeEach } from 'vitest'
import { request, api } from './request'

describe('admin request 封装', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('带 token 请求时注入 Authorization', async () => {
    localStorage.setItem('token', 'admin-1')
    const req = request as any
    const cfg = await req.interceptors.request.handlers[0].fulfilled({
      headers: {}, url: '/x', method: 'get',
    })
    expect(cfg.headers.Authorization).toBe('Bearer admin-1')
  })

  it('无 token 不注入', async () => {
    const req = request as any
    const cfg = await req.interceptors.request.handlers[0].fulfilled({ headers: {}, url: '/y' })
    expect(cfg.headers.Authorization).toBeUndefined()
  })

  it('响应直接返回 data', () => {
    const req = request as any
    const res = req.interceptors.response.handlers[0].fulfilled({ data: { token: 't', user: { id: 1 } } })
    expect(res.token).toBe('t')
  })

  it('401 清除 token 并跳转', async () => {
    localStorage.setItem('token', 'bad')
    const req = request as any
    const err = new Error('x')
    ;(err as any).response = { status: 401, data: { message: '未登录' } }
    await expect(req.interceptors.response.handlers[0].rejected(err)).rejects.toThrow('未登录')
    expect(localStorage.getItem('token')).toBeNull()
  })

  it('非 401 错误仅抛出 message', async () => {
    const req = request as any
    const err = new Error('x')
    ;(err as any).response = { status: 500, data: { message: '服务器错误' } }
    await expect(req.interceptors.response.handlers[0].rejected(err)).rejects.toThrow('服务器错误')
  })

  it('api 泛型封装返回数据体', async () => {
    const getSpy = vi.spyOn(request, 'get').mockResolvedValue({ ok: true })
    const data = await api.get('/api/v1/orders', { status: 2 })
    expect(getSpy).toHaveBeenCalledWith('/api/v1/orders', { params: { status: 2 } })
    expect(data).toEqual({ ok: true })
  })
})