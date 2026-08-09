// BASE_URL：开发联调走本地后端（走 vite proxy），生产由环境注入
import { t } from '@/utils/i18n'

const BASE_URL = ''

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  silent401?: boolean // 401 时不跳转登录（用于启动时的可选请求）
}

let token = uni.getStorageSync('token')

export const setToken = (t: string) => {
  token = t
  uni.setStorageSync('token', t)
}

export const clearToken = () => {
  token = ''
  uni.removeStorageSync('token')
}

const handler = (options: RequestOptions) =>
  new Promise<any>((resolve, reject) => {
    uni.request({
      url: BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data,
      header: { 'Content-Type': 'application/json', Authorization: token ? `Bearer ${token}` : '' },
      success: (res: any) => {
        const data = res.data || {}
        if (res.statusCode >= 400) {
          if (res.statusCode === 401) {
            clearToken()
            if (!options.silent401) {
              uni.reLaunch({ url: '/pages/login/index' })
            }
          } else {
            uni.showToast({ title: data.message || t('error.requestFailed'), icon: 'none' })
          }
          reject(data)
          return
        }
        resolve(data)
      },
      fail: (err: any) => {
        uni.showToast({ title: t('error.network'), icon: 'none' })
        reject(err)
      },
    })
  })

export const request = {
  get: (url: string, opts?: { silent401?: boolean }) =>
    handler({ url, method: 'GET', ...opts }),
  post: (url: string, data?: any, opts?: { silent401?: boolean }) =>
    handler({ url, method: 'POST', data, ...opts }),
  put: (url: string, data?: any) => handler({ url, method: 'PUT', data }),
  delete: (url: string) => handler({ url, method: 'DELETE' }),
}
