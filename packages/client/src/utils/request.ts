const BASE_URL = ''

const getToken = () => uni.getStorageSync('token')

export const request = {
  get: (url: string) => {
    return new Promise((resolve, reject) => {
      uni.request({
        url: BASE_URL + url,
        method: 'GET',
        header: { Authorization: `Bearer ${getToken()}` },
        success: (res) => resolve(res.data as any),
        fail: reject,
      })
    })
  },

  post: (url: string, data?: any) => {
    return new Promise((resolve, reject) => {
      uni.request({
        url: BASE_URL + url,
        method: 'POST',
        data,
        header: { Authorization: `Bearer ${getToken()}` },
        success: (res) => resolve(res.data as any),
        fail: reject,
      })
    })
  },

  put: (url: string, data?: any) => {
    return new Promise((resolve, reject) => {
      uni.request({
        url: BASE_URL + url,
        method: 'PUT',
        data,
        header: { Authorization: `Bearer ${getToken()}` },
        success: (res) => resolve(res.data as any),
        fail: reject,
      })
    })
  },
}
