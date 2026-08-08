import axios, { AxiosRequestConfig } from 'axios'

const getToken = () => localStorage.getItem('token')

const request = axios.create({
  baseURL: '',
  timeout: 10000,
})

request.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    const msg = error.response?.data?.message || '请求失败'
    return Promise.reject(new Error(msg))
  }
)

// 泛型封装：直接返回后端数据体
export const api = {
  get: <T = any>(url: string, params?: any) => request.get<T, T>(url, { params }),
  post: <T = any>(url: string, data?: any) => request.post<T, T>(url, data),
  put: <T = any>(url: string, data?: any) => request.put<T, T>(url, data),
  delete: <T = any>(url: string) => request.delete<T, T>(url),
}

export { request }
export type { AxiosRequestConfig }