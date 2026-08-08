// uni 全局类型声明（uni-app 运行环境提供）
// 独立提供，避免与 @dcloudio/types legacy 声明在 moduleResolution=bundler 下冲突

declare namespace uni {
  interface RequestOptions {
    url: string
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
    data?: any
    header?: any
    success?: (res: any) => void
    fail?: (err: any) => void
    complete?: (res: any) => void
  }
}

declare const uni: {
  getStorageSync(key: string): string
  setStorageSync(key: string, value: unknown): void
  removeStorageSync(key: string): void
  request(options: uni.RequestOptions): any
  showToast(options: { title: string; icon?: string; duration?: number }): void
  navigateBack(options?: any): void
  reLaunch(options: { url: string }): void
  navigateTo(options: { url: string }): void
}