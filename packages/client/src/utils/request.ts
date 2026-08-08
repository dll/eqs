// BASE_URL：开发联调走本地后端（走 vite proxy），生产由环境注入
const BASE_URL = ''

let token = uni.getStorageSync('token')

export const setToken = (t: string) => {
  token = t
  uni.setStorageSync('token', t)
}

export const clearToken = () => {
  token = ''
  uni.removeStorageSync('token')
}

const handler = (options: uni.RequestOptions) =>
  new Promise<any>((resolve, reject) => {
    uni.request({
      ...options,
      url: BASE_URL + options.url,
      header: { 'Content-Type': 'application/json', Authorization: token ? `Bearer ${token}` : '' },
      success: (res: any) => {
        const data = res.data || {}
        if (res.statusCode >= 400) {
          uni.showToast({ title: data.message || '请求失败', icon: 'none' })
          if (res.statusCode === 401) {
            clearToken()
            uni.reLaunch({ url: '/pages/login/index' })
          }
          reject(data)
          return
        }
        resolve(data)
      },
      fail: (err) => {
        uni.showToast({ title: '网络异常', icon: 'none' })
        reject(err)
      },
    })
  })

export const request = {
  get: (url: string) => handler({ url, method: 'GET' }),
  post: (url: string, data?: any) => handler({ url, method: 'POST', data }),
  put: (url: string, data?: any) => handler({ url, method: 'PUT', data }),
  delete: (url: string) => handler({ url, method: 'DELETE' }),
}