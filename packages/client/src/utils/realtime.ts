// 站内实时推送工具
// H5 使用 EventSource 订阅后端 SSE（/api/v1/notify/stream，token 走查询参数）；
// 小程序/App 无 EventSource，降级为每 30s 轮询未读数。
import { request } from '@/utils/request'

export interface NotifyEvent {
  type: string
  title?: string
  content?: string
  ntype?: string
  unread?: number
}

/**
 * 连接通知实时流，返回断开函数。
 * onEvent 收到通知事件（含初始 connected 事件与轮询 unread 事件）。
 */
export const connectNotify = (onEvent: (data: NotifyEvent) => void): (() => void) => {
  let es: any = null
  let timer: any = null
  let stopped = false

  // #ifdef H5
  if (typeof window !== 'undefined' && (window as any).EventSource) {
    try {
      const token = uni.getStorageSync('token')
      if (token) {
        es = new (window as any).EventSource(`/api/v1/notify/stream?token=${encodeURIComponent(token)}`)
        es.addEventListener('notification', (e: any) => {
          if (stopped) return
          try {
            onEvent(JSON.parse(e.data))
          } catch {
            // 忽略无法解析的事件
          }
        })
      }
    } catch {
      es = null
    }
  }
  // #endif

  // 非 H5 / EventSource 不可用时轮询兜底
  if (!es) {
    const poll = async () => {
      if (stopped) return
      try {
        const res = await request.get('/api/v1/notification/unread-count', { silent401: true })
        onEvent({ type: 'unread', unread: res.unread || 0 })
      } catch {
        // 未登录/网络异常时静默
      }
    }
    poll()
    timer = setInterval(poll, 30000)
  }

  return () => {
    stopped = true
    if (es) {
      try {
        es.close()
      } catch {
        // ignore
      }
      es = null
    }
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }
}
